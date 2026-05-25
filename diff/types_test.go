package diff_test

import (
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

// TestModifierTypes_Exist is a compile-time guardrail: it will fail to build
// until all four types related to ModifierSetDiff are present in types.go.
func TestModifierTypes_Exist(t *testing.T) {
	key := diff.ModifierKey{ButtonSetName: "good", FormIndex: 0}
	if key.ButtonSetName != "good" || key.FormIndex != 0 {
		t.Errorf("ModifierKey fields: got %+v", key)
	}

	mc := diff.ModifierChange{
		Key:           key,
		Before:        diff.Button{Label: "good"},
		After:         diff.Button{Label: "good", Pronunciation: "good job kid"},
	}
	_ = mc

	msd := diff.ModifierSetDiff{
		Added:    []diff.ModifierChange{mc},
		Removed:  []diff.ModifierChange{},
		Modified: []diff.ModifierChange{},
	}
	_ = msd

	var d diff.Diff
	_ = d.WordFormChanges // fails to compile until Diff.WordFormChanges is added
}
