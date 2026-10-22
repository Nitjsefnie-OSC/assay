package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// The container bracket (desktools-v2/71). A runner inside a --network none
// container cannot reach the forge, so it never runs verifier admission
// itself. The host brackets the run instead, and one tree digest binds the
// bytes the container ran to the admitted home three times:
//
//  1. host open: `verifyrun --open-bracket` runs `deskdispatch
//     --check-verifier`, digests the home, and writes the receipt file (the
//     check's JSON verbatim, `tree`, `opened`) outside the home;
//  2. container: `verifyrun --sandbox=container-netns --admission-receipt
//     <file> --container-image-digest sha256:<…>` reads that receipt (no
//     admission of its own), digests its own root and refuses unless it equals
//     the receipt's tree, then runs the rows and records Tree-Digest;
//  3. host close: `verifyrun --close-bracket` re-runs the check, compares it
//     with the witness's attestation line, recomputes the home's digest, and
//     only when it equals both the receipt's tree and the witness's
//     Tree-Digest appends an Admission-Bracket line.
//
// A container-sandboxed row is credited only inside a closed bracket.

const (
	containerImagePrefix   = "Container-Image: "
	treeDigestPrefix       = "Tree-Digest: "
	admissionBracketPrefix = "Admission-Bracket: "
	attestationPrefix      = "Verification-Attestation: "
)

var (
	sha256DigestRe = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	hex64Re        = regexp.MustCompile(`^[0-9a-f]{64}$`)
	hex40Re        = regexp.MustCompile(`^[0-9a-f]{40}$`)
	// attestationV1Re is the version-1 attestation line verifierAdmissionReceipt
	// renders. The version-2 form (source-repo=, tracking-source=) never matches.
	attestationV1Re = regexp.MustCompile(`^Verification-Attestation: (\S+)#([0-9]+) run=(\S+) source=(\S+) brief=(\S+) model=(\S+) tier=(\S+)$`)
)

// bracketReceipt is the open step's receipt file.
type bracketReceipt struct {
	Check  json.RawMessage `json:"check"`
	Tree   string          `json:"tree"`
	Opened string          `json:"opened"`
}

// containerBracketRefusal validates the bracket flags of a row-running
// invocation. It is called right after sandboxRefusal, so the CI-lane and
// --in-container refusals keep their own wording. container-netns needs both
// the host receipt and the image digest, whatever DESK_LOOP says and whether
// or not an attestation file is visible: an unattested container run is
// refused before any row runs.
func containerBracketRefusal(mode, receipt, digest string) string {
	if mode != sandboxContainerNetns {
		if receipt != "" || digest != "" {
			return "--admission-receipt and --container-image-digest apply only with --sandbox=" + sandboxContainerNetns + " (and --admission-receipt with --close-bracket)"
		}
		return ""
	}
	if receipt == "" {
		return "refusing --sandbox=" + sandboxContainerNetns + " without --admission-receipt <file>: a container run reads the host open step's receipt, read-only and outside the home, and never runs admission itself. An unattested container run is not a witness; no row ran"
	}
	if digest == "" {
		return "refusing --sandbox=" + sandboxContainerNetns + " without --container-image-digest sha256:<64 hex>: the host names the image the rows ran in, and a run whose image cannot be named is could-not-run. No row ran"
	}
	if !sha256DigestRe.MatchString(digest) {
		return fmt.Sprintf("refusing --container-image-digest %q: want sha256: followed by 64 lowercase hex digits", digest)
	}
	return ""
}

