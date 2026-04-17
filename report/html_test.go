package report_test

import (
	"strings"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
	"github.com/izabelacg/aac-vocab-diff/report"
)

func TestWriteHTML_NoError(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"})
	var sb strings.Builder
	if err := report.WriteHTML(&sb, data); err != nil {
		t.Fatalf("WriteHTML returned error: %v", err)
	}
}

func TestWriteHTML_ContainsLabels(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "old.ce", NewLabel: "new.ce"})
	var sb strings.Builder
	if err := report.WriteHTML(&sb, data); err != nil {
		t.Fatal(err)
	}
	out := sb.String()
	if !strings.Contains(out, "old.ce") {
		t.Errorf("OldLabel missing from HTML output")
	}
	if !strings.Contains(out, "new.ce") {
		t.Errorf("NewLabel missing from HTML output")
	}
}

func TestWriteHTML_ContainsSummaryBox(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "a.ce", NewLabel: "b.ce"})
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	if !strings.Contains(sb.String(), "summary-box") {
		t.Error("expected summary-box in HTML output")
	}
}

func TestWriteHTML_AddedPageSection(t *testing.T) {
	d := diff.Diff{
		OldLabel:   "old.ce",
		NewLabel:   "new.ce",
		AddedPages: []string{"Sports"},
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	if !strings.Contains(out, "Sports") {
		t.Error("expected added page 'Sports' in HTML output")
	}
	if !strings.Contains(out, "New pages") {
		t.Error("expected 'New pages' section title in HTML output")
	}
}

func TestWriteHTML_RemovedPageSection(t *testing.T) {
	d := diff.Diff{
		OldLabel:     "old.ce",
		NewLabel:     "new.ce",
		RemovedPages: []string{"Outdoors"},
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	if !strings.Contains(out, "Outdoors") {
		t.Error("expected removed page 'Outdoors' in HTML output")
	}
	if !strings.Contains(out, "Removed pages") {
		t.Error("expected 'Removed pages' section title in HTML output")
	}
}

func TestWriteHTML_ModifiedButtonRow(t *testing.T) {
	d := diff.Diff{
		OldLabel: "old.ce",
		NewLabel: "new.ce",
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Modified: []diff.ButtonChange{{
				Key:    diff.ButtonKey{Label: "help"},
				Before: diff.Button{Visible: true},
				After:  diff.Button{Visible: false},
			}},
		}},
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	if !strings.Contains(out, "help") {
		t.Error("expected button label 'help' in HTML output")
	}
	if !strings.Contains(out, "btn-modified") {
		t.Error("expected 'btn-modified' CSS class in HTML output")
	}
}

func TestWriteHTML_SummaryStats(t *testing.T) {
	d := diff.Diff{
		OldLabel:     "old.ce",
		NewLabel:     "new.ce",
		AddedPages:   []string{"Sports"},
		RemovedPages: []string{"Old"},
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "yes", Visible: true}},
		}},
	}
	data := report.NewHTMLData(d)
	if data.Stats.PagesAdded != 1 {
		t.Errorf("PagesAdded: got %d, want 1", data.Stats.PagesAdded)
	}
	if data.Stats.PagesRemoved != 1 {
		t.Errorf("PagesRemoved: got %d, want 1", data.Stats.PagesRemoved)
	}
	if data.Stats.PagesChanged != 1 {
		t.Errorf("PagesChanged: got %d, want 1", data.Stats.PagesChanged)
	}
	if data.Stats.BtnsAdded != 1 {
		t.Errorf("BtnsAdded: got %d, want 1", data.Stats.BtnsAdded)
	}
}

func TestWriteHTML_AddedButtonRow(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "soccer", Message: "kick!", Visible: true}},
		}},
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	if !strings.Contains(out, "soccer") {
		t.Error("expected button label 'soccer' in HTML output")
	}
	if !strings.Contains(out, "kick!") {
		t.Error("expected spoken message 'kick!' in HTML output")
	}
	if !strings.Contains(out, "btn-added") {
		t.Error("expected 'btn-added' CSS class in HTML output")
	}
}

