package diff

import (
	"cmp"
	"slices"
	"sort"
)

// CompareButtons orders buttons by label, then message, then fingerprint.
// The fingerprint is unique within a page's ButtonSet, so the order is total
// and independent of map iteration order.
func CompareButtons(a, b Button) int {
	return cmp.Or(
		cmp.Compare(a.Label, b.Label),
		cmp.Compare(a.Message, b.Message),
		cmp.Compare(a.Fingerprint(), b.Fingerprint()),
	)
}

// compareModifierChanges orders by word, then form, then button-set RID;
// the key is unique, so ties between same-named sets are broken deterministically.
func compareModifierChanges(a, b ModifierChange) int {
	return cmp.Or(
		cmp.Compare(a.ButtonSetName, b.ButtonSetName),
		cmp.Compare(a.Key.FormIndex, b.Key.FormIndex),
		cmp.Compare(a.Key.ButtonSetRID, b.Key.ButtonSetRID),
	)
}

func ComputeDiff(oldBtns, newBtns ButtonMap, oldPages, newPages PageSet) Diff {
	// 1. Find added/removed pages using map existence checks (Go's set-difference).
	var added, removed []string
	for p := range newPages {
		if _, ok := oldPages[p]; !ok {
			added = append(added, p)
		}
	}
	for p := range oldPages {
		if _, ok := newPages[p]; !ok {
			removed = append(removed, p)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	// Capture the buttons for wholly added/removed pages so HTML can list them.
	addedPageBtns := ButtonMap{}
	for _, p := range added {
		if bs := newBtns[p]; len(bs) > 0 {
			addedPageBtns[p] = bs
		}
	}
	removedPageBtns := ButtonMap{}
	for _, p := range removed {
		if bs := oldBtns[p]; len(bs) > 0 {
			removedPageBtns[p] = bs
		}
	}

	// Build sets for O(1) skip-checks below.
	addedSet := make(map[string]struct{}, len(added))
	for _, p := range added {
		addedSet[p] = struct{}{}
	}
	removedSet := make(map[string]struct{}, len(removed))
	for _, p := range removed {
		removedSet[p] = struct{}{}
	}

	// 2. Walk every page that exists in either version and find button changes.
	allPages := map[string]struct{}{}
	for p := range oldPages {
		allPages[p] = struct{}{}
	}
	for p := range newPages {
		allPages[p] = struct{}{}
	}
	pageNames := make([]string, 0, len(allPages))
	for p := range allPages {
		pageNames = append(pageNames, p)
	}
	sort.Strings(pageNames)

	var changedPages []PageChange
	for _, page := range pageNames {
		if _, skip := addedSet[page]; skip {
			continue // whole page is new — not a "changed" page
		}
		if _, skip := removedSet[page]; skip {
			continue // whole page removed — not a "changed" page
		}

		oldSet := oldBtns[page] // nil-safe: ranging over nil map is a no-op
		newSet := newBtns[page]

		// Group buttons that appear only in new/old by their (label, message) key.
		// A button is "modified" if exactly one old and one new share the same key.
		addedByKey := map[ButtonKey][]Button{}
		removedByKey := map[ButtonKey][]Button{}

		for fp, btn := range newSet {
			if _, inOld := oldSet[fp]; !inOld {
				k := ButtonKey{btn.Label, btn.Message}
				addedByKey[k] = append(addedByKey[k], btn)
			}
		}
		for fp, btn := range oldSet {
			if _, inNew := newSet[fp]; !inNew {
				k := ButtonKey{btn.Label, btn.Message}
				removedByKey[k] = append(removedByKey[k], btn)
			}
		}

		// Collect all keys that appear in either group.
		allKeys := map[ButtonKey]struct{}{}
		for k := range addedByKey {
			allKeys[k] = struct{}{}
		}
		for k := range removedByKey {
			allKeys[k] = struct{}{}
		}

		var pureAdded, pureRemoved []Button
		var modified []ButtonChange
		for k := range allKeys {
			a, r := addedByKey[k], removedByKey[k]
			if len(a) == 1 && len(r) == 1 {
				// Unambiguous 1-to-1 match → modified
				modified = append(modified, ButtonChange{Key: k, Before: r[0], After: a[0]})
			} else {
				// Multiple or one-sided → treat as pure add/remove
				pureAdded = append(pureAdded, a...)
				pureRemoved = append(pureRemoved, r...)
			}
		}

		if len(pureAdded)+len(pureRemoved)+len(modified) == 0 {
			continue
		}

		// Sort for deterministic output regardless of map iteration order.
		slices.SortFunc(pureAdded, CompareButtons)
		slices.SortFunc(pureRemoved, CompareButtons)
		slices.SortFunc(modified, func(a, b ButtonChange) int {
			return cmp.Or(cmp.Compare(a.Key.Label, b.Key.Label), cmp.Compare(a.Key.Message, b.Key.Message))
		})

		changedPages = append(changedPages, PageChange{
			PageName: page,
			Added:    pureAdded,
			Removed:  pureRemoved,
			Modified: modified,
		})
	}

	return Diff{
		AddedPages:         added,
		RemovedPages:       removed,
		ChangedPages:       changedPages,
		AddedPageButtons:   addedPageBtns,
		RemovedPageButtons: removedPageBtns,
	}
}

func computeModifierDiff(old, new ModifierMap) ModifierSetDiff {
	var added, removed, modified []ModifierChange

	for key, newEntry := range new {
		if oldEntry, ok := old[key]; !ok {
			added = append(added, ModifierChange{Key: key, ButtonSetName: newEntry.Name, Pages: newEntry.Pages, After: newEntry.Button})
		} else if oldEntry.Button.Fingerprint() != newEntry.Button.Fingerprint() {
			modified = append(modified, ModifierChange{Key: key, ButtonSetName: oldEntry.Name, Pages: newEntry.Pages, Before: oldEntry.Button, After: newEntry.Button})
		}
	}
	for key, oldEntry := range old {
		if _, ok := new[key]; !ok {
			removed = append(removed, ModifierChange{Key: key, ButtonSetName: oldEntry.Name, Pages: oldEntry.Pages, Before: oldEntry.Button})
		}
	}

	slices.SortFunc(added, compareModifierChanges)
	slices.SortFunc(removed, compareModifierChanges)
	slices.SortFunc(modified, compareModifierChanges)

	return ModifierSetDiff{Added: added, Removed: removed, Modified: modified}
}

func ComputeModifierDiff(old, new ModifierMap) ModifierSetDiff {
	return computeModifierDiff(old, new)
}
