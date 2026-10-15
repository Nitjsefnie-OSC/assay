package main

// ASSAY_TELEMETRY is the one opt-in a launched desk's wrapped tools read (deskkit.PerfEnabled
// and statusgen's telemetry switch both enable ONLY on the exact value "1"). cellctl composes
// every launched child's environment from an explicit allowlist, so a cell.env or defaults.env
// that configures the opt-in reached no tool until each launch profile carried the key itself.
//
// The contract is the same on every profile that carries it:
//   - ONE key, never the launching shell's whole environment;
//   - the exact value "1" opts in; an explicitly configured anything-else normalizes to "0"
//     (explicit off stays off, even over a launching shell that exported "1");
//   - an unset key stays absent, so a desk nobody configured is byte-identical to before;
//   - precedence is the Env's own: cell.env over defaults.env over the launching process.
//
// Profiles: house and Codex carry the key (launch.go, codex_environment.go); scrubbed carries it
// through the composed allowlist (env.go). Container and operator-owned launchers (container.go,
// the celllaunch runner contract) compose their own environment and do NOT carry it; that
// limitation is documented in docs/cellctl.md rather than implied away.

const envAssayTelemetry = "ASSAY_TELEMETRY"

// telemetryValue resolves the opt-in for this cell: "" when no layer configures the key (leave
// it absent), "1" for exactly "1", "0" for any other configured value.
func (c *Cell) telemetryValue() string {
	if c == nil || c.Env == nil || !c.Env.IsSet(envAssayTelemetry) {
		return ""
	}
	return normalizeTelemetry(c.Env.Get(envAssayTelemetry))
}

// normalizeTelemetry maps a configured value to the canonical "1" / "0". Exact comparison on
// purpose: the consumers compare with ==, so " 1" or "true" would be OFF for them and must not
// be forwarded as if on.
func normalizeTelemetry(v string) string {
	if v == "1" {
		return "1"
	}
	return "0"
}

// telemetryLaunchEnv returns env with the resolved opt-in applied (a copy; env is untouched).
// Absent configuration leaves env as it came, so a desk nobody opted in is unchanged.
func (c *Cell) telemetryLaunchEnv(env []string) []string {
	v := c.telemetryValue()
	if v == "" {
		return env
	}
	return envSet(append([]string(nil), env...), envAssayTelemetry, v)
}
