//go:build unix

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestVerifyrunSignalChild is the subprocess half of
// TestVerifyrunSignalCancelsRow: it runs the real verifyrun entry point so the
// parent can signal it the way a terminal or supervisor would. Inert unless the
// parent set its environment.
func TestVerifyrunSignalChild(t *testing.T) {
	root := os.Getenv("ASSAY_SIGNAL_CHILD_ROOT")
	if root == "" {
		return
	}
	args := []string{"--brief", os.Getenv("ASSAY_SIGNAL_CHILD_BRIEF"), "--root", root, "--dry-run", "--timeout", "30s"}
	code := runVerifyrun(args, os.Stdout, os.Stderr)
	fmt.Printf("child-exit=%d\n", code)
	os.Exit(0)
}

// TestVerifyrunSignalCancelsRow: every row runs in its own process group, so a
// stop signal sent to statusgen's group never reaches the row directly. The
// signal must still stop the row (its delayed work never happens), the row is
// could-not-run, the next row never starts, and no witness is written.
// regression: #4360
func TestVerifyrunSignalCancelsRow(t *testing.T) {
	for _, tc := range []struct {
		sig  syscall.Signal
		name string
	}{
		{syscall.SIGINT, "SIGINT"},
		{syscall.SIGTERM, "SIGTERM"},
		{syscall.SIGHUP, "SIGHUP"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := admittedFixture(t, false)
			brief := filepath.Join(t.TempDir(), "brief.md")
			tick := string(rune(96))
			body := "# Fixture\n\n## Verify\n\n| # | Command | Expect |\n|---|---|---|\n" +
				"| 1 | " + tick + "touch started; sleep 1.5; touch escaped" + tick + " | exit 0 |\n" +
				"| 2 | " + tick + "touch second-ran" + tick + " | exit 0 |\n\n## Evidence\n\n"
			if err := os.WriteFile(brief, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestVerifyrunSignalChild$")
			cmd.Env = append(os.Environ(), "ASSAY_SIGNAL_CHILD_ROOT="+root, "ASSAY_SIGNAL_CHILD_BRIEF="+brief)
			// Its own group, as a terminal's foreground job or a supervised job is.
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			pgid := cmd.Process.Pid
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			defer func() { _ = syscall.Kill(-pgid, syscall.SIGKILL) }()

			deadline := time.Now().Add(20 * time.Second)
			for {
				if _, err := os.Stat(filepath.Join(root, "started")); err == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("row 1 never started; output:\n%s", out.String())
				}
				time.Sleep(20 * time.Millisecond)
			}
			signalled := time.Now()
			if err := syscall.Kill(-pgid, tc.sig); err != nil {
				t.Fatal(err)
			}
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatalf("statusgen did not exit after %s; output:\n%s", tc.name, out.String())
			}
			// Wait past row 1's delayed work before looking for it.
			if remaining := 2*time.Second - time.Since(signalled); remaining > 0 {
				time.Sleep(remaining)
			}
			if _, err := os.Stat(filepath.Join(root, "escaped")); !os.IsNotExist(err) {
				t.Errorf("row 1 outlived %s and did its delayed work; stat=%v", tc.name, err)
			}
			if _, err := os.Stat(filepath.Join(root, "second-ran")); !os.IsNotExist(err) {
				t.Errorf("row 2 started after %s; stat=%v", tc.name, err)
			}
			got := out.String()
			for _, want := range []string{
				"row 1: could-not-run",
				"interrupted by " + tc.name,
				"row 2: could-not-run",
				"no witness was written",
				"child-exit=2",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("output lacks %q:\n%s", want, got)
				}
			}
			for _, bad := range []string{"row 1: pass", "row 1: fail", "row 2: pass", "row 2: fail"} {
				if strings.Contains(got, bad) {
					t.Errorf("an interrupted row carries a verdict (%q):\n%s", bad, got)
				}
			}
		})
	}
}
