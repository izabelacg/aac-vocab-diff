package report

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

//go:embed templates/report.html
var reportTemplateFS embed.FS

// HTMLData is the root value passed to the HTML template.
type HTMLData struct {
	OldLabel  string
	NewLabel  string
	Generated string
	// BackURL, when non-empty, renders a "← New diff" link in the report header.
	// Set this to "/" when serving the report over HTTP so the user can return
	// to the upload form. Leave empty for file-based reports.
	BackURL string
	// ShortOldLabel / ShortNewLabel are OldLabel / NewLabel with the .ce suffix
	// stripped, used in the page <title> so browsers suggest a clean PDF filename.
	ShortOldLabel string
	ShortNewLabel string
	Stats         HTMLStats
	Sections      []HTMLSection
}

// HTMLStats holds the six summary counts shown in the summary box.
type HTMLStats struct {
	PagesAdded   int
	PagesRemoved int
	PagesChanged int
	BtnsAdded    int
	BtnsRemoved  int
	BtnsModified int
}

// HTMLSection groups a titled set of page cards (e.g. "Removed pages").
type HTMLSection struct {
	Title string
	Cards []HTMLCard
}

// HTMLCard represents one page-level card in the report.
type HTMLCard struct {
	Name       string
	Kind       string // "added" | "removed" | "changed"
	Rows       []HTMLRow
	PathSingle string // removed/added pages, or changed when old == new
	PathOld    string // changed page only, when old != new
	PathNew    string // always set alongside PathOld
}

// HTMLRow represents one button row inside a page card.
// Pre-rendered HTML fields use template.HTML so the template does not re-escape them.
type HTMLRow struct {
	Kind        string        // CSS class: "btn-added" | "btn-removed" | "btn-modified"
	BadgeClass  string        // "added" | "removed" | "modified"
	BadgeSymbol string        // "+", "−", "~"
	Label       string        // auto-escaped by html/template
	MessageHTML template.HTML // may contain <s>, <strong>, <span class='dim'>
	PronHTML    template.HTML
	VisibleHTML template.HTML
	ActionsHTML template.HTML
}

// WriteHTML renders the HTML report to w.
func WriteHTML(w io.Writer, data HTMLData) error {
	tmpl, err := template.ParseFS(reportTemplateFS, "templates/report.html")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(w, data)
}

// NewHTMLData converts a diff.Diff into the template data model.
func NewHTMLData(d diff.Diff) HTMLData {
	var btnsAdded, btnsRemoved, btnsModified int
	for _, ch := range d.ChangedPages {
		btnsAdded += len(ch.Added)
		btnsRemoved += len(ch.Removed)
		btnsModified += len(ch.Modified)
	}

	data := HTMLData{
		OldLabel:      d.OldLabel,
		NewLabel:      d.NewLabel,
		ShortOldLabel: strings.TrimSuffix(d.OldLabel, ".ce"),
		ShortNewLabel: strings.TrimSuffix(d.NewLabel, ".ce"),
		Generated:     time.Now().Format("2006-01-02 15:04"),
		Stats: HTMLStats{
			PagesAdded:   len(d.AddedPages),
			PagesRemoved: len(d.RemovedPages),
			PagesChanged: len(d.ChangedPages),
			BtnsAdded:    btnsAdded,
			BtnsRemoved:  btnsRemoved,
			BtnsModified: btnsModified,
		},
	}

	if len(d.RemovedPages) > 0 {
		sec := HTMLSection{Title: "Removed pages"}
		for _, p := range d.RemovedPages {
			rows := pageButtonRows(d.RemovedPageButtons[p], removedRow)
			card := HTMLCard{Name: p, Kind: "removed", Rows: rows}
			card.PathSingle = d.NavPathFromOld[p] // "" when unreachable — template shows nothing
			sec.Cards = append(sec.Cards, card)
		}
		data.Sections = append(data.Sections, sec)
	}

	if len(d.ChangedPages) > 0 {
		sec := HTMLSection{Title: "Changed pages"}
		for _, ch := range d.ChangedPages {
			card := HTMLCard{Name: ch.PageName, Kind: "changed"}
			for _, btn := range ch.Added {
				card.Rows = append(card.Rows, addedRow(btn))
			}
			for _, btn := range ch.Removed {
				card.Rows = append(card.Rows, removedRow(btn))
			}
			for _, bc := range ch.Modified {
				card.Rows = append(card.Rows, modifiedRow(bc))
			}
			oldP := d.NavPathFromOld[ch.PageName]
			newP := d.NavPathFromNew[ch.PageName]
			switch {
			case oldP != "" && newP != "" && oldP != newP:
				card.PathOld, card.PathNew = oldP, newP
			case oldP != "":
				card.PathSingle = oldP
			case newP != "":
				card.PathSingle = newP
				// default: both empty (Home absent or page unreachable) — leave PathSingle=""
			}
			sec.Cards = append(sec.Cards, card)
		}
		data.Sections = append(data.Sections, sec)
	}

	if len(d.AddedPages) > 0 {
		sec := HTMLSection{Title: "New pages"}
		for _, p := range d.AddedPages {
			rows := pageButtonRows(d.AddedPageButtons[p], addedRow)
			card := HTMLCard{Name: p, Kind: "added", Rows: rows}
			card.PathSingle = d.NavPathFromNew[p]
			sec.Cards = append(sec.Cards, card)
		}
		data.Sections = append(data.Sections, sec)
	}

	return data
}

// ── Row builders ─────────────────────────────────────────────────────────────

