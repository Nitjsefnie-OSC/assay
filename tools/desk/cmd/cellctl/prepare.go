package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/cellscratch"
	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// prepareLaunch is the shared bounded cell preparation operation. Only storage
// initialization is shared. Role worktrees, policy, stop flags and execution
// leases remain with the existing per-role launch owner on every invocation.
func (c *Cell) prepareLaunch() (bool, error) {
	began := time.Now()
	if _, err := c.scratchPolicy(); err != nil {
		return false, err
	}
	if err := c.cacheAdmission(); err != nil {
		return false, err
	}
	s, err := cellscratch.Open(filepath.Join(c.Dir, "run", "scratch"))
	if err != nil {
		return false, err
	}
	defer s.Close()
	build, _ := deskkit.Version()
	values := map[string]string{}
	if c.Env != nil {
		for k, v := range c.Env.vals {
			if strings.HasPrefix(k, "CELL_") || strings.Contains(k, "MODEL_") || k == "ASSAY_OPERATOR_GENERATION" {
				values[k] = v
			}
		}
	}
	if c.Env != nil && c.Env.Get("CELL_MODEL_POLICY") != "" {
		profile, _, err := c.cellModelPolicy()
		if err != nil {
			return false, err
		}
		values["profile-digest"] = profile.SHA256
	}
	// Resolved config and the generation of the existing launch owner are part
	// of the binding; the digest is local and never logged or sent to telemetry.
	b, err := json.Marshal(struct {
		Cell, Kind, Build string
		Values            map[string]string
	}{c.Dir, c.Kind, build, values})
	if err != nil {
		return false, err
	}
	digest := sha256.Sum256(b)
	fresh, err := s.Prepare(hex.EncodeToString(digest[:]))
	if err == nil && fresh {
		c.recordLaunchStep("launch-preparation", time.Since(began), nil)
	}
	return fresh, err
}
