package diff

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite" // registers the "sqlite" driver as a side-effect
)

// mustExec is a test helper that runs a SQL statement and fails the test on error.
func mustExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("mustExec: %v", err)
	}
}

// newTestDB creates an in-memory SQLite DB with the full schema and predictable
// fixture data. Closed automatically when the test ends via t.Cleanup.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mustExec(t, db, `
		CREATE TABLE resources (id INTEGER PRIMARY KEY, rid TEXT, name TEXT, type INTEGER);
		CREATE TABLE pages (id INTEGER PRIMARY KEY, resource_id INTEGER UNIQUE);
		CREATE TABLE buttons (id INTEGER PRIMARY KEY, resource_id INTEGER,
			label TEXT, message TEXT, visible INTEGER, pronunciation TEXT);
		CREATE TABLE button_boxes (id INTEGER PRIMARY KEY, resource_id INTEGER UNIQUE);
		CREATE TABLE button_box_instances (id INTEGER PRIMARY KEY,
			page_id INTEGER, button_box_id INTEGER);
		CREATE TABLE button_box_cells (id INTEGER PRIMARY KEY,
			button_box_id INTEGER, resource_id INTEGER,
			location INTEGER, span_x INTEGER, span_y INTEGER);
		CREATE TABLE actions (id INTEGER PRIMARY KEY,
			resource_id INTEGER, rank INTEGER, code INTEGER);
		CREATE TABLE action_data (id INTEGER PRIMARY KEY,
			action_id INTEGER, key INTEGER, value TEXT);
		CREATE TABLE button_sets (id INTEGER PRIMARY KEY, resource_id INTEGER);
		CREATE TABLE button_set_modifiers (id INTEGER PRIMARY KEY,
			button_set_id INTEGER, button_id INTEGER, modifier INTEGER);
	`)
	// Page: id=1, name="Home" (type 7 = page)
	mustExec(t, db, `INSERT INTO resources VALUES (1,'{rid-home}','Home',7)`)
	mustExec(t, db, `INSERT INTO pages VALUES (1,1)`)
	// ButtonBox linked to Home
	mustExec(t, db, `INSERT INTO resources VALUES (2,'{rid-bb}','bb',5)`)
	mustExec(t, db, `INSERT INTO button_boxes VALUES (1,2)`)
	mustExec(t, db, `INSERT INTO button_box_instances VALUES (1,1,1)`)
	// Button: resource_id=3, label="yes", message="yes", visible=1
	mustExec(t, db, `INSERT INTO resources VALUES (3,'{rid-yes}','yes',4)`)
	mustExec(t, db, `INSERT INTO button_box_cells VALUES (1,1,3,0,1,1)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (1,3,'yes','yes',1,'')`)
	return db
}

func TestLoadPages_ReturnsKnownPage(t *testing.T) {
	db := newTestDB(t)
	pages, err := loadPages(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := pages["Home"]; !ok {
		t.Errorf("expected page 'Home', got %v", pages)
	}
}

func TestLoadButtons_ButtonOnPage(t *testing.T) {
	db := newTestDB(t)
	bm, err := loadButtons(db)
	if err != nil {
		t.Fatal(err)
	}
	set, ok := bm["Home"]
	if !ok {
		t.Fatal("no buttons for page 'Home'")
	}
	if len(set) != 1 {
		t.Fatalf("expected 1 button, got %d", len(set))
	}
	for _, btn := range set {
		if btn.Label != "yes" {
			t.Errorf("label: got %q, want 'yes'", btn.Label)
		}
		if !btn.Visible {
			t.Error("expected Visible=true")
		}
	}
}

func TestLoadButtons_EmptyLabelAndMessageSkipped(t *testing.T) {
	db := newTestDB(t)
	// Insert a button with no label or message — should be filtered out.
	mustExec(t, db, `INSERT INTO resources VALUES (9,'{rid-empty}','',4)`)
	mustExec(t, db, `INSERT INTO button_box_cells VALUES (9,1,9,5,1,1)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (9,9,'','',1,'')`)
	bm, err := loadButtons(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(bm["Home"]) != 1 {
		t.Errorf("empty button should be filtered; got %d buttons", len(bm["Home"]))
	}
}