// readBracketReceipt reads and validates an open step's receipt file. The file
// must sit outside the home, carry a sha256 tree and a parseable opened time,
// and embed a complete version-1 check receipt. It returns the receipt, the
// attestation line verifierAdmissionReceipt renders for it, and the receipt's
// home-relative brief.
func readBracketReceipt(path, home, brief string) (bracketReceipt, string, string, error) {
	var r bracketReceipt
	if rel, ok := relWithin(home, path); ok {
		return r, "", "", fmt.Errorf("the receipt %s sits inside the home (%s); hand it over read-only from outside the home", path, rel)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return r, "", "", err
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, "", "", fmt.Errorf("unreadable receipt: %w", err)
	}
	if !sha256DigestRe.MatchString(r.Tree) {
		return r, "", "", fmt.Errorf("the receipt carries no sha256 tree digest from the open step")
	}
	if _, err := time.Parse(time.RFC3339, r.Opened); err != nil {
		return r, "", "", fmt.Errorf("the receipt carries no RFC 3339 opened time from the open step")
	}
	var peek struct {
		Binding struct {
			Version                  int
			Run, Repo, Source, Brief string
		}
	}
	if err := json.Unmarshal(r.Check, &peek); err != nil {
		return r, "", "", fmt.Errorf("unreadable check output in the receipt: %w", err)
	}
	b := peek.Binding
	if b.Version == 2 {
		return r, "", "", fmt.Errorf("version 2 (cross-repository) receipts are not admitted in container mode: they name host paths the container does not have. Run a cross-repository verification on the host")
	}
	if b.Version != 1 {
		return r, "", "", fmt.Errorf("unsupported receipt version %d", b.Version)
	}
	if b.Repo == "" || !hex64Re.MatchString(b.Run) || !hex40Re.MatchString(b.Source) || b.Brief == "" || !filepath.IsLocal(b.Brief) || filepath.ToSlash(filepath.Clean(b.Brief)) != b.Brief {
		return r, "", "", fmt.Errorf("incomplete receipt: it needs a repo, a 64-hex run, a 40-hex source and a home-relative brief")
	}
	line, _, admitted, err := verifierAdmissionReceipt(r.Check, brief)
	if err != nil {
		return r, "", "", err
	}
	if admitted != "" {
		return r, "", "", fmt.Errorf("a cross-repository receipt is not admitted in container mode")
	}
	if rel, ok := relWithin(home, brief); !ok || rel != b.Brief {
		return r, "", "", fmt.Errorf("the receipt admits brief %s, but this run's brief is %s", b.Brief, brief)
	}
	return r, line, b.Brief, nil
}

// containerAdmission is the container half: the host's receipt replaces
// admission, and the mounted root must digest to the receipt's tree before
// any row runs. It returns the attestation line and that digest.
func containerAdmission(receiptPath, root, brief string) (string, string, error) {
	r, line, _, err := readBracketReceipt(receiptPath, root, brief)
	if err != nil {
		return "", "", fmt.Errorf("container admission refused; no Verify rows executed: %w", err)
	}
	tree, err := homeTreeDigest(root)
	if err != nil {
		return "", "", fmt.Errorf("container admission refused; no Verify rows executed: %w", err)
	}
	if tree != r.Tree {
		return "", "", fmt.Errorf("container admission refused; no Verify rows executed: the mounted root %s digests to %s, not the %s the host recorded at the open step. The container is not running the admitted home", root, tree, r.Tree)
	}
	return line, tree, nil
}

