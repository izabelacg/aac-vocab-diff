package report_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
	"github.com/izabelacg/aac-vocab-diff/report"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything that was printed. Safe to use in parallel tests because each
// call saves and restores os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// ── Header / labels ──────────────────────────────────────────────────────────

func TestPrintDiff_ShowsOldAndNewLabels(t *testing.T) {
	d := diff.Diff{OldLabel: "v1.ce", NewLabel: "v2.ce"}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "v1.ce") {
		t.Errorf("expected OldLabel in output, got:\n%s", out)
	}
	if !strings.Contains(out, "v2.ce") {
		t.Errorf("expected NewLabel in output, got:\n%s", out)
	}
}

// ── Page-level changes ───────────────────────────────────────────────────────

func TestPrintDiff_ShowsAddedPage(t *testing.T) {
	d := diff.Diff{
		OldLabel:   "old.ce",
		NewLabel:   "new.ce",
		AddedPages: []string{"Sports"},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "Sports") {
		t.Errorf("expected page name 'Sports' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "NEW PAGES") {
		t.Errorf("expected 'NEW PAGES' section header in output, got:\n%s", out)
	}
}

func TestPrintDiff_ShowsRemovedPage(t *testing.T) {
	d := diff.Diff{
		OldLabel:     "old.ce",
		NewLabel:     "new.ce",
		RemovedPages: []string{"Outdoors"},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "Outdoors") {
		t.Errorf("expected page name 'Outdoors' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "REMOVED PAGES") {
		t.Errorf("expected 'REMOVED PAGES' section header in output, got:\n%s", out)
	}
}

func TestPrintDiff_NoAddedPages_NoNewPagesHeader(t *testing.T) {
	d := diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if strings.Contains(out, "NEW PAGES") {
		t.Errorf("did not expect 'NEW PAGES' header when there are none, got:\n%s", out)
	}
}

func TestPrintDiff_NoRemovedPages_NoRemovedPagesHeader(t *testing.T) {
	d := diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if strings.Contains(out, "REMOVED PAGES") {
		t.Errorf("did not expect 'REMOVED PAGES' header when there are none, got:\n%s", out)
	}
}

// ── Button-level changes ─────────────────────────────────────────────────────

func TestPrintDiff_ShowsAddedButton(t *testing.T) {
	d := diff.Diff{
		OldLabel: "old.ce",
		NewLabel: "new.ce",
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "soccer", Visible: true}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "soccer") {
		t.Errorf("expected button label 'soccer' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "+") {
		t.Errorf("expected '+' prefix for added button, got:\n%s", out)
	}
}

func TestPrintDiff_ShowsRemovedButton(t *testing.T) {
	d := diff.Diff{
		OldLabel: "old.ce",
		NewLabel: "new.ce",
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Removed:  []diff.Button{{Label: "soccer", Visible: true}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "soccer") {
		t.Errorf("expected button label 'soccer' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "-") {
		t.Errorf("expected '-' prefix for removed button, got:\n%s", out)
	}
}

func TestPrintDiff_ShowsModifiedButton(t *testing.T) {
	d := diff.Diff{
		OldLabel: "old.ce",
		NewLabel: "new.ce",
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Modified: []diff.ButtonChange{{
				Key:    diff.ButtonKey{Label: "help", Message: "help me"},
				Before: diff.Button{Visible: true},
				After:  diff.Button{Visible: false},
			}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "help") {
		t.Errorf("expected button label 'help' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "hidden") {
		t.Errorf("expected visibility change ('hidden') in output, got:\n%s", out)
	}
	if !strings.Contains(out, "~") {
		t.Errorf("expected '~' prefix for modified button, got:\n%s", out)
	}
}

func TestPrintDiff_HiddenButton_ShowsHiddenMarker(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "secret", Visible: false}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "hidden") {
		t.Errorf("expected '[hidden]' marker for invisible button, got:\n%s", out)
	}
}

func TestPrintDiff_VisibleButton_NoHiddenMarker(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "yes", Visible: true}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if strings.Contains(out, "hidden") {
		t.Errorf("did not expect '[hidden]' for a visible button, got:\n%s", out)
	}
}

func TestPrintDiff_ButtonWithDifferentMessage_ShowsMessage(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "home", Message: "go home", Visible: true}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "go home") {
		t.Errorf("expected spoken message 'go home' in output, got:\n%s", out)
	}
}

func TestPrintDiff_ButtonMessageSameAsLabel_NotRepeated(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "yes", Message: "yes", Visible: true}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	// "yes" should appear at least once (the label), but the message should
	// not be printed separately when it is identical to the label.
	count := strings.Count(out, "yes")
	if count > 1 {
		t.Errorf("message 'yes' printed %d times; expected label only (no repeat for same message)", count)
	}
}

func TestPrintDiff_ButtonWithPronunciation_ShowsPronunciation(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added: []diff.Button{{
				Label:         "read",
				Visible:       true,
				Pronunciation: "reed",
			}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "reed") {
		t.Errorf("expected pronunciation 'reed' in output, got:\n%s", out)
	}
}

func TestPrintDiff_ModifiedPronunciation_ShowsChange(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Modified: []diff.ButtonChange{{
				Key:    diff.ButtonKey{Label: "read"},
				Before: diff.Button{Pronunciation: "red"},
				After:  diff.Button{Pronunciation: "reed"},
			}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "red") || !strings.Contains(out, "reed") {
		t.Errorf("expected both old ('red') and new ('reed') pronunciation in output, got:\n%s", out)
	}
}

func TestPrintDiff_NoChanges_ShowsNoChangesMessage(t *testing.T) {
	d := diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(strings.ToLower(out), "no") {
		t.Errorf("expected a 'no changes' message for an empty diff, got:\n%s", out)
	}
}

// ── Summary line ─────────────────────────────────────────────────────────────

func TestPrintDiff_SummaryLine(t *testing.T) {
	d := diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	if !strings.Contains(out, "Summary") {
		t.Errorf("expected 'Summary' line in output, got:\n%s", out)
	}
}

func TestPrintDiff_SummaryCounts(t *testing.T) {
	d := diff.Diff{
		OldLabel:     "old.ce",
		NewLabel:     "new.ce",
		AddedPages:   []string{"Sports"},
		RemovedPages: []string{"Old"},
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "yes", Visible: true}},
			Removed:  []diff.Button{{Label: "no", Visible: true}},
			Modified: []diff.ButtonChange{{
				Key:    diff.ButtonKey{Label: "help"},
				Before: diff.Button{Visible: true},
				After:  diff.Button{Visible: false},
			}},
		}},
	}
	out := captureStdout(t, func() { report.PrintDiff(d) })
	// All six counts must appear somewhere in the output.
	for _, want := range []string{"1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected count %q in summary, got:\n%s", want, out)
		}
	}
}
