package diff

import (
	"database/sql"
	"testing"
)

// newNavTestDB extends newTestDB with a second page ("Sports") and a
// code-9 navigate button ("go sports") on Home that points to it.
//
// Resource IDs continue from newTestDB (which uses 1–3):
//
//	10 = Sports page resource (type 7)
//	11 = nav button resource  (type 4)
//
// action id=5, action_data id=5 are free in the base fixture.
func newNavTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := newTestDB(t)

	mustExec(t, db, `INSERT INTO resources VALUES (10,'{rid-sports}','Sports',7)`)
	mustExec(t, db, `INSERT INTO pages VALUES (2,10)`)

	mustExec(t, db, `INSERT INTO resources VALUES (11,'{rid-nav}','go sports',4)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (2,11,'go sports','',1,'')`)
	// Attach the nav button to Home's existing button box (id=1).
	mustExec(t, db, `INSERT INTO button_box_cells VALUES (2,1,11,1,1,1)`)

	// code-9 action on resource_id=11; key 0 → Sports RID.
	mustExec(t, db, `INSERT INTO actions VALUES (5,11,0,9)`)
	mustExec(t, db, `INSERT INTO action_data VALUES (5,5,0,'{rid-sports}')`)

	return db
}

// ── loadPageNavGraph ──────────────────────────────────────────────────────────

func TestLoadPageNavGraph_HomeToSports(t *testing.T) {
	db := newNavTestDB(t)
	g, err := loadPageNavGraph(db)
	if err != nil {
		t.Fatal(err)
	}
	edges, ok := g["Home"]
	if !ok {
		t.Fatal("expected Home to have outgoing edges")
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge from Home, got %d: %v", len(edges), edges)
	}
	if edges[0].Label != "go sports" {
		t.Errorf("edge label: got %q, want %q", edges[0].Label, "go sports")
	}
	if edges[0].Dest != "Sports" {
		t.Errorf("edge dest: got %q, want %q", edges[0].Dest, "Sports")
	}
}

