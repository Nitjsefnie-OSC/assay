//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func cellPrepUp(t *testing.T) {
	for _, n := range []int{1, 5} {
		t.Run(fmt.Sprintf("up%d", n), func(t *testing.T) {
			roles := strings.Join(upFiveRoles[:n], " ")
			w := newUpWorld(t, roles, 10*time.Millisecond)
			c := w.cell()
			hashes := retainedHistory(t, c, 10)
			began := time.Now()
			result := w.up(t, "--role", upFiveRoles[0])
			if n == 5 {
				result = w.up(t)
			}
			if result.Code != 0 {
				t.Fatalf("actual up exit=%d stderr=%s", result.Code, result.Stderr)
			}
			receipt, err := os.ReadFile(filepath.Join(c.Dir, "run", "scratch", "preparation.json"))
			must(t, err)
			again := w.up(t, "--role", upFiveRoles[0])
			if again.Code != 0 {
				t.Fatal(again.Stderr)
			}
			after, err := os.ReadFile(filepath.Join(c.Dir, "run", "scratch", "preparation.json"))
			must(t, err)
			if string(receipt) != string(after) {
				t.Fatal("attach repeated shared preparation")
			}
			assertHistory(t, hashes)
			t.Logf("actual-up roles=%d protected-files=500 samples=1 wall=%s", n, time.Since(began))
		})
	}
}
