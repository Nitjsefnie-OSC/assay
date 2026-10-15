package main

// STUB (fail-first red run only): replaced by the real implementation in the fix commit.

const envAssayTelemetry = "ASSAY_TELEMETRY"

func (c *Cell) telemetryValue() string { return "" }

func (c *Cell) telemetryLaunchEnv(env []string) []string { return env }