func addedRow(btn diff.Button) HTMLRow {
	return HTMLRow{
		Kind:        "btn-added",
		BadgeClass:  "added",
		BadgeSymbol: "+",
		Label:       btn.Label,
		MessageHTML: msgCell(btn.Label, btn.Message),
		PronHTML:    textOrDim(btn.Pronunciation),
		VisibleHTML: visCell(btn.Visible),
		ActionsHTML: actionsHTML(btn.Actions),
	}
}

func removedRow(btn diff.Button) HTMLRow {
	return HTMLRow{
		Kind:        "btn-removed",
		BadgeClass:  "removed",
		BadgeSymbol: "−",
		Label:       btn.Label,
		MessageHTML: msgCell(btn.Label, btn.Message),
		PronHTML:    textOrDim(btn.Pronunciation),
		VisibleHTML: visCell(btn.Visible),
		ActionsHTML: actionsHTML(btn.Actions),
	}
}

func modifiedRow(bc diff.ButtonChange) HTMLRow {
	oldMsg := bc.Before.Message
	if oldMsg == bc.Key.Label {
		oldMsg = ""
	}
	newMsg := bc.After.Message
	if newMsg == bc.Key.Label {
		newMsg = ""
	}

	return HTMLRow{
		Kind:        "btn-modified",
		BadgeClass:  "modified",
		BadgeSymbol: "~",
		Label:       bc.Key.Label,
		MessageHTML: diffField(oldMsg, newMsg),
		PronHTML:    diffField(bc.Before.Pronunciation, bc.After.Pronunciation),
		VisibleHTML: diffField(visStr2(bc.Before.Visible), visStr2(bc.After.Visible)),
		ActionsHTML: actionsDiffHTML(bc.Before.Actions, bc.After.Actions),
	}
}

// ── HTML cell helpers ─────────────────────────────────────────────────────────

func textOrDim(s string) template.HTML {
	if s == "" {
		return "<span class='dim'>—</span>"
	}
	return template.HTML(template.HTMLEscapeString(s))
}

func msgCell(label, message string) template.HTML {
	if message == "" || message == label {
		return "<span class='dim'>same as label</span>"
	}
	return template.HTML(template.HTMLEscapeString(message))
}

func visCell(visible bool) template.HTML {
	if visible {
		return "Yes"
	}
	return "<strong>No</strong>"
}

func visStr2(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// diffField renders a two-value cell: unchanged → plain text; changed → strikethrough + bold.
func diffField(oldVal, newVal string) template.HTML {
	if oldVal == newVal {
		if oldVal == "" {
			return "<span class='dim'>—</span>"
		}
		return template.HTML(template.HTMLEscapeString(oldVal))
	}
	oldHTML := template.HTMLEscapeString(oldVal)
	if oldVal == "" {
		oldHTML = "<span class='dim'>(none)</span>"
	}
	newHTML := template.HTMLEscapeString(newVal)
	if newVal == "" {
		newHTML = "<span class='dim'>(none)</span>"
	}
	return template.HTML(fmt.Sprintf(
		"<s style='color:#cf222e'>%s</s> → <strong style='color:#1a7f37'>%s</strong>",
		oldHTML, newHTML,
	))
}

func actionsHTML(actions []string) template.HTML {
	if len(actions) == 0 {
		return "<span class='dim'>—</span>"
	}
	var sb strings.Builder
	for _, a := range actions {
		sb.WriteString("<span class='action-pill'>")
		sb.WriteString(template.HTMLEscapeString(a))
		sb.WriteString("</span>")
	}
	return template.HTML(sb.String())
}

// pageButtonRows builds rows for a wholly added or removed page.
// rowFn is either addedRow or removedRow.
// Buttons are sorted by label for a deterministic, readable table.
// If the page had no buttons a single "No buttons" dim row is returned.
func pageButtonRows(bs diff.ButtonSet, rowFn func(diff.Button) HTMLRow) []HTMLRow {
	btns := sortedButtons(bs)
	if len(btns) == 0 {
		return []HTMLRow{{
			Kind: "btn-added", BadgeClass: "", BadgeSymbol: "",
			Label:       "",
			MessageHTML: "<td colspan='4'><span class='dim'>No buttons</span></td>",
		}}
	}
	rows := make([]HTMLRow, 0, len(btns))
	for _, btn := range btns {
		rows = append(rows, rowFn(btn))
	}
	return rows
}

// sortedButtons returns all buttons in a ButtonSet sorted by label.
func sortedButtons(bs diff.ButtonSet) []diff.Button {
	btns := make([]diff.Button, 0, len(bs))
	for _, btn := range bs {
		btns = append(btns, btn)
	}
	sort.Slice(btns, func(i, j int) bool { return btns[i].Label < btns[j].Label })
	return btns
}

func actionsDiffHTML(oldActs, newActs []string) template.HTML {
	if len(oldActs) == 0 && len(newActs) == 0 {
		return "<span class='dim'>—</span>"
	}
	oldSet := sliceToSet(oldActs)
	newSet := sliceToSet(newActs)

	// Preserve original order: old actions first, then newly-added ones.
	var sb strings.Builder
	for _, a := range oldActs {
		cls := ""
		if _, removed := newSet[a]; !removed {
			cls = " removed"
		}
		sb.WriteString(fmt.Sprintf("<span class='action-pill%s'>%s</span>",
			cls, template.HTMLEscapeString(a)))
	}
	// Newly added actions (not in old).
	added := make([]string, 0)
	for _, a := range newActs {
		if _, inOld := oldSet[a]; !inOld {
			added = append(added, a)
		}
	}
	sort.Strings(added)
	for _, a := range added {
		sb.WriteString(fmt.Sprintf("<span class='action-pill added'>%s</span>",
			template.HTMLEscapeString(a)))
	}
	if sb.Len() == 0 {
		return "<span class='dim'>—</span>"
	}
	return template.HTML(sb.String())
}
