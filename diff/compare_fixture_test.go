package diff_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

// Fixture file constants for the diff package integration tests.
// A single old/new pair covers all currently tested features.
//
// When you add a new feature that's hard to test against this pair (e.g.
// it requires an isolated change), add a dedicated named pair instead of
// compounding changes onto fixtureNew.
const (
	fixtureOld = "testdata/WordPower60 Basic SS_2026-04-08.ce"
	fixtureNew = "testdata/WordPower60 Basic SS_Custom_2026-05-25.ce"
)

// compareFixtures is a helper that calls CompareFiles and fails the test on error.
func compareFixtures(t *testing.T, old, new string) diff.Diff {
	t.Helper()
	d, err := diff.CompareFiles(old, new)
	if err != nil {
		t.Fatalf("CompareFiles: %v", err)
	}
	return d
}

// findChangedPage returns the PageChange for the named page, or fails the test.
func findChangedPage(t *testing.T, d diff.Diff, name string) diff.PageChange {
	t.Helper()
	for _, ch := range d.ChangedPages {
		if ch.PageName == name {
			return ch
		}
	}
	t.Fatalf("page %q not found in ChangedPages (got: %v)", name, pageNames(d.ChangedPages))
	return diff.PageChange{}
}

func pageNames(pages []diff.PageChange) []string {
	names := make([]string, len(pages))
	for i, p := range pages {
		names[i] = p.PageName
	}
	return names
}

// ── Page-level changes (baseline → fixturePageChanges) ───────────────────────

func TestCompareFixtures_PageChanges_ExactAddedPages(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	want := []string{"Blank Page 1"}
	if !reflect.DeepEqual(d.AddedPages, want) {
		t.Errorf("AddedPages: got %v, want %v", d.AddedPages, want)
	}
}

func TestCompareFixtures_PageChanges_NoRemovedPages(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	if len(d.RemovedPages) != 0 {
		t.Errorf("RemovedPages: got %v, want none", d.RemovedPages)
	}
}

func TestCompareFixtures_PageChanges_ExactChangedPageNames(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	got := pageNames(d.ChangedPages) // already sorted by CompareFiles
	want := []string{"Activities", "Describe 2", "Groups", "My Scenes", "Places", "Social", "school"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ChangedPages names: got %v, want %v", got, want)
	}
}

// Activities: the "math" button gained a "navigate to page" action.
func TestCompareFixtures_PageChanges_Activities_MathGainedNavigateAction(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "Activities")

	if len(ch.Modified) != 1 {
		t.Fatalf("Activities modified count: got %d, want 1", len(ch.Modified))
	}
	if len(ch.Added)+len(ch.Removed) != 0 {
		t.Errorf("Activities should have no Added/Removed, got +%d -%d", len(ch.Added), len(ch.Removed))
	}

	mc := ch.Modified[0]
	if mc.Key.Label != "math" {
		t.Errorf("modified button label: got %q, want 'math'", mc.Key.Label)
	}
	if slices.Contains(mc.Before.Actions, "navigate to page") {
		t.Error("'navigate to page' should NOT be in Before.Actions — it was added in the new file")
	}
	if !slices.Contains(mc.After.Actions, "navigate to page") {
		t.Error("'navigate to page' should be in After.Actions")
	}
}

// Describe 2: the word "then" was added.
func TestCompareFixtures_PageChanges_Describe2_ThenButtonAdded(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "Describe 2")

	if len(ch.Added) != 1 {
		t.Fatalf("Describe 2 added count: got %d, want 1", len(ch.Added))
	}
	if len(ch.Removed)+len(ch.Modified) != 0 {
		t.Errorf("Describe 2 should have no Removed/Modified, got -%d ~%d", len(ch.Removed), len(ch.Modified))
	}

	btn := ch.Added[0]
	if btn.Label != "then" {
		t.Errorf("added button label: got %q, want 'then'", btn.Label)
	}
	if btn.Message != "then" {
		t.Errorf("added button message: got %q, want 'then'", btn.Message)
	}
}

// Groups: two buttons were hidden (visible true → false).
func TestCompareFixtures_PageChanges_Groups_TwoButtonsHidden(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "Groups")

	if len(ch.Modified) != 2 {
		t.Fatalf("Groups modified count: got %d, want 2", len(ch.Modified))
	}
	if len(ch.Added)+len(ch.Removed) != 0 {
		t.Errorf("Groups should have no Added/Removed, got +%d -%d", len(ch.Added), len(ch.Removed))
	}

	for _, mc := range ch.Modified {
		if mc.Before.Visible != true || mc.After.Visible != false {
			t.Errorf("Groups button %q: expected visible true→false, got %v→%v",
				mc.Key.Label, mc.Before.Visible, mc.After.Visible)
		}
	}
}

// My Scenes: "My Family" was removed.
func TestCompareFixtures_PageChanges_MyScenes_MyFamilyRemoved(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "My Scenes")

	if len(ch.Removed) != 1 {
		t.Fatalf("My Scenes removed count: got %d, want 1", len(ch.Removed))
	}
	if len(ch.Added)+len(ch.Modified) != 0 {
		t.Errorf("My Scenes should have no Added/Modified, got +%d ~%d", len(ch.Added), len(ch.Modified))
	}

	btn := ch.Removed[0]
	if btn.Label != "My Family" {
		t.Errorf("removed button label: got %q, want 'My Family'", btn.Label)
	}
}