// relWithin returns p's slash path relative to root when p lies inside root,
// after resolving symlinks on both.
func relWithin(root, p string) (string, bool) {
	resolve := func(s string) string {
		if abs, err := filepath.Abs(s); err == nil {
			s = abs
		}
		if real, err := filepath.EvalSymlinks(s); err == nil {
			return real
		}
		return filepath.Clean(s)
	}
	rel, err := filepath.Rel(resolve(root), resolve(p))
	if err != nil || !filepath.IsLocal(rel) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// homeTreeDigest is the one tree digest every bracket step uses: sha256 over
// the path-sorted list of `<path> NUL <mode> NUL <sha256 of content or link
// target> NUL` for every regular file and symlink under root, excluding the
// top-level .git entry. Paths are slash-separated and relative to root; mode
// is git's (100644, 100755 or 120000). Admission already refuses any untracked
// or ignored file in the home, so on an admitted home the set is exactly the
// tracked tree at the attested source. It reads the file system, not git,
// because a container need not see a linked worktree's git directory. Any
// other kind of entry (a socket, a device) refuses.
func homeTreeDigest(root string) (string, error) {
	type entry struct{ path, mode, sum string }
	var entries []entry
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		var content []byte
		mode := "100644"
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			content, mode = []byte(target), "120000"
		case info.Mode().IsRegular():
			if content, err = os.ReadFile(p); err != nil {
				return err
			}
			if info.Mode()&0o111 != 0 {
				mode = "100755"
			}
		default:
			return fmt.Errorf("%s is neither a regular file nor a symlink", rel)
		}
		sum := sha256.Sum256(content)
		entries = append(entries, entry{rel, mode, hex.EncodeToString(sum[:])})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("tree digest of %s: %w", root, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00", e.path, e.mode, e.sum)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// containerWitnessLines are the lines a container run adds under its
// attestation line.
func containerWitnessLines(image, tree string) string {
	return containerImagePrefix + image + "\n" + treeDigestPrefix + tree + "\n"
}

// writeContainerWitness writes a container run's witness to its own new file
// outside the home. Writing it into the brief would change the home, and the
// close step's tree digest would then refuse the bracket.
func writeContainerWitness(path, home, table string) error {
	if rel, ok := relWithin(home, path); ok {
		return fmt.Errorf("refusing --witness %s: it sits inside the home (%s), which would change the tree the bracket binds", path, rel)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(strings.TrimRight(table, "\n") + "\n"); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// runOpenBracket is `verifyrun --open-bracket`: the host's first half. It
// writes the receipt file only when the host check admits the home.
func runOpenBracket(home, brief, out string, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "statusgen verifyrun --open-bracket: refused, no receipt written: "+format+"\n", a...)
		return verifyrunExitCouldNot
	}
	if rel, ok := relWithin(home, out); ok {
		return refuse("--out %s sits inside the home (%s); the receipt must be outside it", out, rel)
	}
	check, err := hostCheckVerifier(home, brief)
	if err != nil {
		return refuse("%v", err)
	}
	if _, _, admitted, err := verifierAdmissionReceipt(check, brief); err != nil {
		return refuse("%v", err)
	} else if admitted != "" {
		return refuse("version 2 (cross-repository) receipts are not admitted in container mode")
	}
	tree, err := homeTreeDigest(home)
	if err != nil {
		return refuse("%v", err)
	}
	opened := nowFunc().UTC().Format(time.RFC3339)
	body := `{"check":` + string(check) + `,"tree":"` + tree + `","opened":"` + opened + `"}` + "\n"
	if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
		return refuse("%v", err)
	}
	fmt.Fprintf(stdout, "opened %s tree=%s receipt=%s\n", opened, tree, out)
	return verifyrunExitPass
}

// hostCheckVerifier runs the shared admission check on the host and returns
// its JSON output, trimmed and validated as JSON.
func hostCheckVerifier(home, brief string) ([]byte, error) {
	out, err := exec.Command("deskdispatch", "--check-verifier", "--root", home, "--brief", brief).Output()
	if err != nil {
		detail := ""
		if exit, ok := err.(*exec.ExitError); ok {
			detail = strings.TrimSpace(string(exit.Stderr))
		}
		return nil, fmt.Errorf("the host admission check failed: %v: %s", err, detail)
	}
	out = bytes.TrimSpace(out)
	if !json.Valid(out) {
		return nil, fmt.Errorf("the host admission check returned no JSON receipt")
	}
	return out, nil
}

// containerBlock is the container witness at the end of a witness file.
type containerBlock struct {
	lines       []string // the file's lines, without the final newline's empty tail
	span        []string // the hashed span: witness header line through Tree-Digest
	attestation string
	brief       string // the attestation's home-relative brief
	tree        string
}

// findContainerBlock reads a container witness file: it must END with the
// witness table, the version-1 attestation line and the Container-Image and
// Tree-Digest lines, in that order, with at least one sandbox=container-netns
// row in the table. A file that already carries a bracket line is refused.
func findContainerBlock(text string) (containerBlock, error) {
	if strings.Contains(text, "\r") {
		return containerBlock{}, fmt.Errorf("the witness contains carriage returns; its bytes would not hash as written")
	}
	if !strings.HasSuffix(text, "\n") {
		return containerBlock{}, fmt.Errorf("the witness does not end with a newline")
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, admissionBracketPrefix) {
			return containerBlock{}, fmt.Errorf("this witness already carries an %sline", admissionBracketPrefix)
		}
	}
	n := len(lines)
	if n < 4 || !strings.HasPrefix(lines[n-3], attestationPrefix) || !strings.HasPrefix(lines[n-2], containerImagePrefix) || !strings.HasPrefix(lines[n-1], treeDigestPrefix) {
		return containerBlock{}, fmt.Errorf("the witness does not end with the attestation, %s and %slines a container run records", strings.TrimSpace(containerImagePrefix), treeDigestPrefix)
	}
	m := attestationV1Re.FindStringSubmatch(lines[n-3])
	if m == nil {
		return containerBlock{}, fmt.Errorf("the witness's attestation line is not the version-1 form a container run records: %q", lines[n-3])
	}
	if !sha256DigestRe.MatchString(strings.TrimPrefix(lines[n-2], containerImagePrefix)) {
		return containerBlock{}, fmt.Errorf("malformed %q", lines[n-2])
	}
	tree := strings.TrimPrefix(lines[n-1], treeDigestPrefix)
	if !sha256DigestRe.MatchString(tree) {
		return containerBlock{}, fmt.Errorf("malformed %q", lines[n-1])
	}
	containerRows, start := 0, n-3
	for i := n - 4; i >= 0 && strings.HasPrefix(lines[i], "|"); i-- {
		start = i
		if witnessSandboxOf(lines[i]) == sandboxContainerNetns {
			containerRows++
		}
	}
	if containerRows == 0 {
		return containerBlock{}, fmt.Errorf("no witness row with sandbox=%s sits directly above the attestation line", sandboxContainerNetns)
	}
	if !witnessHeaderAt(lines, start) {
		return containerBlock{}, fmt.Errorf("the table above the attestation line does not start with the verifyrun witness header, so the hashed span has no defined start")
	}
	return containerBlock{lines: lines, span: lines[start:], attestation: lines[n-3], brief: m[5], tree: tree}, nil
}

