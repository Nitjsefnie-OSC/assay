//go:build !darwin && !linux

package main

import "testing"

func cellPrepUp(t *testing.T) { t.Skip("existing cockpit lease fixture requires Unix") }