// Places: "movie" was replaced — same label but message changed from "movie" to "movie theatre".
func TestCompareFixtures_PageChanges_Places_MovieMessageChanged(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "Places")

	if len(ch.Added) != 1 || len(ch.Removed) != 1 {
		t.Fatalf("Places: want +1 -1, got +%d -%d", len(ch.Added), len(ch.Removed))
	}

	added := ch.Added[0]
	if added.Label != "movie" || added.Message != "movie theatre" {
		t.Errorf("added button: got label=%q msg=%q, want label='movie' msg='movie theatre'",
			added.Label, added.Message)
	}

	removed := ch.Removed[0]
	if removed.Label != "movie" || removed.Message != "movie" {
		t.Errorf("removed button: got label=%q msg=%q, want label='movie' msg='movie'",
			removed.Label, removed.Message)
	}
}

// Social: "gentle, please" was added.
func TestCompareFixtures_PageChanges_Social_GentlePleaseAdded(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "Social")

	if len(ch.Added) != 1 {
		t.Fatalf("Social added count: got %d, want 1", len(ch.Added))
	}
	if len(ch.Removed)+len(ch.Modified) != 0 {
		t.Errorf("Social should have no Removed/Modified, got -%d ~%d", len(ch.Removed), len(ch.Modified))
	}

	btn := ch.Added[0]
	if btn.Label != "gentle, please" {
		t.Errorf("added button label: got %q, want 'gentle, please'", btn.Label)
	}
}

// school: two buttons added — "school rules" and "take picture".
func TestCompareFixtures_PageChanges_School_TwoButtonsAdded(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)
	ch := findChangedPage(t, d, "school")

	if len(ch.Added) != 2 {
		t.Fatalf("school added count: got %d, want 2", len(ch.Added))
	}
	if len(ch.Removed)+len(ch.Modified) != 0 {
		t.Errorf("school should have no Removed/Modified, got -%d ~%d", len(ch.Removed), len(ch.Modified))
	}

	labels := []string{ch.Added[0].Label, ch.Added[1].Label}
	wantLabels := []string{"school rules", "take picture"}
	if !reflect.DeepEqual(labels, wantLabels) {
		t.Errorf("school added labels: got %v, want %v", labels, wantLabels)
	}
}

// ── Sound (recording) changes — placeholder for plan-sound-changes.md ────────

// The "Home news" sound blob was re-recorded in fixtureNew. The current tool
// compares sound names, not blob content, so this change is invisible today.
// Once plan-sound-changes.md is implemented, remove the t.Skip and assert on
// Diff.SoundChanges.
func TestCompareFixtures_SoundChanges_HomeNewsRerecorded(t *testing.T) {
	t.Skip("not yet implemented: sound blob diffing tracked in plan-sound-changes.md")

	// Once plan-sound-changes.md is implemented, uncomment and verify:
	//
	// d := compareFixtures(t, fixtureOld, fixtureNew)
	//
	// if len(d.SoundChanges) != 1 {
	//     t.Fatalf("SoundChanges: got %d, want 1", len(d.SoundChanges))
	// }
	// sc := d.SoundChanges[0]
	// if sc.SoundName != "Home news" {
	//     t.Errorf("SoundName: got %q, want 'Home news'", sc.SoundName)
	// }
	// // "Play home news" appears on two pages; "Record home news" is excluded
	// // because plan-sound-changes.md tracks only code=4 (play) buttons.
	// if len(sc.Players) != 2 {
	//     t.Fatalf("Players: got %d, want 2 (Social 2 and News)", len(sc.Players))
	// }
	// pages := []string{sc.Players[0].PageName, sc.Players[1].PageName}
	// if !slices.Contains(pages, "Social 2") || !slices.Contains(pages, "News") {
	//     t.Errorf("Players pages: got %v, want [Social 2 News]", pages)
	// }
	// for _, p := range sc.Players {
	//     if p.ButtonLabel != "Play home news" {
	//         t.Errorf("player button label: got %q, want 'Play home news'", p.ButtonLabel)
	//     }
	// }
}

// ── Word-form modifier changes ────────────────────────────────────────────────

func TestCompareFixtures_WordFormChange_GoodPronunciation(t *testing.T) {
	d := compareFixtures(t, fixtureOld, fixtureNew)

	wfc := d.WordFormChanges
	if len(wfc.Modified) != 1 {
		t.Fatalf("WordFormChanges.Modified: got %d, want 1", len(wfc.Modified))
	}
	if len(wfc.Added)+len(wfc.Removed) != 0 {
		t.Errorf("expected no Added/Removed word-form changes, got +%d -%d", len(wfc.Added), len(wfc.Removed))
	}

	mc := wfc.Modified[0]
	if mc.ButtonSetName != "good" {
		t.Errorf("ButtonSetName: got %q, want 'good'", mc.ButtonSetName)
	}
	if mc.Key.FormIndex != 0 {
		t.Errorf("FormIndex: got %d, want 0 (base form)", mc.Key.FormIndex)
	}
	if mc.Before.Pronunciation != "" {
		t.Errorf("Before.Pronunciation: got %q, want empty", mc.Before.Pronunciation)
	}
	if mc.After.Pronunciation != "good job, kid" {
		t.Errorf("After.Pronunciation: got %q, want 'good job, kid'", mc.After.Pronunciation)
	}
}
