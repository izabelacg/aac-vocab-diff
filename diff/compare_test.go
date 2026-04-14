package diff_test

import (
	"reflect"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

// makeSet builds a ButtonSet from a variadic list of Buttons.
// This mirrors what load.go does when reading from SQLite: each button is
// keyed by Fingerprint(). Two buttons are the same set entry iff their
// fingerprints match.
func makeSet(buttons ...diff.Button) diff.ButtonSet {
	s := diff.ButtonSet{}
	for _, b := range buttons {
		s[b.Fingerprint()] = b
	}
	return s
}

func TestComputeDiff_AddedPage(t *testing.T) {
	oldPages := diff.PageSet{"Home": {}}
	newPages := diff.PageSet{"Home": {}, "Sports": {}}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet()},
		diff.ButtonMap{"Home": makeSet(), "Sports": makeSet()},
		oldPages, newPages,
	)
	if !reflect.DeepEqual(result.AddedPages, []string{"Sports"}) {
		t.Errorf("AddedPages: got %v, want [Sports]", result.AddedPages)
	}
}

func TestComputeDiff_RemovedPage(t *testing.T) {
	oldPages := diff.PageSet{"Home": {}, "Sports": {}}
	newPages := diff.PageSet{"Home": {}}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(), "Sports": makeSet()},
		diff.ButtonMap{"Home": makeSet()},
		oldPages, newPages,
	)
	if !reflect.DeepEqual(result.RemovedPages, []string{"Sports"}) {
		t.Errorf("RemovedPages: got %v, want [Sports]", result.RemovedPages)
	}
}

func TestComputeDiff_AddedPagesSorted(t *testing.T) {
	newPages := diff.PageSet{"A": {}, "Home": {}, "Z": {}}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(), "A": makeSet(), "Z": makeSet()},
		diff.ButtonMap{"Home": makeSet(), "A": makeSet(), "Z": makeSet()},
		diff.PageSet{"Home": {}},
		newPages,
	)
	want := []string{"A", "Z"}
	if !reflect.DeepEqual(result.AddedPages, want) {
		t.Errorf("AddedPages: got %v, want %v", result.AddedPages, want)
	}
}

func TestComputeDiff_AddedButton(t *testing.T) {
	yes := diff.Button{Label: "yes", Visible: true}
	no := diff.Button{Label: "no", Visible: true}

	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(yes)},
		diff.ButtonMap{"Home": makeSet(yes, no)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if ch.PageName != "Home" {
		t.Errorf("PageName: got %q", ch.PageName)
	}
	if len(ch.Added) != 1 || ch.Added[0].Label != "no" {
		t.Errorf("Added: got %+v", ch.Added)
	}
}

func TestComputeDiff_RemovedButton(t *testing.T) {
	yes := diff.Button{Label: "yes", Visible: true}
	no := diff.Button{Label: "no", Visible: true}

	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(yes, no)},
		diff.ButtonMap{"Home": makeSet(yes)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if len(ch.Removed) != 1 || ch.Removed[0].Label != "no" {
		t.Errorf("Removed: got %+v", ch.Removed)
	}
}

func TestComputeDiff_ModifiedButton(t *testing.T) {
	before := diff.Button{Label: "help", Message: "help me", Visible: true}
	after := diff.Button{Label: "help", Message: "help me", Visible: false}

	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(before)},
		diff.ButtonMap{"Home": makeSet(after)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if len(ch.Modified) != 1 {
		t.Fatalf("expected 1 modified, got %+v", ch)
	}
	mc := ch.Modified[0]
	if mc.Before.Visible != true || mc.After.Visible != false {
		t.Errorf("Modified Visible: before=%v after=%v", mc.Before.Visible, mc.After.Visible)
	}
	if len(ch.Added) != 0 || len(ch.Removed) != 0 {
		t.Error("modified button should not appear in Added or Removed")
	}
}

