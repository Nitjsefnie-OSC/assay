//go:build darwin || linux

package cellcache

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// regression: #3353
func TestCacheScanAdmission(t *testing.T) {
	p := fixture(t)
	_, err := Prepare(p)
	if err != nil {
		t.Fatal(err)
	}
	// Warm immutable top-level directories make the existing recorded admission
	// usable without waiving the filesystem timestamp margin.
	old := time.Now().Add(-time.Hour)
	for _, name := range []string{"build", "mod"} {
		if err = os.Chtimes(filepath.Join(p.Root, name), old, old); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = Check(p, false); err != nil {
		t.Fatal(err)
	}
	m, err := open(p, false)
	if err != nil {
		t.Fatal(err)
	}
	defer m.close()
	paused, resume := make(chan struct{}), make(chan struct{})
	calls := 0
	m.free = func(string) (uint64, error) {
		calls++
		if calls == 1 {
			close(paused)
			<-resume
		}
		return bigFree, nil
	}
	done := make(chan Report, 1)
	go func() { done <- m.check(false) }()
	<-paused
	if m.lock != nil {
		// Prove the fixture detects the old global lock before doing a
		// second enrollment. Release only the deliberately broken fixture's
		// lock so its worker can finish rather than timing out the test.
		m.lock.Release()
		close(resume)
		<-done
		t.Fatal("maintenance retained admission custody during inventory")
	}
	lease, err := Acquire(p.Env())
	if err != nil {
		close(resume)
		<-done
		t.Fatalf("maintenance held admission lock: %v", err)
	}
	// A consumer admitted during the scan changes data and exits before the
	// maintenance transaction resumes. Active-record listing alone misses this.
	material(t, p, "build/new-output", 10)
	if err = lease.Finish(true); err != nil {
		t.Fatal(err)
	}
	close(resume)
	r := <-done
	if r.Outcome != "could-not-check" || r.Reclaimed != 0 || !strings.Contains(strings.Join(r.Failures, " "), "enrollment changed") {
		t.Fatalf("stale inventory gained authority: %+v", r)
	}
	b, err := os.ReadFile(filepath.Join(p.Root, "build/new-output"))
	if err != nil || !bytes.Equal(b, bytes.Repeat([]byte("x"), 10)) {
		t.Fatal("new consumer data removed", err)
	}
}

// regression: #3353
func TestCacheAdmissionFences(t *testing.T) {
	for _, kind := range []string{"missing", "corrupt", "stale", "dirty", "space", "domain"} {
		t.Run(kind, func(t *testing.T) {
			p := fixture(t)
			_, err := Prepare(p)
			if err != nil {
				t.Fatal(err)
			}
			var report Report
			b, err := os.ReadFile(p.ReportPath())
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(b, &report); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "missing":
				if err = os.Remove(p.ReportPath()); err != nil {
					t.Fatal(err)
				}
			case "corrupt":
				if err = os.WriteFile(p.ReportPath(), []byte("broken"), 0600); err != nil {
					t.Fatal(err)
				}
			case "stale":
				report.After.At = time.Now().Add(-time.Hour)
				b, _ = json.Marshal(report)
				if err = os.WriteFile(p.ReportPath(), b, 0600); err != nil {
					t.Fatal(err)
				}
			case "dirty":
				material(t, p, "build/new", 1)
			case "space":
				p.Floor = ^uint64(0) - 1
			case "domain":
				p.Domain = strings.Repeat("a", 64)
			}
			if l, e := Acquire(p.Env()); e == nil {
				l.Finish(true)
				t.Fatal("unproven accounting admitted a child")
			}
		})
	}
}
