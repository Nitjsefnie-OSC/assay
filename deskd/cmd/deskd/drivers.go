package main

// drivers.go is the ONE file that decides which terminal drivers a deskd
// binary links (the service contract). internal/terminal's registry.go deferred
// this co-ordination point across briefs 13/14/15. It is decided here.
//
// Each driver self-registers from its init(), so importing the package is what
// links it. deskd links the container and local drivers only. It does NOT
// import internal/terminal/pod: per DR-deskd-attach-start no deskd path may
// build the pod driver or contact a cluster API server, whatever kubeconfig or
// in-cluster ServiceAccount the process can see. Leaving the pod driver out of
// the binary is the control that holds for every terminal.DriverFor consumer
// in deskd (the attach supervisor, a session open, a send, a future timer),
// not only for the supervisor, which refuses a pod desk on its own as a
// separate layer. TestDeskdLinksContainerAndLocalOnly and
// `go list -deps ./cmd/deskd` (Verify row 2) pin it. Linking the pod driver is
// a separate human sign-off on cluster access for deskd.
//
// Nothing else under cmd/, internal/serve or internal/mcp imports a driver
// package (Verify row 2), so the read surface never links client-go or the
// container CLI path by accident.

import (
	"github.com/medici-finance/assay/deskd/internal/identity"
	_ "github.com/medici-finance/assay/deskd/internal/terminal/container"
	"github.com/medici-finance/assay/deskd/internal/terminal/local"
)

// configureDrivers hands the local driver deskd's real bind arguments, so its
// belt-and-braces P3 re-check runs against what deskd is actually bound to.
// Call it once, before the attach supervisor starts.
func configureDrivers(addr string, mode identity.AuthMode, validatorConfigured bool) {
	local.SetBindGate(addr, mode, validatorConfigured)
}
