package diff_test

import (
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

func entry(name string, btn diff.Button) diff.ModifierEntry {
	return diff.ModifierEntry{Name: name, Button: btn}
}

func TestComputeModifierDiff_AddedModifier(t *testing.T) {
	oldMods := diff.ModifierMap{}
	newMods := diff.ModifierMap{
		diff.ModifierKey{ButtonSetRID: "rid-good", FormIndex: 0}: entry("good", diff.Button{Label: "good", Message: "good job kid"}),
	}
	result := diff.ComputeModifierDiff(oldMods, newMods)
	if len(result.Added) != 1 {
		t.Fatalf("Added: got %d, want 1", len(result.Added))
	}
	if result.Added[0].Key != (diff.ModifierKey{ButtonSetRID: "rid-good", FormIndex: 0}) {
		t.Errorf("Added key: got %+v", result.Added[0].Key)
	}
	if result.Added[0].ButtonSetName != "good" {
		t.Errorf("Added ButtonSetName: got %q, want 'good'", result.Added[0].ButtonSetName)
	}
	if result.Added[0].After.Message != "good job kid" {
		t.Errorf("Added After.Message: got %q", result.Added[0].After.Message)
	}
	if len(result.Removed) != 0 || len(result.Modified) != 0 {
		t.Errorf("expected no Removed/Modified, got %+v", result)
	}
}

func TestComputeModifierDiff_RemovedModifier(t *testing.T) {
	oldMods := diff.ModifierMap{
		diff.ModifierKey{ButtonSetRID: "rid-go", FormIndex: 0}: entry("go", diff.Button{Label: "go"}),
	}
	newMods := diff.ModifierMap{}
	result := diff.ComputeModifierDiff(oldMods, newMods)
	if len(result.Removed) != 1 {
		t.Fatalf("Removed: got %d, want 1", len(result.Removed))
	}
	if result.Removed[0].Key != (diff.ModifierKey{ButtonSetRID: "rid-go", FormIndex: 0}) {
		t.Errorf("Removed key: got %+v", result.Removed[0].Key)
	}
	if result.Removed[0].ButtonSetName != "go" {
		t.Errorf("Removed ButtonSetName: got %q, want 'go'", result.Removed[0].ButtonSetName)
	}
	if len(result.Added) != 0 || len(result.Modified) != 0 {
		t.Errorf("expected no Added/Modified, got %+v", result)
	}
}

func TestComputeModifierDiff_ModifiedModifier(t *testing.T) {
	key := diff.ModifierKey{ButtonSetRID: "rid-good", FormIndex: 0}
	before := diff.Button{Label: "good", Pronunciation: ""}
	after := diff.Button{Label: "good", Pronunciation: "good job kid"}

	result := diff.ComputeModifierDiff(
		diff.ModifierMap{key: entry("good", before)},
		diff.ModifierMap{key: entry("good", after)},
	)
	if len(result.Modified) != 1 {
		t.Fatalf("Modified: got %d, want 1", len(result.Modified))
	}
	mc := result.Modified[0]
	if mc.Key != key {
		t.Errorf("Modified key: got %+v, want %+v", mc.Key, key)
	}
	if mc.ButtonSetName != "good" {
		t.Errorf("Modified ButtonSetName: got %q, want 'good'", mc.ButtonSetName)
	}
	if mc.Before.Pronunciation != "" || mc.After.Pronunciation != "good job kid" {
		t.Errorf("Modified Before/After pronunciation: %q / %q", mc.Before.Pronunciation, mc.After.Pronunciation)
	}
	if len(result.Added) != 0 || len(result.Removed) != 0 {
		t.Errorf("expected no Added/Removed, got %+v", result)
	}
}

func TestComputeModifierDiff_UnchangedModifier(t *testing.T) {
	key := diff.ModifierKey{ButtonSetRID: "rid-eat", FormIndex: 0}
	btn := diff.Button{Label: "eat", Visible: true}
	result := diff.ComputeModifierDiff(
		diff.ModifierMap{key: entry("eat", btn)},
		diff.ModifierMap{key: entry("eat", btn)},
	)
	if len(result.Added)+len(result.Removed)+len(result.Modified) != 0 {
		t.Errorf("expected no changes for identical modifiers, got %+v", result)
	}
}

func TestComputeModifierDiff_EmptyMaps(t *testing.T) {
	result := diff.ComputeModifierDiff(diff.ModifierMap{}, diff.ModifierMap{})
	if len(result.Added)+len(result.Removed)+len(result.Modified) != 0 {
		t.Errorf("expected empty result for empty maps, got %+v", result)
	}
}

// Results must be sorted by ButtonSetName then FormIndex so output is
// deterministic regardless of map iteration order.
func TestComputeModifierDiff_SortedByWordThenFormIndex(t *testing.T) {
	oldMods := diff.ModifierMap{}
	newMods := diff.ModifierMap{
		diff.ModifierKey{ButtonSetRID: "rid-zebra", FormIndex: 0}: entry("zebra", diff.Button{Label: "zebra"}),
		diff.ModifierKey{ButtonSetRID: "rid-apple", FormIndex: 6}: entry("apple", diff.Button{Label: "apple-6"}),
		diff.ModifierKey{ButtonSetRID: "rid-apple", FormIndex: 0}: entry("apple", diff.Button{Label: "apple-0"}),
	}
	result := diff.ComputeModifierDiff(oldMods, newMods)
	if len(result.Added) != 3 {
		t.Fatalf("Added: got %d, want 3", len(result.Added))
	}
	wantNames := []string{"apple", "apple", "zebra"}
	wantForms := []int{0, 6, 0}
	for i, got := range result.Added {
		if got.ButtonSetName != wantNames[i] || got.Key.FormIndex != wantForms[i] {
			t.Errorf("Added[%d]: got name=%q form=%d, want name=%q form=%d",
				i, got.ButtonSetName, got.Key.FormIndex, wantNames[i], wantForms[i])
		}
	}
}

// CompareFiles must populate Diff.WordFormChanges. Using the same file on
// both sides guarantees zero changes — a non-nil but empty result confirms
// the field is wired up rather than left at its zero value by accident.
func TestCompareFiles_WordFormChanges_SameFileNoChanges(t *testing.T) {
	d, err := diff.CompareFiles(fixtureFile, fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	wfc := d.WordFormChanges
	if len(wfc.Added)+len(wfc.Removed)+len(wfc.Modified) != 0 {
		t.Errorf("expected zero word-form changes for identical files, got %+v", wfc)
	}
}
