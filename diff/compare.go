package diff

import (
	"sort"
)

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
		sort.Slice(pureAdded, func(i, j int) bool { return pureAdded[i].Label < pureAdded[j].Label })
		sort.Slice(pureRemoved, func(i, j int) bool { return pureRemoved[i].Label < pureRemoved[j].Label })
		sort.Slice(modified, func(i, j int) bool { return modified[i].Key.Label < modified[j].Key.Label })

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
