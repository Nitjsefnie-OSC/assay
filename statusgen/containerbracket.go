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
//     the receipt's tree, runs the rows, digests its root again, and only when
//     the rows left it unchanged records Tree-Digest, Container-Image and the
//     attestation line under the table;
//  3. host close: `verifyrun --close-bracket` re-runs the check, compares it
//     with the witness's attestation line, recomputes the home's digest, and
//     only when it equals both the receipt's tree and the witness's
//     Tree-Digest appends an Admission-Bracket line whose witness= is
//     bracketDigest of the span above it.
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

// containerWitnessLines are the lines a container run writes under its
// witness table: Tree-Digest between the table and the attestation line (Task
// 3), the image digest beside it, the attestation last.
func containerWitnessLines(image, tree, attestation string) string {
	return treeDigestPrefix + tree + "\n" + containerImagePrefix + image + "\n" + attestation + "\n"
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
	attestation string
	brief       string // the attestation's home-relative brief
	tree        string
}

// findContainerBlock reads a container witness file: it must END with the
// witness table and then the Tree-Digest, Container-Image and version-1
// attestation lines, in that order, with at least one sandbox=container-netns
// row in the table directly above. A file that already carries a bracket line
// is refused.
func findContainerBlock(text string) (containerBlock, error) {
	if strings.Contains(text, "\r") {
		return containerBlock{}, fmt.Errorf("the witness contains carriage returns; a container run writes LF lines")
	}
	if !strings.HasSuffix(text, "\n") {
		return containerBlock{}, fmt.Errorf("the witness does not end with a newline")
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), strings.TrimSpace(admissionBracketPrefix)) {
			return containerBlock{}, fmt.Errorf("this witness already carries an %sline", admissionBracketPrefix)
		}
	}
	n := len(lines)
	if n < 4 || !strings.HasPrefix(lines[n-3], treeDigestPrefix) || !strings.HasPrefix(lines[n-2], containerImagePrefix) || !strings.HasPrefix(lines[n-1], attestationPrefix) {
		return containerBlock{}, fmt.Errorf("the witness does not end with the %s, %sand attestation lines a container run records", strings.TrimSpace(treeDigestPrefix), containerImagePrefix)
	}
	tree := strings.TrimPrefix(lines[n-3], treeDigestPrefix)
	if !sha256DigestRe.MatchString(tree) {
		return containerBlock{}, fmt.Errorf("malformed %q", lines[n-3])
	}
	if !sha256DigestRe.MatchString(strings.TrimPrefix(lines[n-2], containerImagePrefix)) {
		return containerBlock{}, fmt.Errorf("malformed %q", lines[n-2])
	}
	m := attestationV1Re.FindStringSubmatch(lines[n-1])
	if m == nil {
		return containerBlock{}, fmt.Errorf("the witness's attestation line is not the version-1 form a container run records: %q", lines[n-1])
	}
	containerRows, ownHeader := 0, false
	for i := n - 4; i >= 0 && strings.HasPrefix(lines[i], "|"); i-- {
		if isWitnessHeaderLine(lines[i]) {
			ownHeader = true
			break
		}
		if witnessSandboxOf(lines[i]) == sandboxContainerNetns {
			containerRows++
		}
	}
	if containerRows == 0 {
		return containerBlock{}, fmt.Errorf("no witness row with sandbox=%s sits directly above the %sline", sandboxContainerNetns, treeDigestPrefix)
	}
	if !ownHeader {
		if _, header := witnessSpanDigest(lines, n); header < 0 {
			return containerBlock{}, fmt.Errorf("no witness table header lies above the append position, so the bracket would have no span")
		}
		// The span would start at another table's header and take in that
		// table's Tree-Digest, image and attestation lines.
		return containerBlock{}, fmt.Errorf("the container rows above the %sline carry no witness table header of their own, so the bracket's span would start at another table", treeDigestPrefix)
	}
	return containerBlock{lines: lines, attestation: lines[n-1], brief: m[5], tree: tree}, nil
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
	witness, err := bracketDigest(above)
	if err != nil {
		return refuse("%v", err)
	}
	m := attestationV1Re.FindStringSubmatch(block.attestation)
	line := fmt.Sprintf("%srun=%s source=%s tree=%s witness=%s opened=%s closed=%s before=receipt after=check-verifier",
		admissionBracketPrefix, m[3], m[4], tree, witness, receipt.Opened, closed.Format(time.RFC3339))
	if err := os.WriteFile(witnessPath, []byte(above+line+"\n"), info.Mode().Perm()); err != nil {
		return refuse("%v", err)
	}
	fmt.Fprintln(stdout, line)
	return verifyrunExitPass
}