// runCloseBracket is `verifyrun --close-bracket`: the host's second half. It
// appends the Admission-Bracket line only when the host check re-admits
// exactly the run the witness and the receipt name AND the home's recomputed
// tree digest equals both the receipt's tree and the witness's Tree-Digest.
// Any mismatch or refusal writes nothing.
func runCloseBracket(home, brief, receiptPath, witnessPath string, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "statusgen verifyrun --close-bracket: refused, nothing appended: "+format+"\n", a...)
		return verifyrunExitCouldNot
	}
	if rel, ok := relWithin(home, witnessPath); ok {
		return refuse("the witness %s sits inside the home (%s)", witnessPath, rel)
	}
	raw, err := os.ReadFile(witnessPath)
	if err != nil {
		return refuse("%v", err)
	}
	info, err := os.Stat(witnessPath)
	if err != nil {
		return refuse("%v", err)
	}
	block, err := findContainerBlock(string(raw))
	if err != nil {
		return refuse("%v", err)
	}
	receipt, receiptLine, _, err := readBracketReceipt(receiptPath, home, brief)
	if err != nil {
		return refuse("%v", err)
	}
	if receiptLine != block.attestation {
		return refuse("the receipt does not name the run the witness names:\n  witness: %s\n  receipt: %s", block.attestation, receiptLine)
	}
	check, err := hostCheckVerifier(home, brief)
	if err != nil {
		return refuse("%v", err)
	}
	hostLine, _, hostHome, err := verifierAdmissionReceipt(check, brief)
	if err != nil {
		return refuse("%v", err)
	}
	if hostHome != "" {
		return refuse("the host check returned a cross-repository receipt; container mode admits version 1 only")
	}
	if hostLine != block.attestation {
		return refuse("the host check does not re-admit the run the witness names (run, source, brief, model or tier differ):\n  witness: %s\n  host:    %s", block.attestation, hostLine)
	}
	tree, err := homeTreeDigest(home)
	if err != nil {
		return refuse("%v", err)
	}
	if tree != receipt.Tree {
		return refuse("the home's tree digest %s is not the %s recorded at the open step: the home moved between open and close", tree, receipt.Tree)
	}
	if tree != block.tree {
		return refuse("the home's tree digest %s is not the witness's Tree-Digest %s: the witness was not taken on this home", tree, block.tree)
	}
	opened, _ := time.Parse(time.RFC3339, receipt.Opened)
	closed := nowFunc().UTC()
	if closed.Before(opened) {
		return refuse("the receipt's opened time %s is after the close %s", receipt.Opened, closed.Format(time.RFC3339))
	}
	above := strings.Join(block.lines, "\n") + "\n"
	witness := witnessSpanHash(block.span)
	m := attestationV1Re.FindStringSubmatch(block.attestation)
	line := fmt.Sprintf("%srun=%s source=%s tree=%s witness=%s opened=%s closed=%s before=receipt after=check-verifier",
		admissionBracketPrefix, m[3], m[4], tree, witness, receipt.Opened, closed.Format(time.RFC3339))
	if err := os.WriteFile(witnessPath, []byte(above+line+"\n"), info.Mode().Perm()); err != nil {
		return refuse("%v", err)
	}
	fmt.Fprintln(stdout, line)
	return verifyrunExitPass
}

