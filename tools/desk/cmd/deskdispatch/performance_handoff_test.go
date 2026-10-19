package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/medici-finance/assay/perfrec"
)

// regression: desktools-v2/64
func TestPerfPrivateHandoff(t *testing.T) {
	cache := t.TempDir()
	oldCache, oldNow := packetCacheDir, packetNow
	t.Cleanup(func() { packetCacheDir, packetNow = oldCache, oldNow })
	packetCacheDir = func() (string, error) { return cache, nil }
	now := time.Now().UTC()
	packetNow = func() time.Time { return now }
	t.Setenv("ASSAY_TELEMETRY", "1")
	c := perfrec.ResolveContext(perfrec.ContextInput{Profile: perfrec.Value{Value: "workflow", Present: true}, AgentID: perfrec.Value{Value: strings.Repeat("b", 32), Present: true}, DeskRole: perfrec.Value{Value: "worker-desk", Present: true}, Role: perfrec.Value{Value: "review", Present: true}})
	if got := writePerformanceHandoff(perfrec.Context{}); got != "" {
		t.Fatal("disabled context wrote handoff")
	}
	dir := filepath.Join(cache, "assay", "packets")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("disabled context touched cache")
	}
	path := writePerformanceHandoff(c)
	if path == "" {
		t.Fatal("missing handoff")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("handoff not private")
	}
	info, err = os.Stat(dir)
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatal("directory not private")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Environment      []string
		DeskRole         string `json:"deskRoleOverride"`
		ExpiresBeforeUTC string `json:"expiresBeforeUtc"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ExpiresBeforeUTC != time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7).Format(time.RFC3339) {
		t.Fatal("handoff expiration is not seven UTC days")
	}
	if payload.DeskRole != "worker-desk" || !strings.Contains(strings.Join(payload.Environment, "\n"), "ASSAY_PERF_AGENT_ID="+strings.Repeat("b", 32)) {
		t.Fatal("handoff lost typed context")
	}
	if strings.Contains(string(data), "ASSAY_TELEMETRY") {
		t.Fatal("handoff manufactured consent")
	}
	prompt, err := assemblePrompt(dispatchOpts{kit: "worker"}, dispatchPlan{performancePath: path, childPerformance: c}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, path) || strings.Contains(prompt, strings.Repeat("b", 32)) || strings.Contains(prompt, "ASSAY_PERF_AGENT_ID=") {
		t.Fatal("prompt exposed attribution or lost handoff")
	}
	old := filepath.Join(dir, "expired"+performanceSuffix)
	if err := os.WriteFile(old, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	expired := midnight.AddDate(0, 0, -7)
	if err := os.Chtimes(old, expired, expired); err != nil {
		t.Fatal(err)
	}
	if writePerformanceHandoff(c) == "" {
		t.Fatal("second handoff failed")
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("expired handoff retained")
	}
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if writePerformanceHandoff(c) != "" {
		t.Fatal("public directory accepted")
	}
}