// The witness digest of an Admission-Bracket line (desktools-v2/71 Task 1).
// The close step writes it as witness= and the credit rule recomputes it with
// the read-only `statusgen verifyrun --bracket-digest --witness <file>`, so both
// go through witnessSpanDigest.
//
//   - header: a line equal, after trimming, to the first line of witnessHeader
//     or witnessHeaderFailFirst.
//   - span: fixed by the bracket's position alone. It starts at the nearest
//     witness header ABOVE the position and ends at the line immediately
//     before it. For an existing Admission-Bracket line the position is that
//     line; for the close step it is the line about to be appended. Other
//     witness tables in the same file, above or below, are outside the span.
//     A position with no header above it has no span.
//   - normalisation: trailing whitespace stripped, blank lines dropped, lines
//     joined with LF (no final LF), and each table row's Runner cell (the
//     column the span's header names Runner) rewritten in three steps: every
//     exact `(on-behalf-of human:<subject>)` removed (case-sensitive; subject
//     one or more characters other than whitespace and parentheses), every run
//     of spaces and tabs collapsed to one space, the cell trimmed. The bytes
//     between the cell's two delimiters become one space, the value and one
//     space; every other byte of the row is kept. deskevidence stamps exactly
//     that form, before any trailing source tag such as `(git-config)`, so the
//     stamp is outside the digest while every other byte of the span is in it.
//
// It is never a raw sha256sum of the landed bytes, which the stamp changes.

// witnessSpanDigest digests the span that ends just above lines[pos]. header
// is the span's header line index, -1 when no witness header lies above pos.
func witnessSpanDigest(lines []string, pos int) (digest string, header int) {
	header = -1
	for i := pos - 1; i >= 0; i-- {
		if isWitnessHeaderLine(lines[i]) {
			header = i
			break
		}
	}
	if header < 0 {
		return "", -1
	}
	runner := witnessRunnerColumn(lines[header])
	var kept []string
	for i := header; i < pos; i++ {
		l := strings.TrimRight(lines[i], " \t\r\v\f")
		if l == "" {
			continue
		}
		if i > header && strings.HasPrefix(strings.TrimSpace(l), "|") {
			l = normaliseRunnerCell(l, runner)
		}
		kept = append(kept, l)
	}
	sum := sha256.Sum256([]byte(strings.Join(kept, "\n")))
	return "sha256:" + hex.EncodeToString(sum[:]), header
}

// bracketDigest is the close step's digest: the span above the line it is
// about to append to text.
func bracketDigest(text string) (string, error) {
	lines := strings.Split(text, "\n")
	digest, header := witnessSpanDigest(lines, len(lines))
	if header < 0 {
		return "", fmt.Errorf("no witness table header lies above the append position, so the bracket would have no span")
	}
	return digest, nil
}

// bracketCheck is one `--bracket-digest` output line.
type bracketCheck struct {
	bracket, header   int // 1-based line numbers; header 0 when there is none
	witness, recorded string
	verdict           string // match | mismatch | no-table | shared
	why               string // a digest match the span's own lines do not support
}

func (c bracketCheck) String() string {
	header, witness := "none", "none"
	if c.header > 0 {
		header, witness = fmt.Sprintf("L%d", c.header), c.witness
	}
	return fmt.Sprintf("bracket=L%d header=%s witness=%s recorded=%s verdict=%s", c.bracket, header, witness, c.recorded, c.verdict)
}

var bracketWitnessRe = regexp.MustCompile(`(?:^|\s)witness=(\S+)`)

// bracketChecks checks every Admission-Bracket line in text, in file order.
// A bracket whose nearest header another bracket also has is `shared` (its
// rows would lie in two spans, so neither credits them).
func bracketChecks(text string) []bracketCheck {
	lines := strings.Split(text, "\n")
	var checks []bracketCheck
	byHeader := map[int]int{}
	for i, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), strings.TrimSpace(admissionBracketPrefix)) {
			continue
		}
		c := bracketCheck{bracket: i + 1, recorded: "none"}
		if m := bracketWitnessRe.FindStringSubmatch(l); m != nil {
			c.recorded = m[1]
		}
		digest, header := witnessSpanDigest(lines, i)
		if header >= 0 && digest == c.recorded {
			c.why = spanBindingProblem(lines[header:i], l)
		}
		switch {
		case header < 0:
			c.verdict = "no-table"
		case digest == c.recorded && c.why == "":
			c.header, c.witness, c.verdict = header+1, digest, "match"
		default:
			c.header, c.witness, c.verdict = header+1, digest, "mismatch"
		}
		if header >= 0 {
			byHeader[header]++
		}
		checks = append(checks, c)
	}
	for i := range checks {
		if checks[i].header > 0 && byHeader[checks[i].header-1] > 1 {
			checks[i].verdict = "shared"
		}
	}
	return checks
}

