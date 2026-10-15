package main

import (
	"fmt"
	"strings"

	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
)

// readact.go — condition reviewer-approved on the READ + ACT install tier (#3062).
//
// THE SHAPE. `deskapps init --tier team` creates two Apps, `<prefix>-read` and
// `<prefix>-act`, and binds EVERY desk role — reviewer and worker included — to the act App.
// So the App that reviews a worker's PR is the App that AUTHORED it, and the forge refuses an
// account approving its own pull request. On that tier the reviewer App's verdicts on its own
// PRs can only ever be COMMENTS: the APPROVED at head the full-suite gate requires cannot
// exist, and the lane used to loop forever (CHECK → re-dispatch → another comment).
//
// THE GATE ON THAT SHAPE. The second actor the author ≠ approver property needs is the
// HUMAN: the tier's own documentation has always said "Approves a PR: you". So when, and only
// when, the roster binds the reviewer role to the SAME App as the worker role, the correctness
// gate is satisfied by an APPROVED at the CURRENT head from an ACCOUNTABLE TRUSTED HUMAN —
// deskkit.TrustedHumanAuthor, the roster's own accountable-human set (ASSAY_HUMAN_LOGIN_MAP
// plus the blessing authority). It is the same set the rest of the tools already use for
// "a person, not a role App or a shared machine account"; this file invents no second one.
//
// WHAT DOES NOT CHANGE.
//   - The full-suite shape (a reviewer App distinct from the worker App) runs
//     checkReviewerApproved exactly as before. A human APPROVED is never consulted there — not
//     as a substitute, not as a tie-breaker.
//   - An unbound reviewer role is still could-not-check (flip's RequireRole), never this lane.
//   - On the shared-App shape the reviewer App's OWN decisive verdict at head still governs
//     first, through the unchanged checkReviewerApproved: a standing CHANGES_REQUESTED still
//     blocks (with its two declared exemptions, unchanged), and an APPROVED still passes —
//     which the forge only lets the App post on a PR it did NOT author, so it is still a
//     second actor. That arm additionally refuses an APPROVED whose author IS the bound App
//     (a forge that does not enforce the self-approval block), so it can never be the hole.
//   - Every other condition — checks-green, mergeable, desk-decided, model-floor,
//     security-verdict, head-stable, caller-role, pr-open-draft — is untouched.

// workerRole is the roster role whose App authors desk PRs. Read only to recognise the
// shared-App (Read + Act) shape; never compared against a review author here.
const workerRole = "worker"

// readActTierName is how a refusal names the shape, so an operator recognises the tier from
// the install-time choice that produced it.
const readActTierName = "Read + Act"

// sharedReviewerApp reports whether the roster binds the reviewer role to the SAME App as the
// worker role — the Read + Act tier's shape. An unbound worker role is NOT that shape: it is a
// roster with no worker App, and the reviewer lane stays the full-suite one.
func sharedReviewerApp(reviewerLogin string) bool {
	worker, ok := deskkit.RoleAppLogin(workerRole)
	if !ok {
		return false
	}
	return deskkit.SameActor(worker, reviewerLogin)
}

// checkCorrectnessGate is condition reviewer-approved's single entry point. It returns the
// identity whose verdict satisfied the gate, for the OK line. sharedApp is decided ONCE per
// flip by the caller from the roster, so the pre-mutation re-check runs the same lane.
func checkCorrectnessGate(sharedApp bool, reviewerLogin, author string, reviews []reviewInfo, head string,
	live liveBodyRead, pr int, runsAtHead func() ([]deskkit.CheckRun, error)) (string, error) {
	if !sharedApp {
		// Full suite: unchanged.
		return reviewerLogin, checkReviewerApproved(reviewerLogin, reviews, head, live, pr, runsAtHead)
	}
	return checkReadActApproved(reviewerLogin, author, reviews, head, live, pr, runsAtHead)
}

