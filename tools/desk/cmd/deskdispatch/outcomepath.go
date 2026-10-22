package main

// outcomepath.go — where a verifier writes its outcome record before landing it.
//
// The verifier home must stay exactly the admitted source: an untracked file in it refuses
// `--check-verifier` ("unattested worktree file"). The outcome record therefore has to be
// written somewhere else and handed to `deskevidence --outcome-record` as an explicit input,
// the same way an Evidence fragment is. Without a named place, each verifier invented one,
// and a path under the outcome-records directory inside the home was the obvious wrong guess.
// This file names that place, per run, the way packetPathFor names the packet.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// outcomeDirName is the directory beside --prompt-file that holds one subdirectory per claim.
const outcomeDirName = "outcomes"

// outcomeDirFor decides where this verifier run's outcome record goes, without creating it.
//
//   - With --prompt-file F, beside F: "<directory of F>/outcomes/<claim key>/".
//   - Without it, "<cache>/assay/outcomes/<claim key>/" in the user's cache directory —
//     never the agent's worktree, never the configuration directory.
//
// Both forms end in a per-claim subdirectory, so runs that share a prompt directory never
// share an outcome directory. Each branch validates the claim key (outcomeClaimDir) before
// it becomes that path component.
//
// It REFUSES, and never falls back to another location, when the resolved directory is the
// verifier home or the dispatcher root, or sits under either: a record written there is an
// untracked file in a tree that admission requires to be exactly the attested source.
// Containment is checked on symlink-resolved paths, so an alias cannot hide the home.
// home may be "" on a dry run, where no worktree exists yet; the root check still applies.
func outcomeDirFor(o dispatchOpts, plan dispatchPlan, home string) (string, error) {
	var dir string
	if pf := strings.TrimSpace(o.promptFile); pf != "" {
		abs, err := filepath.Abs(pf)
		if err != nil {
			return "", err
		}
		if dir, err = outcomeClaimDir(filepath.Join(filepath.Dir(abs), outcomeDirName), plan.claimKey); err != nil {
			return "", err
		}
	} else {
		cache, err := packetCacheDir()
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(cache) {
			return "", fmt.Errorf("cache directory %q is not absolute", cache)
		}
		if dir, err = outcomeClaimDir(filepath.Join(cache, "assay", "outcomes"), plan.claimKey); err != nil {
			return "", err
		}
	}
	for _, guarded := range []struct{ what, path string }{{"verifier home", home}, {"dispatcher root", o.root}} {
		p := strings.TrimSpace(guarded.path)
		if p == "" {
			continue
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if vpWithin(p, dir) {
			return "", deskkit.Refused(fmt.Sprintf("refused: the outcome-record directory %s resolves inside the %s %s; "+
				"an outcome record must be written OUTSIDE it (an untracked file there refuses verifier admission). "+
				"Keep --prompt-file outside both, or omit it to use the user cache directory", dir, guarded.what, p))
		}
	}
	return dir, nil
}

// outcomeClaimDir joins the claim key under base as ONE path component. The key is
// validated as it stands, never rewritten: a key that sanitizing would change (a separator,
// a "..") or that fails the item-key shape is refused, so a traversal-shaped key can neither
// climb out of base nor be silently renamed into a different claim's directory. The joined
// result must sit directly under base.
func outcomeClaimDir(base, claimKey string) (string, error) {
	key := strings.TrimSpace(claimKey)
	if key == "" || key != claimKey || sanitizeSegment(key) != key || strings.ContainsAny(key, `/\`) ||
		strings.Contains(key, "..") || !itemKeyRe.MatchString(key) {
		return "", deskkit.Refused(fmt.Sprintf("refused: claim key %q does not make a single outcome-directory name", claimKey))
	}
	base = filepath.Clean(base)
	dir := filepath.Join(base, key)
	if filepath.Dir(dir) != base || filepath.Base(dir) != key {
		return "", deskkit.Refused(fmt.Sprintf("refused: claim key %q does not stay one directory under %s", claimKey, base))
	}
	return dir, nil
}

// allocateOutcomeDir resolves the directory (outcomeDirFor), prunes week-old cache
// directories as packets are pruned, and creates it owner-only. A path that already exists
// as anything but a real directory (a symlink, a file) is refused rather than followed.
func allocateOutcomeDir(o dispatchOpts, plan dispatchPlan, home string) (string, error) {
	dir, err := outcomeDirFor(o, plan, home)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(o.promptFile) == "" {
		pruneOutcomeDirs(filepath.Dir(dir), packetNow().Add(-packetMaxAge))
	}
	if fi, lerr := os.Lstat(dir); lerr == nil && !fi.IsDir() {
		return "", deskkit.Refused("refused: the outcome-record path " + dir + " exists and is not a directory")
	} else if lerr != nil && !errors.Is(lerr, os.ErrNotExist) {
		return "", lerr
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// pruneOutcomeDirs removes this tool's own per-claim outcome directories in dir last
// modified before cutoff. Best effort and narrow: only real directories (never a symlink,
// never a file) directly under the outcomes cache directory.
func pruneOutcomeDirs(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || e.Type()&os.ModeSymlink != 0 {
			continue
		}
		if info, err := e.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(dir, e.Name()))
		}
	}
}

// writeOutcomeRecordLine names the outcome directory and the exact landing command. Only the
// verifier assignment calls it, and only when a directory was resolved.
func writeOutcomeRecordLine(b *strings.Builder, plan dispatchPlan, repo, home string) {
	if plan.outcomePath == "" {
		return
	}
	fmt.Fprintf(b, "## Outcome record — write it OUTSIDE the home\n\n"+
		"Outcome directory (dispatcher-named, absolute, outside the home): `%s`. Write the outcome record there "+
		"and nowhere else: never write the record inside the home `%s` (an untracked file there refuses the "+
		"admission check). Land it with exactly:\n\n```\ndeskevidence %s main --outcome-record %s --root %s\n```\n\n",
		plan.outcomePath, home, repo, filepath.Join(plan.outcomePath, "<file>.json"), home)
}