// isWitnessHeaderLine reports whether a line, trimmed, is the first line of a
// verifyrun witness header, plain or fail-first.
// bracketTreeRe reads an Admission-Bracket line's tree= value.
var bracketTreeRe = regexp.MustCompile(`(?:^|\s)tree=(\S+)`)

// spanBindingProblem says why a bracket whose digest matches is still not
// bound to its own span: the Tree-Digest, Container-Image and attestation lines
// credited with it must be the ones inside that span, one of each, and the
// bracket's tree= must equal that span's Tree-Digest, never a line from
// another bracket's span. It is "" when the span binds.
func spanBindingProblem(span []string, bracket string) string {
	var trees []string
	images, attestations := 0, 0
	for _, l := range span {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, strings.TrimSpace(treeDigestPrefix)):
			trees = append(trees, strings.TrimSpace(strings.TrimPrefix(l, strings.TrimSpace(treeDigestPrefix))))
		case strings.HasPrefix(l, strings.TrimSpace(containerImagePrefix)):
			images++
		case strings.HasPrefix(l, attestationPrefix):
			attestations++
		}
	}
	if len(trees) != 1 || images != 1 || attestations != 1 {
		return fmt.Sprintf("the span holds %d %s, %d %s and %d attestation lines, not one of each", len(trees), strings.TrimSpace(treeDigestPrefix), images, strings.TrimSpace(containerImagePrefix), attestations)
	}
	m := bracketTreeRe.FindStringSubmatch(bracket)
	if m == nil || m[1] != trees[0] {
		got := "none"
		if m != nil {
			got = m[1]
		}
		return fmt.Sprintf("the bracket's tree=%s is not its own span's %s %s", got, strings.TrimSpace(treeDigestPrefix), trees[0])
	}
	return ""
}

func isWitnessHeaderLine(l string) bool {
	l = strings.TrimSpace(l)
	for _, header := range []string{witnessHeader, witnessHeaderFailFirst} {
		if first, _, _ := strings.Cut(header, "\n"); l == first {
			return true
		}
	}
	return false
}

// witnessRunnerColumn is the Runner cell's index in a witness header line.
func witnessRunnerColumn(header string) int {
	for i, c := range splitRowEscaped(header) {
		if strings.TrimSpace(c) == "Runner" {
			return i
		}
	}
	return -1
}

// onBehalfOfStampRe is exactly the stamp deskevidence and verifyrun write.
// Nothing looser matches: `(On-Behalf-Of human:x)`, `(on-behalf-of human: x)`
// or `(on-behalf-of <desk role>)` stay in the digest.
var onBehalfOfStampRe = regexp.MustCompile(`\(on-behalf-of human:[^\s()]+\)`)

var cellWhitespaceRe = regexp.MustCompile(`[ \t]+`)

// normaliseRunnerCell rewrites a table row's Runner cell (cell index col,
// counted as splitRowEscaped counts them: `\|` is cell content): stamp
// removed, whitespace runs collapsed, trimmed. The rest of the row is kept
// byte for byte. A row with too few cells is returned unchanged.
func normaliseRunnerCell(row string, col int) string {
	if col < 0 {
		return row
	}
	var pipes []int
	for i := 0; i < len(row); i++ {
		switch {
		case row[i] == '\\' && i+1 < len(row) && row[i+1] == '|':
			i++
		case row[i] == '|':
			pipes = append(pipes, i)
		}
	}
	// The row's leading pipe opens cell 0, so cell col sits between pipes
	// col and col+1.
	if !strings.HasPrefix(strings.TrimSpace(row), "|") || len(pipes) < col+2 {
		return row
	}
	cell := onBehalfOfStampRe.ReplaceAllString(row[pipes[col]+1:pipes[col+1]], "")
	cell = strings.TrimSpace(cellWhitespaceRe.ReplaceAllString(cell, " "))
	return row[:pipes[col]+1] + " " + cell + " " + row[pipes[col+1]:]
}
