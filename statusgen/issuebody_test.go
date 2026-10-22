package main

import (
	"strings"
	"testing"
)

const ibLink = "[docs/streams/x/brief-01.md](https://github.com/example-org/tracker/blob/0123abc/docs/streams/x/brief-01.md)"

func ibLinkFn() string { return ibLink }

// ibFill returns a section of n table rows, about 60 characters each.
func ibFill(heading string, n int) string {
	var b strings.Builder
	b.WriteString(heading + "\n\n")
	for i := 0; i < n; i++ {
		b.WriteString("| 1 | `go test ./...` | 0 | ok — résumé 😀 | 2026-10-21 | x |\n")
	}
	return b.String() + "\n"
}

// regression: #4091
func TestForgeBodyLen(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"é", 1}, // two bytes, one character
		{"😀", 2}, // four bytes, one code point, two UTF-16 units
		{"a\nb", 4},
	} {
		if got := forgeBodyLen(tc.in); got != tc.want {
			t.Errorf("forgeBodyLen(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// regression: #4091
func TestCapBodyFitsUnchanged(t *testing.T) {
	head, tail := "<!-- verify-gate: x/01 -->\n\n## H\n\n", "### Before you close\n"
	sections := []string{"### Verify\n\nrow\n\n", "### Evidence\n\nrow\n\n"}
	called := false
	got := capIssueBody(head, sections, tail, func() string { called = true; return ibLink })
	if want := head + strings.Join(sections, "") + tail; got != want {
		t.Errorf("a body under the cap must pass through unchanged;\ngot  %q\nwant %q", got, want)
	}
	if called {
		t.Error("the link is needed only for a cut, but was rendered for a body that fits")
	}
}

// TestCapBodyCutsLargestFirst pins the water-filling: a small section
// (Trade-offs) after a huge one stays whole, and only the huge one is cut.
//
// regression: #4091
func TestCapBodyCutsLargestFirst(t *testing.T) {
	head := "<!-- verify-gate: x/01 -->\n\n## Human sign-off required — x/01\n\n"
	tail := "### Before you close\n\n- [ ] I accept this brief as done\n"
	big := ibFill("### Recorded results (Evidence)", 3000)
	small := "### Trade-offs\n\nWhy we want it: cards stay fileable.\n\n"
	got := capIssueBody(head, []string{big, small}, tail, ibLinkFn)
	if n := forgeBodyLen(got); n > issueBodyCap {
		t.Fatalf("capped body is %d, over the cap %d", n, issueBodyCap)
	}
	if !strings.HasPrefix(got, head) || !strings.HasSuffix(got, tail) {
		t.Error("head and tail must be kept whole")
	}
	if !strings.Contains(got, small) {
		t.Errorf("the small section must be kept whole; body tail:\n%s", got[len(got)-600:])
	}
	if c := strings.Count(got, issueBodyCutMark); c != 1 {
		t.Errorf("want exactly one cut notice (the big section), got %d", c)
	}
	if !strings.Contains(got, ibLink) {
		t.Error("the cut notice must carry the full-text link")
	}
	// The kept part of the big section ends on a whole table row.
	cut := strings.Index(got, "\n\n> "+issueBodyCutMark)
	if cut < 0 || !strings.HasSuffix(got[:cut], "| x |\n") && !strings.HasSuffix(got[:cut], "| x |") {
		t.Errorf("the cut must land on a line boundary; text before the notice: %q", got[max(0, cut-80):max(0, cut)])
	}
}

// TestCapBodyClosesConstructs: a cut inside a fenced code block or an
// HTML comment would swallow the notice and everything after it (the
// checklist), so the cut closes the construct before the notice.
//
// regression: #4091
func TestCapBodyClosesConstructs(t *testing.T) {
	head := "<!-- verify-gate: x/01 -->\n\n"
	tail := "### Before you close\n"
	for name, open := range map[string]string{
		"backtick fence": "````text\n",
		"tilde fence":    "~~~\n",
		"html comment":   "<!--\n",
	} {
		t.Run(name, func(t *testing.T) {
			sec := ibFill("### Evidence\n\n"+strings.TrimSuffix(open, "\n"), 3000)
			got := capIssueBody(head, []string{sec}, tail, ibLinkFn)
			at := strings.Index(got, "> "+issueBodyCutMark)
			if at < 0 {
				t.Fatal("no cut stated")
			}
			before := got[:at]
			if strings.HasPrefix(open, "<!--") {
				if strings.LastIndex(before, "<!--") > strings.LastIndex(before, "-->") {
					t.Error("the notice sits inside an unclosed HTML comment")
				}
				return
			}
			fence := strings.TrimSuffix(open, "\n")
			if c := strings.Count(before, "\n"+fence[:3]); c%2 != 0 {
				t.Errorf("the notice sits inside an unclosed %s (%d fence lines before it)", name, c)
			}
			if !strings.Contains(before, "\n"+strings.TrimRight(fence, "text")+"\n") {
				t.Errorf("the closing fence must repeat the opening run %q", strings.TrimRight(fence, "text"))
			}
		})
	}
}

// TestCapBodyBackstop: a head that alone overruns the cap (an enormous
// rationale) is still cut under the cap, keeping the marker on line one.
//
// regression: #4091
func TestCapBodyBackstop(t *testing.T) {
	marker := "<!-- verify-gate: x/01 -->"
	head := marker + "\n\n> **Why:** " + strings.Repeat("long rationale 😀 ", 9000) + "\n\n"
	got := capIssueBody(head, []string{"### Verify\n\nrow\n\n"}, "### Before you close\n", ibLinkFn)
	if n := forgeBodyLen(got); n > issueBodyCap {
		t.Fatalf("capped body is %d, over the cap %d", n, issueBodyCap)
	}
	if !strings.HasPrefix(got, marker+"\n") {
		t.Error("the marker must stay the first line")
	}
	if !strings.Contains(got, issueBodyCutMark+" this issue body was cut") {
		t.Error("the backstop cut must be stated")
	}
}