// checkReadActApproved is the correctness gate on the shared-App shape. See the file doc.
func checkReadActApproved(reviewerLogin, author string, reviews []reviewInfo, head string,
	live liveBodyRead, pr int, runsAtHead func() ([]deskkit.CheckRun, error)) (string, error) {
	lane := correctnessLane(reviews)

	// 1. The bound App's own DECISIVE verdict governs first, through the unchanged gate. On
	//    this tier that can only be a verdict on a PR the App did not author (the forge blocks
	//    the rest), so it is still a second actor — and a standing block still blocks.
	switch deskkit.ReduceAppVerdict(reviewerLogin, lane, head) {
	case deskkit.AppVerdictApproved, deskkit.AppVerdictChangesRequested:
		if err := checkReviewerApproved(reviewerLogin, reviews, head, live, pr, runsAtHead); err != nil {
			return "", err
		}
		if strings.TrimSpace(author) == "" {
			return "", deskkit.Unverifiable(fmt.Sprintf(
				"condition %s: the roster binds the reviewer and worker roles to one App (%s) — the %s tier — "+
					"and PR #%d's author could not be read, so whether %s's APPROVED is a second actor's "+
					"cannot be established.", condReviewerApproved, deskkit.StripControl(reviewerLogin),
				readActTierName, pr, deskkit.StripControl(reviewerLogin)), nil)
		}
		if deskkit.SameActor(author, reviewerLogin) {
			return "", deskkit.Refused(fmt.Sprintf(
				"condition %s: %s APPROVED PR #%d, which it authored — on the %s tier the reviewer and "+
					"worker roles are one App, so that approval is the author approving itself and is not "+
					"a second actor's. Awaiting a human approval at head %s.",
				condReviewerApproved, deskkit.StripControl(reviewerLogin), pr, readActTierName, short(head)))
		}
		return reviewerLogin, nil
	}

	// 2. No decisive verdict from the App at head (None or Stale): the tier's documented
	//    approver is the human. Reduce each ACCOUNTABLE trusted human's own reviews to their
	//    latest decisive verdict, with the SAME reducer the App lane uses.
	var humans []string
	seen := map[string]bool{}
	var ignored []string
	ignoredSeen := map[string]bool{}
	for _, r := range lane {
		login := strings.TrimSpace(r.AuthorLogin)
		if login == "" {
			continue
		}
		key := strings.ToLower(login)
		if deskkit.SameActor(login, reviewerLogin) {
			continue
		}
		if !deskkit.TrustedHumanAuthor(login) || deskkit.SameActor(login, author) {
			if r.State == "APPROVED" && !ignoredSeen[key] {
				ignoredSeen[key] = true
				ignored = append(ignored, deskkit.StripControl(login))
			}
			continue
		}
		if !seen[key] {
			seen[key] = true
			humans = append(humans, login)
		}
	}

	approvedBy := ""
	staleLogin, staleAt := "", ""
	for _, h := range humans {
		switch deskkit.ReduceAppVerdict(h, lane, head) {
		case deskkit.AppVerdictChangesRequested:
			// Fail closed: a trusted human's standing block at head is decisive whoever else
			// approved.
			return "", deskkit.Refused(fmt.Sprintf(
				"condition %s: on the %s tier the correctness gate is a trusted human's verdict, and %s's "+
					"latest verdict at %s is CHANGES_REQUESTED — blocked.",
				condReviewerApproved, readActTierName, deskkit.StripControl(h), short(head)))
		case deskkit.AppVerdictApproved:
			if approvedBy == "" {
				approvedBy = h
			}
		case deskkit.AppVerdictStale:
			if staleLogin == "" {
				staleLogin = h
				staleAt = "an unknown commit"
				if last, ok := deskkit.LastAppDecisiveReview(h, lane); ok && last.CommitID != "" {
					staleAt = short(last.CommitID)
				}
			}
		}
	}
	if approvedBy != "" {
		return approvedBy, nil
	}

	tier := fmt.Sprintf("the roster binds the reviewer and worker roles to one App (%s) — the %s tier, where "+
		"that App cannot approve a PR it authored, so its review verdicts are comments and the approver is a "+
		"trusted human", deskkit.StripControl(reviewerLogin), readActTierName)
	note := ""
	if len(ignored) > 0 {
		note = fmt.Sprintf(" An APPROVED from %s was not counted: not an accountable human in the roster "+
			"(or the PR's own author).", strings.Join(ignored, ", "))
	}
	if staleLogin != "" {
		return "", deskkit.Refused(fmt.Sprintf(
			"condition %s: %s. %s's latest verdict on PR #%d is pinned at %s, but the PR's current head is "+
				"%s — a verdict that is not at this code is stale; awaiting a human approval at head %s.%s",
			condReviewerApproved, tier, deskkit.StripControl(staleLogin), pr, staleAt, short(head),
			short(head), note))
	}
	return "", deskkit.Refused(fmt.Sprintf(
		"condition %s: %s. No trusted human has APPROVED PR #%d at %s — awaiting a human approval at head.%s",
		condReviewerApproved, tier, pr, short(head), note))
}

// correctnessLane is checkReviewerApproved's rule 1, shared: every review EXCEPT those
// carrying a security marker, lifted into the reducer's shape. A security pass must never
// satisfy the correctness gate — on either tier.
func correctnessLane(reviews []reviewInfo) []deskkit.AppReview {
	lane := make([]deskkit.AppReview, 0, len(reviews))
	for _, r := range reviews {
		if hasSecurityMarker(r.Body) {
			continue
		}
		lane = append(lane, deskkit.AppReview{
			AuthorLogin: r.User.Login,
			State:       r.State,
			CommitID:    r.CommitID,
			SubmittedAt: r.SubmittedAt,
		})
	}
	return lane
}
