package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

// The cell database rows of `cellctl check` (docs/postgres.md, docs/cellctl.md). A cell points at
// its database with two cell.env keys: CELL_DB_DSN, a libpq connection string carrying no
// password, and CELL_DB_PASSFILE, the pgpass file libpq reads the password from. launch.go
// composes them into every role window as ASSAY_CELL_DB_DSN and PGPASSFILE.
//
// What these rows prove is custody of the pointer, nothing more: that the DSN does not carry the
// password itself and that the credential file is not readable by another OS user. They never
// claim that cells are separated from each other — on one OS user they are not at the credential
// layer (docs/postgres.md §"Isolation boundary").

// dsnPasswordKeyword matches libpq's `password` keyword in the key=value form, or as a query
// parameter of the URI form, case-insensitively, with optional spaces around the `=`.
var dsnPasswordKeyword = regexp.MustCompile(`(?i)(^|[\s?&])password\s*=`)

// dsnCarriesPassword reports whether a libpq connection string carries a password: a `password=`
// keyword or URI parameter, or a password in the userinfo of a postgres:// or postgresql:// URI
// (`user:secret@host`). The URI is read by hand rather than with net/url, which refuses libpq's
// multi-host form (`host1:5432,host2:5432`).
func dsnCarriesPassword(dsn string) bool {
	if dsnPasswordKeyword.MatchString(dsn) {
		return true
	}
	lower := strings.ToLower(dsn)
	for _, scheme := range []string{"postgresql://", "postgres://"} {
		if !strings.HasPrefix(lower, scheme) {
			continue
		}
		authority := dsn[len(scheme):]
		if i := strings.IndexAny(authority, "/?#"); i >= 0 {
			authority = authority[:i]
		}
		at := strings.LastIndex(authority, "@")
		if at < 0 {
			return false
		}
		return strings.Contains(authority[:at], ":")
	}
	return false
}

// checkCellDB prints the cell database rows: a MISS on a container cell (which cannot reach the
// database yet), and on any other kind the DSN and credential-file rows. With CELL_DB_DSN unset
// it prints nothing, as checkRoleContext does for CELL_ROLE_CONTEXT: every `check` transcript of
// a cell with no database stays byte-identical, which is what the shell-oracle parity harness and
// the pre-migration golden transcripts compare (an unconditional n/a row would need an oracle
// edit).
func (c *Cell) checkCellDB(k *checker) {
	dsn, passfile := c.cellDBValues()
	if dsn == "" {
		return
	}
	if c.Kind == "container" {
		// A container cell's plan (cell-containers-v1) refuses any environment or mount it does
		// not name, and the cell runs on its own network, so the DSN and the pgpass file cannot
		// reach it without a schema change. Fail rather than leave the key silently unused.
		k.chk(false, "cell database: container cells cannot reach the cell database yet — the cell-containers-v1 plan carries no CELL_DB_DSN environment, no pgpass mount and no route to the loopback-only server; unset CELL_DB_DSN or run this cell as a host kind")
		return
	}
	if dsnCarriesPassword(dsn) {
		// The value is withheld: it holds a password.
		k.chk(false, "cell database: CELL_DB_DSN carries a password (a password= keyword or URI userinfo) — remove it; libpq reads the password from CELL_DB_PASSFILE")
	} else {
		k.chk(true, "cell database: CELL_DB_DSN carries no password: %s", dsn)
	}
	switch {
	case passfile == "":
		k.chk(false, "cell database: CELL_DB_PASSFILE is set (the pgpass file exported as PGPASSFILE)")
	case !filepath.IsAbs(passfile):
		k.chk(false, "cell database: CELL_DB_PASSFILE is an absolute path: %s", passfile)
	case !isRegular(passfile) || !isReadable(passfile):
		k.chk(false, "cell database: CELL_DB_PASSFILE is a readable file: %s", passfile)
	default:
		// The same helper as the PEM and config-home mode rows. The mode keeps out other OS
		// users only; every host cell runs as the same user (docs/postgres.md).
		k.chk(modeIs(passfile, 0o600) || modeIs(passfile, 0o400),
			"cell database: CELL_DB_PASSFILE mode 0600 or 0400: %s", passfile)
	}
}