func TestComputeDiff_NoChanges(t *testing.T) {
	btn := diff.Button{Label: "yes", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(btn)},
		diff.ButtonMap{"Home": makeSet(btn)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 0 || len(result.AddedPages) != 0 || len(result.RemovedPages) != 0 {
		t.Errorf("expected no changes, got %+v", result)
	}
}

// Whole-page additions are listed in AddedPages only; they must not produce a
// PageChange row (those buttons are not "changes on an existing page").
func TestComputeDiff_NewPageNotInChangedPages(t *testing.T) {
	yes := diff.Button{Label: "yes", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(yes)},
		diff.ButtonMap{
			"Home":   makeSet(yes),
			"Sports": makeSet(diff.Button{Label: "soccer", Visible: true}),
		},
		diff.PageSet{"Home": {}},
		diff.PageSet{"Home": {}, "Sports": {}},
	)
	if !reflect.DeepEqual(result.AddedPages, []string{"Sports"}) {
		t.Fatalf("AddedPages: got %v", result.AddedPages)
	}
	for _, ch := range result.ChangedPages {
		if ch.PageName == "Sports" {
			t.Errorf("did not expect Sports in ChangedPages, got %+v", ch)
		}
	}
}

func TestComputeDiff_RemovedPageNotInChangedPages(t *testing.T) {
	yes := diff.Button{Label: "yes", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{
			"Home":   makeSet(yes),
			"Sports": makeSet(diff.Button{Label: "soccer", Visible: true}),
		},
		diff.ButtonMap{"Home": makeSet(yes)},
		diff.PageSet{"Home": {}, "Sports": {}},
		diff.PageSet{"Home": {}},
	)
	if !reflect.DeepEqual(result.RemovedPages, []string{"Sports"}) {
		t.Fatalf("RemovedPages: got %v", result.RemovedPages)
	}
	for _, ch := range result.ChangedPages {
		if ch.PageName == "Sports" {
			t.Errorf("did not expect Sports in ChangedPages, got %+v", ch)
		}
	}
}

// When more than one button shares the same ButtonKey on the removed side
// and/or the added side, pairings are ambiguous — they must be reported as
// plain Added/Removed, not Modified.
func TestComputeDiff_AmbiguousSameKey_NotModified(t *testing.T) {
	a1 := diff.Button{Label: "dup", Message: "m", Visible: true}
	a2 := diff.Button{Label: "dup", Message: "m", Visible: false}
	b1 := diff.Button{Label: "dup", Message: "m", Visible: true, Pronunciation: "x"}
	b2 := diff.Button{Label: "dup", Message: "m", Visible: false, Pronunciation: "y"}

	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(a1, a2)},
		diff.ButtonMap{"Home": makeSet(b1, b2)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if len(ch.Modified) != 0 {
		t.Fatalf("expected no Modified entries for ambiguous key, got %+v", ch.Modified)
	}
	if len(ch.Added) != 2 || len(ch.Removed) != 2 {
		t.Fatalf("want 2 Added and 2 Removed, got Added=%v Removed=%v", ch.Added, ch.Removed)
	}
}

func TestComputeDiff_ChangedPagesSortedByPageName(t *testing.T) {
	a := diff.Button{Label: "a", Visible: true}
	b := diff.Button{Label: "b", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Z": makeSet(), "M": makeSet()},
		diff.ButtonMap{"Z": makeSet(a), "M": makeSet(b)},
		diff.PageSet{"Z": {}, "M": {}},
		diff.PageSet{"Z": {}, "M": {}},
	)
	if len(result.ChangedPages) != 2 {
		t.Fatalf("expected 2 changed pages, got %d", len(result.ChangedPages))
	}
	if result.ChangedPages[0].PageName != "M" || result.ChangedPages[1].PageName != "Z" {
		t.Errorf("ChangedPages order: got [%s, %s], want [M, Z]",
			result.ChangedPages[0].PageName, result.ChangedPages[1].PageName)
	}
}

func TestComputeDiff_AddedRemovedButtonsSortedByLabel(t *testing.T) {
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet()},
		diff.ButtonMap{"Home": makeSet(
			diff.Button{Label: "zebra", Visible: true},
			diff.Button{Label: "apple", Visible: true},
		)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if len(ch.Added) != 2 {
		t.Fatalf("Added: got %+v", ch.Added)
	}
	if ch.Added[0].Label != "apple" || ch.Added[1].Label != "zebra" {
		t.Errorf("Added sort: got labels [%s, %s], want [apple, zebra]", ch.Added[0].Label, ch.Added[1].Label)
	}

	result2 := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(
			diff.Button{Label: "zebra", Visible: true},
			diff.Button{Label: "apple", Visible: true},
		)},
		diff.ButtonMap{"Home": makeSet()},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result2.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page (removed case), got %+v", result2.ChangedPages)
	}
	ch2 := result2.ChangedPages[0]
	if ch2.Removed[0].Label != "apple" || ch2.Removed[1].Label != "zebra" {
		t.Errorf("Removed sort: got labels [%s, %s], want [apple, zebra]", ch2.Removed[0].Label, ch2.Removed[1].Label)
	}
}

