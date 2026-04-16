package diff

import (
	"database/sql"
	"fmt"
	"sort"

	_ "modernc.org/sqlite"
)

const pageQuery = "SELECT name FROM resources WHERE type = 7 ORDER BY name"

const buttonQuery = `
SELECT
    r_page.name     AS page,
    b.resource_id,
    b.label         AS label,
    b.message       AS message,
    b.visible       AS visible,
    b.pronunciation AS pronunciation
FROM resources r_page
JOIN pages p                  ON p.resource_id     = r_page.id
JOIN button_box_instances bbi ON bbi.page_id       = p.id
JOIN button_boxes bb          ON bb.id             = bbi.button_box_id
JOIN button_box_cells bbc     ON bbc.button_box_id = bb.id
JOIN buttons b                ON b.resource_id     = bbc.resource_id
ORDER BY r_page.name, b.label;
`

const actionsQuery = `
SELECT a.resource_id, a.rank, a.code, ad.key, ad.value, r_target.name
FROM actions a
LEFT JOIN action_data ad     ON ad.action_id = a.id
LEFT JOIN resources r_target ON r_target.rid = ad.value
ORDER BY a.resource_id, a.rank, ad.key
`

var actionLabels = map[int]string{
	3:  "speak",
	4:  "play sound",
	5:  "record",
	6:  "navigate back",
	8:  "show image",
	9:  "navigate to page",
	10: "speak (auto)",
	16: "punctuation",
	23: "toggle mute",
	24: "save message",
	28: "clear message",
	40: "volume up",
	41: "volume down",
	42: "speak (repeat)",
	43: "delete word",
	44: "battery status",
	45: "time & date",
	60: "backspace",
	63: "stop",
	64: "play video",
	65: "text to speech",
	66: "copy text",
	67: "paste text",
	68: "share text",
	70: "open app",
	71: "word form",
	73: "pronoun",
	74: "find word",
	77: "clear all",
	82: "send to display",
}

// actionRow holds one row from actionsQuery.
// sql.NullInt64 / sql.NullString handle nullable columns without panicking.
type actionRow struct {
	resourceID int
	rank       int
	code       int
	key        sql.NullInt64
	value      sql.NullString
	targetName sql.NullString
}

// LoadPages opens the .c4v SQLite file and returns the set of page names.
func LoadPages(dbPath string) (PageSet, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", dbPath, err)
	}
	defer db.Close()
	return loadPages(db)
}

// LoadButtons opens the .c4v SQLite file and returns all buttons grouped by page.
func LoadButtons(dbPath string) (ButtonMap, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", dbPath, err)
	}
	defer db.Close()
	return loadButtons(db)
}

// loadPages queries the DB for all page names (type 7 resources).
func loadPages(db *sql.DB) (PageSet, error) {
	rows, err := db.Query(pageQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := PageSet{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		pages[name] = struct{}{}
	}
	return pages, rows.Err()
}

// loadButtons queries the DB and builds a ButtonMap keyed by page name.
func loadButtons(db *sql.DB) (ButtonMap, error) {
	// 1. Fetch all action rows first, grouped by resource_id.
	//    We do this upfront so the button loop can look them up in O(1).
	aRows, err := db.Query(actionsQuery)
	if err != nil {
		return nil, err
	}
	defer aRows.Close()

	actionsByRID := map[int][]actionRow{}
	for aRows.Next() {
		var ar actionRow
		if err := aRows.Scan(&ar.resourceID, &ar.rank, &ar.code,
			&ar.key, &ar.value, &ar.targetName); err != nil {
			return nil, err
		}
		actionsByRID[ar.resourceID] = append(actionsByRID[ar.resourceID], ar)
	}
	if err := aRows.Err(); err != nil {
		return nil, err
	}

	// 2. Fetch button rows and build the ButtonMap.
	bRows, err := db.Query(buttonQuery)
	if err != nil {
		return nil, err
	}
	defer bRows.Close()

	bm := ButtonMap{}
	for bRows.Next() {
		var (
			page          string
			rid           int
			label, msg    sql.NullString // nullable in the schema
			visible       int
			pronunciation sql.NullString
		)
		if err := bRows.Scan(&page, &rid, &label, &msg, &visible, &pronunciation); err != nil {
			return nil, err
		}

		// Skip empty placeholder cells
		if label.String == "" && msg.String == "" {
			continue
		}

		btn := Button{
			Label:         label.String,
			Message:       msg.String,
			Visible:       visible == 1, // SQLite stores booleans as 0/1
			Pronunciation: pronunciation.String,
			Actions:       buildActionSummary(actionsByRID[rid]),
		}
		if bm[page] == nil {
			bm[page] = ButtonSet{}
		}
		bm[page][btn.Fingerprint()] = btn
	}
	return bm, bRows.Err()
}

func navigateDestination(targetName, value, vocab string) string {
	dest := targetName
	if dest == "" {
		dest = value
	}
	if dest == "" {
		dest = "?"
	}
	if vocab != "" {
		dest = vocab + "/" + dest // cross-vocab: "VocabName/PageName"
	}
	return dest
}

// buildActionSummary collapses raw SQL rows for one button into human-readable action strings.
func buildActionSummary(rows []actionRow) []string {
	// Group by rank — each rank is one action; multiple rows = multiple params.
	type rankEntry struct {
		code   int
		params map[int64]actionRow // key index → row
	}
	byRank := map[int]*rankEntry{}
	for _, ar := range rows {
		if _, ok := byRank[ar.rank]; !ok {
			byRank[ar.rank] = &rankEntry{code: ar.code, params: map[int64]actionRow{}}
		}
		if ar.key.Valid {
			byRank[ar.rank].params[ar.key.Int64] = ar
		}
	}

	ranks := make([]int, 0, len(byRank))
	for r := range byRank {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)

	var parts []string
	for _, rank := range ranks {
		e := byRank[rank]
		lbl, ok := actionLabels[e.code]
		if !ok {
			lbl = fmt.Sprintf("action#%d", e.code)
		}

		switch e.code {
		case 9: // navigate to page — key 0 = target RID, key 1 = vocab name
			p0 := e.params[0]
			dest := navigateDestination(p0.targetName.String, p0.value.String, e.params[1].value.String)
			parts = append(parts, fmt.Sprintf("%s: %q", lbl, dest))
		case 4: // play sound — key 0 = sound resource
			s := e.params[0].targetName.String
			if s == "" {
				s = e.params[0].value.String
			}
			if s == "" {
				s = "?"
			}
			parts = append(parts, fmt.Sprintf("%s: %q", lbl, s))
		case 70: // open app — key 0 = bundle ID
			app := e.params[0].value.String
			if app == "" {
				app = "?"
			}
			parts = append(parts, fmt.Sprintf("%s: %s", lbl, app))
		case 71: // word form — key 0 = form index
			form := e.params[0].value.String
			if form == "" {
				form = "?"
			}
			parts = append(parts, fmt.Sprintf("%s: %s", lbl, form))
		default:
			parts = append(parts, lbl)
		}
	}
	return parts
}