func TestWriteHTML_HiddenButtonShowsNo(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{
			PageName: "Home",
			Added:    []diff.Button{{Label: "secret", Visible: false}},
		}},
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	if !strings.Contains(sb.String(), "No") {
		t.Error("expected 'No' for hidden button visible cell")
	}
}

func TestNewHTMLData_ShortLabelsStripExtension(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "old.ce", NewLabel: "new.ce"})
	if data.ShortOldLabel != "old" {
		t.Errorf("ShortOldLabel: got %q, want %q", data.ShortOldLabel, "old")
	}
	if data.ShortNewLabel != "new" {
		t.Errorf("ShortNewLabel: got %q, want %q", data.ShortNewLabel, "new")
	}
}

func TestWriteHTML_TitleUsesShortLabels(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "old.ce", NewLabel: "new.ce"})
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()

	titleStart := strings.Index(out, "<title>")
	titleEnd := strings.Index(out, "</title>")
	if titleStart < 0 || titleEnd < 0 {
		t.Fatal("could not find <title> tag in output")
	}
	title := out[titleStart : titleEnd+len("</title>")]
	if strings.Contains(title, ".ce") {
		t.Errorf("page title contains .ce extension, browsers would use it in the PDF filename: %q", title)
	}
	if !strings.Contains(title, "old") || !strings.Contains(title, "new") {
		t.Errorf("page title missing label text, got: %q", title)
	}
}

func TestWriteHTML_ContainsPrintButton(t *testing.T) {
	data := report.NewHTMLData(diff.Diff{OldLabel: "old.ce", NewLabel: "new.ce"})
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	if !strings.Contains(out, "window.print()") {
		t.Error("report missing Save as PDF button (window.print())")
	}
	if !strings.Contains(out, "Save as PDF") {
		t.Error("report missing 'Save as PDF' button label")
	}
}

func TestNewHTMLData_EscapesSpecialChars(t *testing.T) {
	d := diff.Diff{
		OldLabel: "a & b.ce",
		NewLabel: "c > d.ce",
	}
	data := report.NewHTMLData(d)
	var sb strings.Builder
	_ = report.WriteHTML(&sb, data)
	out := sb.String()
	// html/template should auto-escape & and > in string fields.
	if strings.Contains(out, "a & b.ce") {
		t.Error("raw '&' should be HTML-escaped in output")
	}
	if !strings.Contains(out, "a &amp; b.ce") {
		t.Error("expected HTML-escaped '&amp;' for label with ampersand")
	}
}

// ── Nav path: NewHTMLData model ───────────────────────────────────────────────

func TestNewHTMLData_RemovedPageFillsPathSingle(t *testing.T) {
	d := diff.Diff{
		RemovedPages:   []string{"Sports"},
		NavPathFromOld: map[string]string{"Sports": "Home → Sports"},
	}
	data := report.NewHTMLData(d)
	card := data.Sections[0].Cards[0]
	if card.PathSingle != "Home → Sports" {
		t.Errorf("PathSingle: got %q, want %q", card.PathSingle, "Home → Sports")
	}
	if card.PathOld != "" || card.PathNew != "" {
		t.Errorf("PathOld/PathNew should be empty for removed page, got %q/%q", card.PathOld, card.PathNew)
	}
}

func TestNewHTMLData_AddedPageFillsPathSingle(t *testing.T) {
	d := diff.Diff{
		AddedPages:     []string{"Sports"},
		NavPathFromNew: map[string]string{"Sports": "Home → Sports"},
	}
	data := report.NewHTMLData(d)
	card := data.Sections[0].Cards[0]
	if card.PathSingle != "Home → Sports" {
		t.Errorf("PathSingle: got %q, want %q", card.PathSingle, "Home → Sports")
	}
}