func TestLoadPageNavGraph_NoNavActionsReturnsEmptyGraph(t *testing.T) {
	db := newTestDB(t) // Home page + "yes" button with no code-9 action
	g, err := loadPageNavGraph(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(g) != 0 {
		t.Errorf("expected empty graph, got %v", g)
	}
}

func TestLoadPageNavGraph_DeduplicatesSameDestTwoButtons(t *testing.T) {
	db := newNavTestDB(t)
	// Add a second button on Home that also navigates to Sports.
	mustExec(t, db, `INSERT INTO resources VALUES (12,'{rid-nav2}','sports shortcut',4)`)
	mustExec(t, db, `INSERT INTO buttons VALUES (3,12,'sports shortcut','',1,'')`)
	mustExec(t, db, `INSERT INTO button_box_cells VALUES (3,1,12,2,1,1)`)
	mustExec(t, db, `INSERT INTO actions VALUES (6,12,0,9)`)
	mustExec(t, db, `INSERT INTO action_data VALUES (6,6,0,'{rid-sports}')`)

	g, err := loadPageNavGraph(db)
	if err != nil {
		t.Fatal(err)
	}
	edges := g["Home"]
	if len(edges) != 1 {
		t.Errorf("expected dedup to 1 edge, got %d: %v", len(edges), edges)
	}
	// "go sports" < "sports shortcut" lexicographically → first label wins.
	if edges[0].Label != "go sports" {
		t.Errorf("expected first label %q, got %q", "go sports", edges[0].Label)
	}
}

// ── AllShortestPaths ──────────────────────────────────────────────────────────

func TestAllShortestPaths_OneHop(t *testing.T) {
	// Button "Sports" on Home navigates to "sports-list".
	// Root (Home) is omitted; only non-root pages get breadcrumbs.
	g := PageNavGraph{"Home": {{Label: "Sports", Dest: "sports-list"}}}
	pages := PageSet{"Home": {}, "sports-list": {}}
	paths := AllShortestPaths(g, pages)
	want := "Home → Sports"
	if got := paths["sports-list"]; got != want {
		t.Errorf("sports-list path: got %q, want %q", got, want)
	}
}

func TestAllShortestPaths_RootPageOmitted(t *testing.T) {
	// Root page has no incoming edges — no breadcrumb for it.
	g := PageNavGraph{"Home": {{Label: "Sports", Dest: "sports-list"}}}
	pages := PageSet{"Home": {}, "sports-list": {}}
	paths := AllShortestPaths(g, pages)
	if _, ok := paths["Home"]; ok {
		t.Error("root page should be absent from paths map")
	}
}

func TestAllShortestPaths_TwoHops(t *testing.T) {
	g := PageNavGraph{
		"Home":        {{Label: "Sports", Dest: "sports-list"}},
		"sports-list": {{Label: "Indoor", Dest: "indoor-list"}},
	}
	pages := PageSet{"Home": {}, "sports-list": {}, "indoor-list": {}}
	paths := AllShortestPaths(g, pages)
	want := "Home → Sports → Indoor"
	if got := paths["indoor-list"]; got != want {
		t.Errorf("indoor-list path: got %q, want %q", got, want)
	}
}

func TestAllShortestPaths_ShortestPathWins(t *testing.T) {
	// Two equal-length routes to "target"; lex-first root edge ("A") wins.
	g := PageNavGraph{
		"Home": {
			{Label: "A", Dest: "via-a"},
			{Label: "B", Dest: "via-b"},
		},
		"via-a": {{Label: "Go", Dest: "target"}},
		"via-b": {{Label: "Go", Dest: "target"}},
	}
	pages := PageSet{"Home": {}, "via-a": {}, "via-b": {}, "target": {}}
	paths := AllShortestPaths(g, pages)
	want := "Home → A → Go"
	if got := paths["target"]; got != want {
		t.Errorf("target path: got %q, want %q", got, want)
	}
}

func TestAllShortestPaths_UnreachablePageOmitted(t *testing.T) {
	g := PageNavGraph{"Home": {}} // Sports has no incoming or outgoing edges
	pages := PageSet{"Home": {}, "Sports": {}}
	paths := AllShortestPaths(g, pages)
	if _, ok := paths["Sports"]; ok {
		t.Error("unreachable page should be absent from paths map")
	}
}

func TestAllShortestPaths_LexFirstLabelWhenTieDest(t *testing.T) {
	g := PageNavGraph{"Home": {
		{Label: "Aardvark", Dest: "sports-list"},
		{Label: "Zebra", Dest: "sports-list"},
	}}
	pages := PageSet{"Home": {}, "sports-list": {}}
	paths := AllShortestPaths(g, pages)
	want := "Home → Aardvark"
	if got := paths["sports-list"]; got != want {
		t.Errorf("tie-break path: got %q, want %q", got, want)
	}
}

func TestAllShortestPaths_MultipleRoots(t *testing.T) {
	// Two disconnected root → leaf pairs; both leaves get breadcrumbs.
	g := PageNavGraph{
		"RootA": {{Label: "Go A", Dest: "leaf-a"}},
		"RootB": {{Label: "Go B", Dest: "leaf-b"}},
	}
	pages := PageSet{"RootA": {}, "RootB": {}, "leaf-a": {}, "leaf-b": {}}
	paths := AllShortestPaths(g, pages)
	if got := paths["leaf-a"]; got != "RootA → Go A" {
		t.Errorf("leaf-a: got %q, want %q", got, "RootA → Go A")
	}
	if got := paths["leaf-b"]; got != "RootB → Go B" {
		t.Errorf("leaf-b: got %q, want %q", got, "RootB → Go B")
	}
	if _, ok := paths["RootA"]; ok {
		t.Error("RootA should be omitted (it is a root)")
	}
	if _, ok := paths["RootB"]; ok {
		t.Error("RootB should be omitted (it is a root)")
	}
}

func TestAllShortestPaths_EmptyGraphReturnsEmptyMap(t *testing.T) {
	// No edges at all — every page is a root, none get breadcrumbs.
	g := PageNavGraph{}
	pages := PageSet{"A": {}, "B": {}}
	paths := AllShortestPaths(g, pages)
	if len(paths) != 0 {
		t.Errorf("expected empty map, got %v", paths)
	}
}