func TestComputeDiff_ModifiedSortedByKeyLabel(t *testing.T) {
	m1 := diff.Button{Label: "beta", Message: "m", Visible: true}
	m1a := diff.Button{Label: "beta", Message: "m", Visible: false}
	m2 := diff.Button{Label: "alpha", Message: "m", Visible: true}
	m2a := diff.Button{Label: "alpha", Message: "m", Visible: false}

	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(m1, m2)},
		diff.ButtonMap{"Home": makeSet(m1a, m2a)},
		diff.PageSet{"Home": {}}, diff.PageSet{"Home": {}},
	)
	if len(result.ChangedPages) != 1 {
		t.Fatalf("expected 1 changed page, got %+v", result.ChangedPages)
	}
	ch := result.ChangedPages[0]
	if len(ch.Modified) != 2 {
		t.Fatalf("Modified: got %+v", ch.Modified)
	}
	if ch.Modified[0].Key.Label != "alpha" || ch.Modified[1].Key.Label != "beta" {
		t.Errorf("Modified sort: got [%s, %s], want [alpha, beta]",
			ch.Modified[0].Key.Label, ch.Modified[1].Key.Label)
	}
}

func TestComputeDiff_AddedPageButtons_ContainsNewPageButtons(t *testing.T) {
	soccer := diff.Button{Label: "soccer", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet()},
		diff.ButtonMap{"Home": makeSet(), "Sports": makeSet(soccer)},
		diff.PageSet{"Home": {}},
		diff.PageSet{"Home": {}, "Sports": {}},
	)
	btns, ok := result.AddedPageButtons["Sports"]
	if !ok {
		t.Fatal("expected AddedPageButtons to contain 'Sports'")
	}
	if len(btns) != 1 {
		t.Errorf("expected 1 button for Sports, got %d", len(btns))
	}
	for _, btn := range btns {
		if btn.Label != "soccer" {
			t.Errorf("button label: got %q, want 'soccer'", btn.Label)
		}
	}
}

func TestComputeDiff_RemovedPageButtons_ContainsOldPageButtons(t *testing.T) {
	soccer := diff.Button{Label: "soccer", Visible: true}
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet(), "Sports": makeSet(soccer)},
		diff.ButtonMap{"Home": makeSet()},
		diff.PageSet{"Home": {}, "Sports": {}},
		diff.PageSet{"Home": {}},
	)
	btns, ok := result.RemovedPageButtons["Sports"]
	if !ok {
		t.Fatal("expected RemovedPageButtons to contain 'Sports'")
	}
	if len(btns) != 1 {
		t.Errorf("expected 1 button for Sports, got %d", len(btns))
	}
	for _, btn := range btns {
		if btn.Label != "soccer" {
			t.Errorf("button label: got %q, want 'soccer'", btn.Label)
		}
	}
}

func TestComputeDiff_AddedPageButtons_EmptyPageHasNoEntry(t *testing.T) {
	result := diff.ComputeDiff(
		diff.ButtonMap{"Home": makeSet()},
		diff.ButtonMap{"Home": makeSet(), "Empty": makeSet()},
		diff.PageSet{"Home": {}},
		diff.PageSet{"Home": {}, "Empty": {}},
	)
	if _, ok := result.AddedPageButtons["Empty"]; ok {
		t.Error("expected no entry for a page with no buttons")
	}
}

func TestComputeDiff_NilOldButtonMapUsesEmptySets(t *testing.T) {
	yes := diff.Button{Label: "yes", Visible: true}
	result := diff.ComputeDiff(
		nil,
		diff.ButtonMap{"Home": makeSet(yes)},
		nil,
		diff.PageSet{"Home": {}},
	)
	if len(result.AddedPages) != 1 || result.AddedPages[0] != "Home" {
		t.Errorf("AddedPages: got %v, want [Home]", result.AddedPages)
	}
	if len(result.ChangedPages) != 0 {
		t.Errorf("new-only page should not appear in ChangedPages: %+v", result.ChangedPages)
	}
}
