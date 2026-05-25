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

// Fixture IDs used below start at 10 (resources) and 20 (buttons) to avoid
// collisions with the page/button-box data inserted by newTestDB.

func TestLoadModifiers_EmptyDB(t *testing.T) {
	db := newTestDB(t) // no button_sets or button_set_modifiers rows
	mm, err := loadModifiers(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(mm) != 0 {
		t.Errorf("expected empty ModifierMap, got %d entries", len(mm))
	}
}

func TestLoadModifiers_Basic(t *testing.T) {
	db := newTestDB(t)
	// Word "good" as a button-set resource; resource type does not matter for
	// the modifier query — only resources.name (the word label) is selected.
	mustExec(t, db, `INSERT INTO resources VALUES (10,'{rid-good-bs}','good',5)`)
	mustExec(t, db, `INSERT INTO button_sets VALUES (1,10)`)
	// Modifier button: buttons.id=20 is what button_set_modifiers.button_id references.
	// buttons.resource_id=21 is what actionsQuery uses for action lookup.
	mustExec(t, db, `INSERT INTO buttons VALUES (20,21,'good','good job kid',1,'')`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (1,1,20,0)`)

	mm, err := loadModifiers(db)
	if err != nil {
		t.Fatal(err)
	}
	key := ModifierKey{ButtonSetName: "good", FormIndex: 0}
	btn, ok := mm[key]
	if !ok {
		t.Fatalf("expected key %+v in ModifierMap; got %d entries", key, len(mm))
	}
	if btn.Label != "good" {
		t.Errorf("Label: got %q, want 'good'", btn.Label)
	}
	if btn.Message != "good job kid" {
		t.Errorf("Message: got %q, want 'good job kid'", btn.Message)
	}
	if !btn.Visible {
		t.Error("Visible: expected true")
	}
}

func TestLoadModifiers_Pronunciation(t *testing.T) {
	db := newTestDB(t)
	mustExec(t, db, `INSERT INTO resources VALUES (10,'{rid-read-bs}','read',5)`)
	mustExec(t, db, `INSERT INTO button_sets VALUES (1,10)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (20,21,'read','',1,'reed')`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (1,1,20,0)`)

	mm, err := loadModifiers(db)
	if err != nil {
		t.Fatal(err)
	}
	btn := mm[ModifierKey{ButtonSetName: "read", FormIndex: 0}]
	if btn.Pronunciation != "reed" {
		t.Errorf("Pronunciation: got %q, want 'reed'", btn.Pronunciation)
	}
}

func TestLoadModifiers_MultipleFormsOfSameWord(t *testing.T) {
	db := newTestDB(t)
	mustExec(t, db, `INSERT INTO resources VALUES (10,'{rid-eat-bs}','eat',5)`)
	mustExec(t, db, `INSERT INTO button_sets VALUES (1,10)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (20,21,'eat','',1,'')`)
	mustExec(t, db, `INSERT INTO buttons VALUES (21,22,'eating','',1,'')`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (1,1,20,0)`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (2,1,21,6)`)

	mm, err := loadModifiers(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(mm) != 2 {
		t.Errorf("expected 2 entries, got %d", len(mm))
	}
	if _, ok := mm[ModifierKey{"eat", 0}]; !ok {
		t.Error("missing form 0 for 'eat'")
	}
	if btn, ok := mm[ModifierKey{"eat", 6}]; !ok {
		t.Error("missing form 6 for 'eat'")
	} else if btn.Label != "eating" {
		t.Errorf("form 6 label: got %q, want 'eating'", btn.Label)
	}
}

func TestLoadModifiers_MultipleDifferentWords(t *testing.T) {
	db := newTestDB(t)
	mustExec(t, db, `INSERT INTO resources VALUES (10,'{rid-go-bs}','go',5)`)
	mustExec(t, db, `INSERT INTO button_sets VALUES (1,10)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (20,21,'go','',1,'')`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (1,1,20,0)`)

	mustExec(t, db, `INSERT INTO resources VALUES (11,'{rid-eat-bs}','eat',5)`)
	mustExec(t, db, `INSERT INTO button_sets VALUES (2,11)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (21,22,'eat','',1,'')`)
	mustExec(t, db, `INSERT INTO button_set_modifiers VALUES (2,2,21,0)`)

	mm, err := loadModifiers(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(mm) != 2 {
		t.Errorf("expected 2 entries (one per word), got %d", len(mm))
	}
	if _, ok := mm[ModifierKey{"go", 0}]; !ok {
		t.Error("missing key {go, 0}")
	}
	if _, ok := mm[ModifierKey{"eat", 0}]; !ok {
		t.Error("missing key {eat, 0}")
	}
}