func TestNewHTMLData_ChangedPageSamePath_SingleLine(t *testing.T) {
	d := diff.Diff{
		ChangedPages:   []diff.PageChange{{PageName: "Sports"}},
		NavPathFromOld: map[string]string{"Sports": "Home → Sports"},
		NavPathFromNew: map[string]string{"Sports": "Home → Sports"},
	}
	data := report.NewHTMLData(d)
	card := data.Sections[0].Cards[0]
	if card.PathSingle != "Home → Sports" {
		t.Errorf("PathSingle: got %q, want %q", card.PathSingle, "Home → Sports")
	}
	if card.PathOld != "" || card.PathNew != "" {
		t.Errorf("PathOld/PathNew should be empty when paths are identical, got %q/%q", card.PathOld, card.PathNew)
	}
}

func TestNewHTMLData_ChangedPageDiffPaths_SplitLines(t *testing.T) {
	d := diff.Diff{
		ChangedPages:   []diff.PageChange{{PageName: "Sports"}},
		NavPathFromOld: map[string]string{"Sports": "Home → Old Cat → Sports"},
		NavPathFromNew: map[string]string{"Sports": "Home → New Cat → Sports"},
	}
	data := report.NewHTMLData(d)
	card := data.Sections[0].Cards[0]
	if card.PathOld != "Home → Old Cat → Sports" {
		t.Errorf("PathOld: got %q", card.PathOld)
	}
	if card.PathNew != "Home → New Cat → Sports" {
		t.Errorf("PathNew: got %q", card.PathNew)
	}
	if card.PathSingle != "" {
		t.Errorf("PathSingle should be empty for split paths, got %q", card.PathSingle)
	}
}

func TestNewHTMLData_NoNavPath_FieldsEmpty(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{PageName: "Sports"}},
		// no NavPathFromOld / NavPathFromNew
	}
	data := report.NewHTMLData(d)
	card := data.Sections[0].Cards[0]
	if card.PathSingle != "" || card.PathOld != "" || card.PathNew != "" {
		t.Errorf("all path fields should be empty when nav maps are nil, got single=%q old=%q new=%q",
			card.PathSingle, card.PathOld, card.PathNew)
	}
}

// ── Nav path: WriteHTML rendering ────────────────────────────────────────────

func TestWriteHTML_SingleNavPath(t *testing.T) {
	d := diff.Diff{
		OldLabel:       "old.ce",
		NewLabel:       "new.ce",
		RemovedPages:   []string{"Sports"},
		NavPathFromOld: map[string]string{"Sports": "Home → Sports"},
	}
	var sb strings.Builder
	_ = report.WriteHTML(&sb, report.NewHTMLData(d))
	out := sb.String()
	if !strings.Contains(out, "Home → Sports") {
		t.Error("expected breadcrumb 'Home → Sports' in HTML output")
	}
	if !strings.Contains(out, "page-path") {
		t.Error("expected 'page-path' CSS class in HTML output")
	}
}

func TestWriteHTML_SplitNavPath(t *testing.T) {
	d := diff.Diff{
		OldLabel:       "old.ce",
		NewLabel:       "new.ce",
		ChangedPages:   []diff.PageChange{{PageName: "Sports"}},
		NavPathFromOld: map[string]string{"Sports": "Home → Old Cat → Sports"},
		NavPathFromNew: map[string]string{"Sports": "Home → New Cat → Sports"},
	}
	var sb strings.Builder
	_ = report.WriteHTML(&sb, report.NewHTMLData(d))
	out := sb.String()
	if !strings.Contains(out, "path-label") {
		t.Error("expected 'path-label' class for Old:/New: split")
	}
	if !strings.Contains(out, "Old Cat") || !strings.Contains(out, "New Cat") {
		t.Error("expected both old and new path text in HTML output")
	}
}

func TestWriteHTML_NoNavPath_NoPagePathClass(t *testing.T) {
	d := diff.Diff{
		ChangedPages: []diff.PageChange{{PageName: "Sports"}},
		// no nav paths — page-path div should not appear
	}
	var sb strings.Builder
	_ = report.WriteHTML(&sb, report.NewHTMLData(d))
	if strings.Contains(sb.String(), `class="page-path"`) {
		t.Error("page-path element should not appear when no nav path is set")
	}
}
