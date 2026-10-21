package deskkit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func snapshotObserver(t *testing.T) func(string) int {
	t.Helper()
	activationIndex.Lock()
	activationIndex.slot = nil
	activationIndex.Unlock()
	old := activationWork
	counts := map[string]int{}
	var mu sync.Mutex
	activationWork = func(event string) { mu.Lock(); counts[event]++; mu.Unlock() }
	t.Cleanup(func() {
		activationWork = old
		activationIndex.Lock()
		activationIndex.slot = nil
		activationIndex.Unlock()
	})
	return func(event string) int { mu.Lock(); defer mu.Unlock(); return counts[event] }
}
func snapshotParity(t *testing.T, root string) {
	t.Helper()
	got, skips, err := activationManifests(root)
	want, wantSkips, wantErr := DiscoverManifests(root)
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(skips, wantSkips) || fmt.Sprint(err) != fmt.Sprint(wantErr) {
		t.Fatalf("snapshot differs from reference: got=%+v %v %v want=%+v %v %v", got, skips, err, want, wantSkips, wantErr)
	}
}
func snapshotFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

// regression: F-activation-snapshot
func TestActivationSnapshotMatchesFreshDiscovery(t *testing.T) {
	count := snapshotObserver(t)
	root := t.TempDir()
	writeManifest(t, root, "test/verbs", []string{"assay.desk.verbs"}, nil, nil)
	snapshotParity(t, root)
	before := count("enumeration")
	snapshotParity(t, root)
	// snapshotParity itself calls the reference exactly once.
	if count("enumeration") != before+1 || count("reuse") != 1 || count("validation-directory") == 0 {
		t.Fatal("warm fixture repeated full enumeration or hid validation")
	}
	changes := []struct {
		name  string
		apply func()
	}{
		{"nested untracked", func() {
			snapshotFile(t, filepath.Join(root, "untracked", "deep", "component.yaml"), "component: test/new\n")
		}},
		{"ignored", func() {
			snapshotFile(t, filepath.Join(root, ".gitignore"), "ignored/\n")
			snapshotFile(t, filepath.Join(root, "ignored", "component.yaml"), "component: test/ignored\n")
		}},
		{"dirty", func() {
			snapshotFile(t, filepath.Join(root, "test-verbs", "component.yaml"), "component: test/edited\n")
		}},
		{"duplicate", func() { snapshotFile(t, filepath.Join(root, "a", "component.yaml"), "component: test/edited\n") }},
		{"malformed", func() { snapshotFile(t, filepath.Join(root, "bad", "component.yaml"), "not: [valid: yaml") }},
		{"missing name", func() { snapshotFile(t, filepath.Join(root, "empty", "component.yaml"), "provides: []\n") }},
		{"removal", func() {
			if err := os.Remove(filepath.Join(root, "ignored", "component.yaml")); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range changes {
		t.Run(c.name, func(t *testing.T) { c.apply(); snapshotParity(t, root); snapshotParity(t, root) })
	}
	// In-place data tampering by a caller cannot poison subsequent reuse.
	ms, _, _ := activationManifests(root)
	ms[0].Component = "poison"
	snapshotParity(t, root)
}

// regression: F-activation-snapshot
func TestActivationSnapshotIsNotAuthority(t *testing.T) {
	count := snapshotObserver(t)
	root := t.TempDir()
	writeManifest(t, root, "test/verbs", []string{"assay.desk.verbs"}, []string{"test.provider"}, nil)
	writeManifest(t, root, "test/provider", []string{"test.provider"}, []string{"assay.roster.trust", "assay.roster.ext.repo-aliases"}, nil)
	withRoster(t, goldenRoster())
	if msg := VerbActivationRefusal(root); msg != "" {
		t.Fatal(msg)
	}
	r := goldenRoster()
	r[EnvRepoAliases] = "not=a=valid=shape"
	withRoster(t, r)
	if msg := VerbActivationRefusal(root); !strings.Contains(msg, "INACTIVE") {
		t.Fatalf("cached discovery became authority: %q", msg)
	}
	withRoster(t, map[string]string{})
	if msg := VerbActivationRefusal(root); !strings.Contains(msg, "unset or malformed") {
		t.Fatalf("lower API did not refuse current trust: %q", msg)
	}
	if count("reuse") < 2 {
		t.Fatal("configuration checks did not exercise reused discovery")
	}
	withRoster(t, goldenRoster())
	writeManifest(t, root, "test/provider", []string{"test.provider"}, []string{"test.changed-requirement"}, nil)
	if msg := VerbActivationRefusal(root); !strings.Contains(msg, "test.changed-requirement") {
		t.Fatal("provider edit did not recompute activation:", msg)
	}
	snapshotParity(t, root)
	// A root spelling reused after replacement is not the same source.
	other := t.TempDir()
	writeManifest(t, other, "test/other", []string{"assay.desk.verbs"}, []string{"test.missing"}, nil)
	snapshotParity(t, other)
	snapshotParity(t, root)
	old := root + "-old"
	if err := os.Rename(root, old); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(old) })
	snapshotFile(t, filepath.Join(root, "component.yaml"), "component: test/replaced\nprovides: [assay.desk.verbs]\ninject:\n  required: [{key: test.missing}]\n")
	if msg := VerbActivationRefusal(root); !strings.Contains(msg, "test.missing") {
		t.Fatal("root replacement reused acceptance:", msg)
	}
}

// regression: F-activation-snapshot
func TestActivationSnapshotCorruptionAndBounds(t *testing.T) {
	snapshotObserver(t)
	root := t.TempDir()
	writeManifest(t, root, "test/verbs", []string{"assay.desk.verbs"}, nil, nil)
	snapshotParity(t, root)
	for _, mode := range []string{"corrupt", "oversized", "unsupported"} {
		t.Run(mode, func(t *testing.T) {
			activationIndex.Lock()
			slot := activationIndex.slot
			if slot == nil {
				t.Fatal("expected populated slot")
			}
			switch mode {
			case "corrupt":
				slot.payload = []byte(`{"Revision":"activation-discovery/1","Manifests":[]}`) // valid JSON, wrong checksum
			case "oversized":
				slot.payload = make([]byte, manifestLimit+1)
				slot.checksum = sha256.Sum256(slot.payload)
			case "unsupported":
				var p manifestPayload
				_ = json.Unmarshal(slot.payload, &p)
				p.Revision = "future"
				slot.payload, _ = json.Marshal(p)
				slot.checksum = sha256.Sum256(slot.payload)
			}
			activationIndex.Unlock()
			snapshotParity(t, root)
		})
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); snapshotParity(t, root) }()
	}
	wg.Wait()
	nested := filepath.Join(root, "nested")
	snapshotFile(t, filepath.Join(nested, "component.yaml"), "component: test/nested\n")
	snapshotParity(t, root)
	for _, marker := range []string{"file", "directory", "removed"} {
		switch marker {
		case "file":
			snapshotFile(t, filepath.Join(nested, ".git"), "gitdir: absent\n")
		case "directory":
			if err := os.Remove(filepath.Join(nested, ".git")); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(nested, ".git"), 0700); err != nil {
				t.Fatal(err)
			}
		case "removed":
			if err := os.Remove(filepath.Join(nested, ".git")); err != nil {
				t.Fatal(err)
			}
		}
		snapshotParity(t, root)
	}
	snapshotFile(t, filepath.Join(root, "large", "component.yaml"), "component: test/large\n#"+strings.Repeat("x", manifestLimit))
	snapshotParity(t, root)
	if activationIndex.slot != nil {
		t.Fatal("oversized source retained")
	}
}
