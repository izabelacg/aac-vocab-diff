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
