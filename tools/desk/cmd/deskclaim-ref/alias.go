package main

import (
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// Per-PR key spellings (medici-finance/assay#3700).
//
// A claim on one PR is keyed `<label>--pr-<N>` (correctness lane) or `<label>--pr-<N>--security`
// (security lane), where <label> is whatever short repo spelling the CALLER uses. Two callers
// that spell the label differently write two refs under refs/dispatch/, so `acquire` on the
// second read the PR as free although it was actively owned, and two sessions drove one PR.
//
// The claim lives in the PR's own repo and a PR number is unique within a repo, so every claim
// with the same <N> and the same lane suffix IS the same claim; no alias table is needed. The
// lanes stay separate (the suffix, not the caller, decides the lane), and keys that are not
// per-PR keys have no alias. This mirrors tools/dispatch-claim.sh in repos that ship it: the two
// tools speak one protocol, so they must agree on what counts as the same claim.

var prKeyRe = regexp.MustCompile(`^.+--pr-([0-9]+)(--security)?$`)

// prKeyParts returns the PR number and lane ("" or "--security") of a per-PR key.
func prKeyParts(id string) (num, lane string, ok bool) {
	m := prKeyRe.FindStringSubmatch(id)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

// aliasIDs lists the OTHER claim ids spelling the same PR and lane as id, in sorted order. The
// status is ClaimReadHeld when the claim set was read (even if no alias exists) and
// ClaimReadUnverifiable when it could not be: never "assume none".
func aliasIDs(id string) ([]string, deskkit.ClaimReadStatus) {
	num, lane, ok := prKeyParts(id)
	if !ok {
		return nil, deskkit.ClaimReadHeld
	}
	ids, status := store.List()
	if status == deskkit.ClaimReadUnverifiable {
		return nil, status
	}
	var out []string
	for _, other := range ids {
		if other == id {
			continue
		}
		if n, l, ok := prKeyParts(other); ok && n == num && l == lane {
			out = append(out, other)
		}
	}
	sort.Strings(out)
	return out, deskkit.ClaimReadHeld
}

// claimLive reports whether the claim under id is LIVE: its recorded branch exists on the remote
// (branch-as-claim), or it is inside its state's TTL. live=false with verifiable=true is a free
// or expired claim; verifiable=false is could-not-check.
func claimLive(id string) (ref deskkit.ClaimStoreRecord, live, verifiable bool) {
	ref, status := store.Read(id)
	switch status {
	case deskkit.ClaimReadFree:
		return ref, false, true
	case deskkit.ClaimReadUnverifiable:
		return ref, false, false
	}
	age, ok := ageMinutes(ref.Date)
	if !ok {
		return ref, false, false
	}
	if exists, ok := store.BranchExists(fieldOf(ref.Msg, "branch")); ok && exists {
		return ref, true, true
	}
	ttl := dispatchedTTL()
	if fieldOf(ref.Msg, "state") == "claimed" {
		ttl = claimedTTL()
	}
	return ref, age < ttl, true
}

// checkAliases refuses when another spelling of the same PR and lane is live. phase "pre" runs
// before the create. Phase "post" runs after our create succeeded and settles two spellings
// created in the same instant: the claim with the earlier forge-stamped date wins (ties go to
// the smaller id), and the loser removes its own ref and refuses. stop=false means clear.
func checkAliases(id, phase string) (rc int, stop bool) {
	others, status := aliasIDs(id)
	if status == deskkit.ClaimReadUnverifiable {
		errf("unverifiable: could not list dispatch claims to check other spellings of %s%s", id, causeSuffix())
		return exitUnverifiable, true
	}
	for _, other := range others {
		ref, live, verifiable := claimLive(other)
		if !verifiable {
			errf("unverifiable: could not read the claim on %s, another spelling of %s%s", other, id, causeSuffix())
			return exitUnverifiable, true
		}
		if !live {
			continue
		}
		if phase == "post" {
			mine, st := store.Read(id)
			if st != deskkit.ClaimReadHeld {
				errf("unverifiable: could not re-read our own claim on %s%s", id, causeSuffix())
				return exitUnverifiable, true
			}
			won, ok := claimsBefore(id, mine.Date, other, ref.Date)
			if !ok {
				errf("unverifiable: unreadable claim date on %s or %s", id, other)
				return exitUnverifiable, true
			}
			if won {
				continue
			}
			if outcome, _ := store.Remove(id); outcome != deskkit.ClaimWriteApplied {
				errf("unverifiable: lost the race to %s but could not release %s/%s%s", other, refPrefix, id, causeSuffix())
				return exitUnverifiable, true
			}
		}
		reportHolder(other, ref.Msg, ref.Date)
		logf("  %s is another spelling of the same PR as %s — not dispatching", other, id)
		return exitRefused, true
	}
	return exitOK, false
}

// claimsBefore reports whether claim a sorts before claim b: earlier forge-stamped date first,
// then the smaller id. ok=false when either date is unreadable.
func claimsBefore(aID, aDate, bID, bDate string) (before, ok bool) {
	at, err1 := time.Parse(time.RFC3339, aDate)
	bt, err2 := time.Parse(time.RFC3339, bDate)
	if err1 != nil || err2 != nil {
		return false, false
	}
	if !at.Equal(bt) {
		return at.Before(bt), true
	}
	return aID < bID, true
}

// printAliasHeld is what `show` appends for every live alias of id: a FREE here is not free when
// another spelling of the same PR holds it. A claim set that cannot be read adds nothing; the
// verb's own line already says what it read.
func printAliasHeld(id string) {
	others, status := aliasIDs(id)
	if status == deskkit.ClaimReadUnverifiable {
		return
	}
	for _, other := range others {
		ref, live, verifiable := claimLive(other)
		if !verifiable || !live {
			continue
		}
		ageStr := "?"
		if age, ok := ageMinutes(ref.Date); ok {
			ageStr = strconv.Itoa(age)
		}
		logf("ALIAS-HELD %s — %s at=%s age=%sm (same PR as %s)", other, ref.Msg, ref.Date, ageStr, id)
	}
}
