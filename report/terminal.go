package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

// PrintDiff writes a human-readable summary of d to stdout.
func PrintDiff(d diff.Diff) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("  VOCAB DIFF")
	fmt.Printf("  OLD: %s\n", d.OldLabel)
	fmt.Printf("  NEW: %s\n", d.NewLabel)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	if len(d.AddedPages) > 0 {
		fmt.Printf("NEW PAGES (%d):\n", len(d.AddedPages))
		for _, p := range d.AddedPages {
			fmt.Printf("  + %s\n", p)
		}
		fmt.Println()
	}

	if len(d.RemovedPages) > 0 {
		fmt.Printf("REMOVED PAGES (%d):\n", len(d.RemovedPages))
		for _, p := range d.RemovedPages {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
	}

	if len(d.ChangedPages) > 0 {
		fmt.Printf("CHANGED PAGES (%d):\n\n", len(d.ChangedPages))
		for _, ch := range d.ChangedPages {
			fmt.Printf("  Page: %s\n", ch.PageName)
			for _, btn := range ch.Added {
				fmt.Println(formatButton("+", btn, "speaks"))
			}
			for _, btn := range ch.Removed {
				fmt.Println(formatButton("-", btn, "spoke"))
			}
			for _, bc := range ch.Modified {
				fmt.Printf("    ~ %q (modified)\n", bc.Key.Label)
				if bc.Before.Visible != bc.After.Visible {
					fmt.Printf("        visible: %s → %s\n",
						visStr(bc.Before.Visible), visStr(bc.After.Visible))
				}
				if bc.Before.Pronunciation != bc.After.Pronunciation {
					fmt.Printf("        pronunciation: %s → %s\n",
						quotedOrNone(bc.Before.Pronunciation),
						quotedOrNone(bc.After.Pronunciation))
				}
				oldSet := sliceToSet(bc.Before.Actions)
				newSet := sliceToSet(bc.After.Actions)
				var removedActs, addedActs []string
				for act := range oldSet {
					if _, ok := newSet[act]; !ok {
						removedActs = append(removedActs, act)
					}
				}
				for act := range newSet {
					if _, ok := oldSet[act]; !ok {
						addedActs = append(addedActs, act)
					}
				}
				sort.Strings(removedActs)
				sort.Strings(addedActs)
				for _, act := range removedActs {
					fmt.Printf("        action removed: %s\n", act)
				}
				for _, act := range addedActs {
					fmt.Printf("        action added:   %s\n", act)
				}
			}
			fmt.Println()
		}
	} else {
		fmt.Println("No button changes found on existing pages.")
	}

	wfc := d.WordFormChanges
	total := len(wfc.Added) + len(wfc.Removed) + len(wfc.Modified)
	if total > 0 {
		fmt.Printf("WORD-FORM CHANGES (%d):\n\n", total)
		printModifierGroup(wfc.Added, "+", "speaks", func(mc diff.ModifierChange) diff.Button { return mc.After })
		printModifierGroup(wfc.Removed, "-", "spoke", func(mc diff.ModifierChange) diff.Button { return mc.Before })
		for i := 0; i < len(wfc.Modified); {
			setName := wfc.Modified[i].ButtonSetName
			j := i
			for j < len(wfc.Modified) && wfc.Modified[j].ButtonSetName == setName {
				j++
			}
			fmt.Printf("  Word form: %s%s\n", setName, pagesSuffix(wfc.Modified[i].Pages))
			for _, mc := range wfc.Modified[i:j] {
				fmt.Printf("    ~ %q (%s, modified)\n", mc.After.Label, formLabel(mc.Key.FormIndex))
				if mc.Before.Pronunciation != mc.After.Pronunciation {
					fmt.Printf("        pronunciation: %s → %s\n",
						quotedOrNone(mc.Before.Pronunciation),
						quotedOrNone(mc.After.Pronunciation))
				}
			}
			fmt.Println()
			i = j
		}
	}

	totalAdded, totalRemoved, totalModified := 0, 0, 0
	for _, ch := range d.ChangedPages {
		totalAdded += len(ch.Added)
		totalRemoved += len(ch.Removed)
		totalModified += len(ch.Modified)
	}
	fmt.Printf("%s\n", strings.Repeat("─", 60))
	fmt.Printf("Summary: %d page(s) added, %d page(s) removed, %d page(s) with button changes.\n",
		len(d.AddedPages), len(d.RemovedPages), len(d.ChangedPages))
	fmt.Printf("         %d button(s) added, %d button(s) removed, %d button(s) with changed properties.\n\n",
		totalAdded, totalRemoved, totalModified)
}

// printModifierGroup prints Added or Removed modifier changes grouped by button set name.
// btnFn extracts the relevant Button from each change (mc.After for added, mc.Before for removed).
func printModifierGroup(changes []diff.ModifierChange, prefix, verb string, btnFn func(diff.ModifierChange) diff.Button) {
	for i := 0; i < len(changes); {
		setName := changes[i].ButtonSetName
		j := i
		for j < len(changes) && changes[j].ButtonSetName == setName {
			j++
		}
		fmt.Printf("  Word form: %s%s\n", setName, pagesSuffix(changes[i].Pages))
		for _, mc := range changes[i:j] {
			b := btnFn(mc)
			line := fmt.Sprintf("    %s %q (%s)", prefix, b.Label, formLabel(mc.Key.FormIndex))
			if b.Message != "" && b.Message != b.Label {
				line += fmt.Sprintf("  →  %s: %q", verb, b.Message)
			}
			if b.Pronunciation != "" {
				line += fmt.Sprintf("  →  pronounced: %q", b.Pronunciation)
			}
			if !b.Visible {
				line += " [hidden]"
			}
			fmt.Println(line)
		}
		fmt.Println()
		i = j
	}
}

func pagesSuffix(pages []string) string {
	if len(pages) == 0 {
		return ""
	}
	const max = 3
	label := "page"
	if len(pages) > 1 {
		label = "pages"
	}
	if len(pages) <= max {
		return "  (" + label + ": " + strings.Join(pages, ", ") + ")"
	}
	return fmt.Sprintf("  (%s: %s, +%d more)", label, strings.Join(pages[:max], ", "), len(pages)-max)
}

func formLabel(index int) string {
	if index == 0 {
		return "base"
	}
	return fmt.Sprintf("form #%d", index)
}

func formatButton(prefix string, btn diff.Button, verb string) string {
	vis := ""
	if !btn.Visible {
		vis = " [hidden]"
	}
	msg := ""
	if btn.Message != "" && btn.Message != btn.Label {
		msg = fmt.Sprintf("  →  %s: %q", verb, btn.Message)
	}
	pron := ""
	if btn.Pronunciation != "" {
		pron = fmt.Sprintf("  →  pronounced: %q", btn.Pronunciation)
	}
	return fmt.Sprintf("    %s %q%s%s%s", prefix, btn.Label, msg, pron, vis)
}

func visStr(v bool) string {
	if v {
		return "shown"
	}
	return "hidden"
}

func quotedOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return fmt.Sprintf("%q", s)
}

func sliceToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}