// The witness= hash of an Admission-Bracket line covers an explicit span,
// normalised so a landing does not break it (desktools-v2/71):
//
//   - span: from the verifyrun witness header line (`| # | Command | ...`) of
//     the table directly above the attestation line, through the Tree-Digest
//     line, each line ending in one newline. Nothing above the header (Evidence
//     prose, other tables, an acceptance descriptor) is in it.
//   - normalisation: every ` (on-behalf-of ...)` parenthetical in a table line
//     is deleted before hashing. deskevidence stamps exactly that annotation
//     into an App-run Runner cell that carries none when it lands Evidence, and
//     verifyrun writes the same annotation only when a principal resolves,
//     which inside a network-less container it may not. Deleting it from every
//     row makes the hash the same before and after a landing; the stamp itself
//     is bound by the landing, not by the bracket.
//
// A reviewer recomputes it with
//
//	sed -E 's/ \(on-behalf-of [^()]*\)//g' span.txt | sha256sum
//
// where span.txt holds the span's lines exactly.
var onBehalfOfParenRe = regexp.MustCompile(` \(on-behalf-of [^()]*\)`)

func witnessSpanHash(span []string) string {
	h := sha256.New()
	for _, l := range span {
		if strings.HasPrefix(l, "|") {
			l = onBehalfOfParenRe.ReplaceAllString(l, "")
		}
		io.WriteString(h, l+"\n")
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// witnessHeaderAt reports whether lines[i:] starts with a verifyrun witness
// header (the plain or the fail-first form, both two lines).
func witnessHeaderAt(lines []string, i int) bool {
	for _, header := range []string{witnessHeader, witnessHeaderFailFirst} {
		hl := strings.Split(header, "\n")
		if i+len(hl) <= len(lines) && strings.Join(lines[i:i+len(hl)], "\n") == header {
			return true
		}
	}
	return false
}
