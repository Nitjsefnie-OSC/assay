package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// The container bracket (desktools-v2/71). Every test here runs on fixtures:
// a git home with a committed brief, a fake deskdispatch on PATH that logs
// each call, and the --network none interface list stubbed in. No container
// engine, network or forge is used.

const (
	fixtureRun    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	fixtureSource = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	fixtureImage  = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	fixtureOpened = "2026-10-22T00:00:00Z"
)

type containerFixture struct {
	home, brief, marker string
	check               string // the fake deskdispatch prints this file
	receipt             string // the bracket receipt the container reads
	witness             string // where the container writes its witness
	log                 string
}

// newContainerFixture builds the home, its brief (one Verify row that touches
// a marker OUTSIDE the home) and a fake deskdispatch that prints the file at
// FAKE_DD_RECEIPT, exits FAKE_DD_EXIT, and logs its arguments.
func newContainerFixture(t *testing.T) containerFixture {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-only: the fake deskdispatch is a #!/bin/sh script on PATH")
	}
	stubIfaces(t, netnoneIfaces)
	t.Setenv("GITHUB_ACTIONS", "")
	home := t.TempDir()
	outside := t.TempDir()
	fx := containerFixture{
		home:    home,
		brief:   filepath.Join(home, "docs", "brief.md"),
		marker:  filepath.Join(outside, "ran"),
		check:   filepath.Join(outside, "check.json"),
		receipt: filepath.Join(outside, "receipt.json"),
		witness: filepath.Join(outside, "witness.md"),
		log:     filepath.Join(outside, "deskdispatch.log"),
	}
	gitInit(t, home, "fixture", "fixture@example.invalid")
	if err := os.MkdirAll(filepath.Dir(fx.brief), 0o755); err != nil {
		t.Fatal(err)
	}
	tick := "`"
	body := "# Fixture\n\n## Verify\n\n| # | Command | Expect |\n|---|---|---|\n| 1 | " + tick + "touch " + fx.marker + tick + " | exit 0 |\n\n## Evidence\n\n## Review\n\nnone\n"
	if err := os.WriteFile(fx.brief, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "source.txt"), []byte("attested source\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, home, "add", ".")
	runGit(t, home, "commit", "-m", "fixture")
	bin := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$FAKE_DD_LOG\"\ncat \"$FAKE_DD_RECEIPT\"\nexit \"${FAKE_DD_EXIT:-0}\"\n"
	if err := os.WriteFile(filepath.Join(bin, "deskdispatch"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_DD_LOG", fx.log)
	t.Setenv("FAKE_DD_RECEIPT", fx.check)
	t.Setenv("FAKE_DD_EXIT", "0")
	return fx
}

// checkJSON is a `deskdispatch --check-verifier` receipt.
func checkJSON(version int, run, source, brief, model string) string {
	extra := ""
	if version == 2 {
		extra = `,"SourceRepo":"example-org/two","TrackingSource":"` + source + `"`
	}
	top := `"Issue":7`
	if version == 2 {
		top += `,"BriefRoot":"/tracking","SourceHome":"/home"`
	}
	return fmt.Sprintf(`{%s,"Binding":{"Version":%d,"Run":%q,"Repo":"example-org/one","Source":%q,"Brief":%q,"Model":%q,"Tier":"strong"%s}}`, top, version, run, source, brief, model, extra)
}

var validCheck = checkJSON(1, fixtureRun, fixtureSource, "docs/brief.md", "m")

func putFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func getFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(raw)
}

// writeBracketReceipt writes the open step's receipt file directly, with the
// home's current tree digest, so a container test starts no deskdispatch.
func (fx containerFixture) writeBracketReceipt(t *testing.T, check string) {
	t.Helper()
	tree, err := homeTreeDigest(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	putFile(t, fx.receipt, `{"check":`+check+`,"tree":"`+tree+`","opened":"`+fixtureOpened+`"}`+"\n")
}

func (fx containerFixture) containerArgs(extra ...string) []string {
	return append([]string{"--sandbox=" + sandboxContainerNetns, "--root", fx.home, "--brief", fx.brief}, extra...)
}

func (fx containerFixture) runArgs() []string {
	return fx.containerArgs("--admission-receipt", fx.receipt, "--container-image-digest", fixtureImage, "--witness", fx.witness)
}

// assertNothingRan: no row executed, the brief is unchanged and no witness
// file was written.
func (fx containerFixture) assertNothingRan(t *testing.T, before string) {
	t.Helper()
	if _, err := os.Stat(fx.marker); err == nil {
		t.Fatal("a Verify row executed although the run was refused")
	}
	if got := getFile(t, fx.brief); got != before {
		t.Fatalf("a refused run changed the brief:\n%s", got)
	}
	if _, err := os.Stat(fx.witness); err == nil {
		t.Fatal("a refused run wrote a witness file")
	}
}

// regression: #4071
// TestContainerNetnsRequiresAdmissionReceipt: container-netns without a host
// receipt is refused before any row runs, with DESK_LOOP unset and no
// attestation file in the home (the case the conditional admission let
// through unattested), and equally with DESK_LOOP naming a verifier. The
// control proves the row does run once a receipt is given.
func TestContainerNetnsRequiresAdmissionReceipt(t *testing.T) {
	fx := newContainerFixture(t)
	for _, loop := range []string{"", "verify-desk"} {
		t.Setenv("DESK_LOOP", loop)
		before := getFile(t, fx.brief)
		res, stderr := captureVerifyrun(t, fx.containerArgs("--container-image-digest", fixtureImage, "--witness", fx.witness))
		if res.code != verifyrunExitUsageError || !strings.Contains(stderr, "without --admission-receipt") {
			t.Fatalf("DESK_LOOP=%q: exit %d stderr %q, want the missing-receipt usage refusal", loop, res.code, stderr)
		}
		fx.assertNothingRan(t, before)
	}
	if calls := getFile(t, fx.log); calls != "" {
		t.Fatalf("a container run started deskdispatch: %q", calls)
	}

	// Control: the same run with a receipt executes the row.
	t.Setenv("DESK_LOOP", "")
	fx.writeBracketReceipt(t, validCheck)
	res, stderr := captureVerifyrun(t, fx.runArgs())
	if res.code != verifyrunExitPass {
		t.Fatalf("control: exit %d stderr %q", res.code, stderr)
	}
	if _, err := os.Stat(fx.marker); err != nil {
		t.Fatal("control: the row did not run with a receipt")
	}
}

// regression: #4071
// TestContainerNetnsReceipt: a valid version-1 receipt yields exactly the
// attestation line the in-process path renders, plus the image and
// Tree-Digest lines; a version-2 receipt, a brief mismatch, an incomplete
// receipt, a receipt inside the home, a missing or malformed image digest,
// and a mounted tree that does not digest to the receipt's tree (one file
// changed, one file added) are each refused before any row runs.
// deskdispatch is never started, even with DESK_LOOP naming a verifier.
func TestContainerNetnsReceipt(t *testing.T) {
	fx := newContainerFixture(t)
	t.Setenv("DESK_LOOP", "verify-desk")
	source := filepath.Join(fx.home, "source.txt")
	refusals := []struct {
		name, check, why string
		args             []string
		mutate           func()
		restore          func()
	}{
		{name: "version 2", check: checkJSON(2, fixtureRun, fixtureSource, "docs/brief.md", "m"), why: "version 2 (cross-repository) receipts are not admitted"},
		{name: "brief mismatch", check: checkJSON(1, fixtureRun, fixtureSource, "docs/other.md", "m"), why: "the receipt admits brief docs/other.md"},
		{name: "incomplete", check: checkJSON(1, fixtureRun, fixtureSource, "docs/brief.md", ""), why: "incomplete pre-work verifier receipt"},
		{name: "short run", check: checkJSON(1, "r", fixtureSource, "docs/brief.md", "m"), why: "incomplete receipt"},
		{name: "missing digest", check: validCheck, why: "without --container-image-digest",
			args: []string{"--admission-receipt", "RECEIPT", "--witness", "WITNESS"}},
		{name: "malformed digest", check: validCheck, why: `"sha256:XYZ"`,
			args: []string{"--admission-receipt", "RECEIPT", "--container-image-digest", "sha256:XYZ", "--witness", "WITNESS"}},
		{name: "changed file", check: validCheck, why: "not the sha256:",
			mutate:  func() { putFile(t, source, "TAMPERED source\n") },
			restore: func() { putFile(t, source, "attested source\n") }},
		{name: "added file", check: validCheck, why: "not the sha256:",
			mutate: func() { putFile(t, filepath.Join(fx.home, "planted.txt"), "planted\n") },
			restore: func() {
				if err := os.Remove(filepath.Join(fx.home, "planted.txt")); err != nil {
					t.Fatal(err)
				}
			}},
	}
	for _, c := range refusals {
		fx.writeBracketReceipt(t, c.check)
		if c.mutate != nil {
			c.mutate()
		}
		args := fx.runArgs()
		if c.args != nil {
			args = fx.containerArgs(c.args...)
			for i := range args {
				switch args[i] {
				case "RECEIPT":
					args[i] = fx.receipt
				case "WITNESS":
					args[i] = fx.witness
				}
			}
		}
		before := getFile(t, fx.brief)
		res, stderr := captureVerifyrun(t, args)
		if res.code == verifyrunExitPass || !strings.Contains(stderr, c.why) {
			t.Fatalf("%s: exit %d stderr %q, want a refusal naming %q", c.name, res.code, stderr, c.why)
		}
		fx.assertNothingRan(t, before)
		if c.restore != nil {
			c.restore()
		}
	}

	// A receipt inside the home is refused: it would be an input the rows see.
	fx.writeBracketReceipt(t, validCheck)
	inside := filepath.Join(fx.home, "receipt.json")
	putFile(t, inside, getFile(t, fx.receipt))
	before := getFile(t, fx.brief)
	if res, stderr := captureVerifyrun(t, fx.containerArgs("--admission-receipt", inside, "--container-image-digest", fixtureImage, "--witness", fx.witness)); res.code == verifyrunExitPass || !strings.Contains(stderr, "inside the home") {
		t.Fatalf("a receipt inside the home: exit %d stderr %q", res.code, stderr)
	}
	fx.assertNothingRan(t, before)
	if err := os.Remove(inside); err != nil {
		t.Fatal(err)
	}

	// A row that changes the mounted home leaves no witness: the tree the
	// rows end on must be the tree they started on.
	planted := filepath.Join(fx.home, "planted-by-row")
	putFile(t, fx.brief, strings.Replace(before, "touch "+fx.marker, "touch "+planted, 1))
	fx.writeBracketReceipt(t, validCheck)
	if res, stderr := captureVerifyrun(t, fx.runArgs()); res.code == verifyrunExitPass || !strings.Contains(stderr, "the rows changed the mounted home") {
		t.Fatalf("a row that changed the home: exit %d stderr %q", res.code, stderr)
	}
	if _, err := os.Stat(fx.witness); err == nil {
		t.Fatal("a run whose rows changed the home wrote a witness")
	}
	if err := os.Remove(planted); err != nil {
		t.Fatal(err)
	}
	putFile(t, fx.brief, before)
	fx.writeBracketReceipt(t, validCheck)

	res, stderr := captureVerifyrun(t, fx.runArgs())
	if res.code != verifyrunExitPass {
		t.Fatalf("valid receipt: exit %d stderr %q", res.code, stderr)
	}
	want, _, _, err := verifierAdmissionReceipt([]byte(validCheck), fx.brief)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := homeTreeDigest(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	got := getFile(t, fx.witness)
	if !strings.HasSuffix(got, "\n"+want+"\n"+containerImagePrefix+fixtureImage+"\n"+treeDigestPrefix+tree+"\n") {
		t.Fatalf("the witness does not end with the in-process attestation line %q, the image line and the Tree-Digest line:\n%s", want, got)
	}
	if !strings.Contains(got, "sandbox="+sandboxContainerNetns) {
		t.Fatalf("the witness row does not record the container sandbox:\n%s", got)
	}
	if getFile(t, fx.brief) != before {
		t.Fatal("a container run wrote into the brief inside the home")
	}
	if calls := getFile(t, fx.log); calls != "" {
		t.Fatalf("a container run started deskdispatch: %q", calls)
	}
}

// regression: #4071
// TestCloseBracket: the open step writes a receipt only on an admitted home;
// the close step appends one Admission-Bracket line only when the host check
// re-admits the run the witness names and the home's recomputed tree digest
// equals both the receipt's tree and the witness's Tree-Digest. A witness
// Tree-Digest that differs, a home that moved between open and close, a
// changed source, a different run, a refused host check and an already
// bracketed witness each append nothing and exit non-zero.
func TestCloseBracket(t *testing.T) {
	fx := newContainerFixture(t)
	t.Setenv("DESK_LOOP", "")
	openArgs := []string{"--open-bracket", "--root", fx.home, "--brief", fx.brief, "--out", fx.receipt}

	// A refused open check writes no receipt.
	putFile(t, fx.check, validCheck)
	t.Setenv("FAKE_DD_EXIT", "1")
	if res, stderr := captureVerifyrun(t, openArgs); res.code == verifyrunExitPass || !strings.Contains(stderr, "host admission check failed") {
		t.Fatalf("refused open check: exit %d stderr %q", res.code, stderr)
	}
	if _, err := os.Stat(fx.receipt); err == nil {
		t.Fatal("a refused open check wrote a receipt")
	}
	t.Setenv("FAKE_DD_EXIT", "0")
	if res, stderr := captureVerifyrun(t, openArgs); res.code != verifyrunExitPass {
		t.Fatalf("open: exit %d stderr %q", res.code, stderr)
	}
	if res, stderr := captureVerifyrun(t, fx.runArgs()); res.code != verifyrunExitPass {
		t.Fatalf("container run: exit %d stderr %q", res.code, stderr)
	}
	witness := getFile(t, fx.witness)
	tree := regexp.MustCompile(`(?m)^Tree-Digest: (\S+)$`).FindStringSubmatch(witness)
	if tree == nil {
		t.Fatalf("no Tree-Digest line in the witness:\n%s", witness)
	}
	closeArgs := []string{"--close-bracket", "--root", fx.home, "--brief", fx.brief, "--admission-receipt", fx.receipt, "--witness", fx.witness}
	source := filepath.Join(fx.home, "source.txt")
	otherTree := "sha256:" + strings.Repeat("e", 64)

	refusals := []struct {
		name, check, exit, why string
		witness                string
		tamper                 bool
	}{
		{"witness from another tree", validCheck, "0", "not the witness's Tree-Digest", strings.Replace(witness, tree[1], otherTree, 1), false},
		{"home moved between open and close", validCheck, "0", "recorded at the open step", witness, true},
		{"changed source", checkJSON(1, fixtureRun, strings.Repeat("c", 40), "docs/brief.md", "m"), "0", "does not re-admit", witness, false},
		{"different run", checkJSON(1, strings.Repeat("d", 64), fixtureSource, "docs/brief.md", "m"), "0", "does not re-admit", witness, false},
		{"refused host check", validCheck, "1", "host admission check failed", witness, false},
		{"already bracketed", validCheck, "0", "already carries", witness + admissionBracketPrefix + "run=x\n", false},
		{"no witness header", validCheck, "0", "no defined start", strings.SplitN(witness, "\n", 3)[2], false},
	}
	for _, c := range refusals {
		putFile(t, fx.check, c.check)
		putFile(t, fx.witness, c.witness)
		t.Setenv("FAKE_DD_EXIT", c.exit)
		if c.tamper {
			putFile(t, source, "TAMPERED source\n")
		}
		res, stderr := captureVerifyrun(t, closeArgs)
		if res.code == verifyrunExitPass || !strings.Contains(stderr, c.why) {
			t.Fatalf("%s: exit %d stderr %q, want a refusal naming %q", c.name, res.code, stderr, c.why)
		}
		if got := getFile(t, fx.witness); got != c.witness {
			t.Fatalf("%s: a refused close changed the witness:\n%s", c.name, got)
		}
		if c.tamper {
			putFile(t, source, "attested source\n")
		}
	}

	putFile(t, fx.check, validCheck)
	// Close on a witness whose Runner cell already carries an on-behalf-of
	// annotation, the way verifyrun writes it when a principal resolves: the
	// hash is over the span with the annotation deleted, never the raw bytes.
	rows := strings.Split(witness, "\n")
	stamp := func(row string) string { return strings.TrimSuffix(row, " |") + " (on-behalf-of human:someone) |" }
	annotated := strings.Join(append(append(append([]string{}, rows[:2]...), stamp(rows[2])), rows[3:]...), "\n")
	if sha256.Sum256([]byte(annotated)) == sha256.Sum256([]byte(docNormalise(annotated))) {
		t.Fatal("the fixture annotation did not change the raw bytes; the normalisation is untested")
	}
	putFile(t, fx.witness, annotated)
	t.Setenv("FAKE_DD_EXIT", "0")
	res, stderr := captureVerifyrun(t, closeArgs)
	if res.code != verifyrunExitPass {
		t.Fatalf("matching close: exit %d stderr %q", res.code, stderr)
	}
	got := getFile(t, fx.witness)
	above, line, _ := strings.Cut(got, admissionBracketPrefix)
	if above != annotated || strings.Count(got, admissionBracketPrefix) != 1 || !strings.HasSuffix(line, "\n") || strings.Count(line, "\n") != 1 {
		t.Fatalf("the close did not append exactly one line under the unchanged witness:\n%s", got)
	}
	want := fmt.Sprintf(`^run=%s source=%s tree=%s witness=sha256:%x opened=%s closed=\S+ before=receipt after=check-verifier\n$`,
		fixtureRun, fixtureSource, regexp.QuoteMeta(tree[1]), sha256.Sum256([]byte(docNormalise(annotated))), fixtureOpenedOrAny)
	if !regexp.MustCompile(want).MatchString(line) {
		t.Fatalf("the bracket line %q does not match %s", line, want)
	}
	// A landing of an UNannotated witness puts it under Evidence prose and
	// stamps an on-behalf-of annotation into the App Runner cell. The span
	// (header line through Tree-Digest) with the annotation deleted, as the
	// documented sed recipe does, hashes to the same witness= value; the raw
	// stamped bytes would not.
	hash := regexp.MustCompile(`witness=(sha256:[0-9a-f]{64})`).FindStringSubmatch(line)[1]
	landed := "## Evidence\n\nVerifier prose above the block.\n\n" + annotated + admissionBracketPrefix + line
	span := landed[strings.LastIndex(landed, "| # | Command |"):strings.Index(landed, admissionBracketPrefix)]
	if got := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(docNormalise(span)))); got != hash {
		t.Fatalf("the landed, stamped span hashes to %s, not the bracket's %s:\n%s", got, hash, span)
	}
	if fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(span))) == hash {
		t.Fatal("the fixture stamp did not change the raw bytes; the normalisation is untested")
	}
	if !strings.Contains(getFile(t, fx.log), "--check-verifier --root "+fx.home+" --brief "+fx.brief) {
		t.Fatalf("the close step did not run the host check: %q", getFile(t, fx.log))
	}
	if res, stderr := captureVerifyrun(t, closeArgs); res.code == verifyrunExitPass || !strings.Contains(stderr, "already carries") {
		t.Fatalf("a closed witness was closed a second time: exit %d stderr %q", res.code, stderr)
	}
	if again := getFile(t, fx.witness); again != got {
		t.Fatalf("a second close changed the witness:\n%s", again)
	}
}

// fixtureOpenedOrAny matches the open step's own opened time, which is the
// wall clock at the open.
const fixtureOpenedOrAny = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`

// docNormalise is the reviewer's sed recipe from the docs, applied in Go:
// sed -E 's/ \(on-behalf-of [^()]*\)//g'.
func docNormalise(s string) string {
	return regexp.MustCompile(` \(on-behalf-of [^()]*\)`).ReplaceAllString(s, "")
}
