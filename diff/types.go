package diff

import (
	"fmt"
	"strings"
)

// Button holds every diffable field of a single TouchChat™ button.
type Button struct {
	Label         string
	Message       string // empty string means the device speaks the label text
	Visible       bool
	Pronunciation string
	Actions       []string // human-readable, e.g. ["speak", "navigate to page: \"Home\""]
}

// Fingerprint encodes all fields into a single string so a Button can be used
// as a map key. We need this because Button contains Actions []string, and
// slices are not comparable in Go — you can't use a Button directly as a map key.
//
// \x1f (unit separator) and \x1e (record separator) are ASCII control
// characters that won't appear in any real button text, so they're safe
// delimiters that won't cause false collisions.
func (b Button) Fingerprint() string {
	return fmt.Sprintf("%s\x1f%s\x1f%v\x1f%s\x1f%s",
		b.Label, b.Message, b.Visible, b.Pronunciation,
		strings.Join(b.Actions, "\x1e"))
}

// ButtonKey identifies a button by label+message only — intentionally ignoring
// Visible, Pronunciation, and Actions. Used in the diff to detect "modified"
// buttons: if old and new share the same ButtonKey but have different
// fingerprints, the button was edited rather than replaced.
type ButtonKey struct{ Label, Message string }

// ── Diff classification rules ────────────────────────────────────────────────
//
// These rules define when a button is considered added, removed, or modified.
// They are intentional design decisions, not incidental implementation details.
// Revisit this block if the desired user-facing behaviour changes.
//
// IDENTITY
//   A button's identity is its (Label, Message) pair — the ButtonKey.
//   All other fields (Visible, Pronunciation, Actions) are considered
//   "properties" of that identity, not part of the identity itself.
//
//   Consequence: changing a button's label or message is treated as
//   removing the old button and adding a new one, NOT as a modification.
//   There is no concept of "renaming" a button.
//
// ADDED
//   A button is added when it appears on a page in the new file but has
//   no matching ButtonKey on that same page in the old file.
//   Also: an entire page present in new but absent in old is "added" —
//   its buttons are not individually reported, only the page name is.
//
// REMOVED
//   A button is removed when it appears on a page in the old file but has
//   no matching ButtonKey on that same page in the new file.
//   Also: an entire page present in old but absent in new is "removed" —
//   same rule as above, only the page name is reported.
//
// MODIFIED
//   A button is modified when ALL of the following hold:
//     1. The same ButtonKey (Label + Message) exists on the page in both
//        old and new.
//     2. At least one property (Visible, Pronunciation, or Actions) differs.
//     3. Exactly one old button and exactly one new button share that key
//        on the page — so the pairing is unambiguous.
//
//   If condition 3 is NOT met (e.g. two buttons with the same label and
//   message exist on the same page in old or new), all of them are treated
//   as pure added/removed instead. We cannot reliably say which old button
//   corresponds to which new button, so we do not guess.
//
// UNCHANGED
//   A button whose fingerprint (all five fields) is identical in old and
//   new is silently ignored — it does not appear in any diff output.

// PageSet is the set of page names present in a vocabulary file.
// Using map[string]struct{} is Go's idiomatic set: the empty struct{}
// value takes zero bytes — we only care whether a key exists, not its value.
//
//	pages := PageSet{"Home": {}, "Sports": {}}
//	_, exists := pages["Home"]   // true
type PageSet = map[string]struct{}

// ButtonSet is the set of buttons on a single page.
// Key   = Button.Fingerprint() — unique per distinct button state
// Value = the Button itself, so we can read its fields after a lookup
//
// Set-difference in Go (buttons present in new but not old = added):
//
//	for fp, btn := range newSet {
//	    if _, inOld := oldSet[fp]; !inOld {
//	        // btn was added
//	    }
//	}
type ButtonSet = map[string]Button

// ButtonMap is the top-level structure loaded from a .c4v database.
// Key   = page name  (e.g. "Home", "Sports")
// Value = ButtonSet  (all buttons on that page, keyed by fingerprint)
//
// Example:
//
//	bm := ButtonMap{
//	    "Home":   ButtonSet{"yes\x1f...": Button{Label: "yes", ...}},
//	    "Sports": ButtonSet{"soccer\x1f...": Button{Label: "soccer", ...}},
//	}
type ButtonMap = map[string]ButtonSet

// Diff is the structural result of comparing an old snapshot to a new one.
// OldLabel and NewLabel are typically filled by the CLI from file paths;
// ComputeDiff leaves them empty.
type Diff struct {
	OldLabel           string
	NewLabel           string
	AddedPages         []string     // sorted
	RemovedPages       []string     // sorted
	ChangedPages       []PageChange // sorted by page name
	AddedPageButtons   ButtonMap    // buttons on each newly added page (from new file)
	RemovedPageButtons ButtonMap    // buttons on each removed page (from old file)
}

// PageChange summarizes button-level changes on a single page that exists in
// both versions (whole-page add/remove are reported only in AddedPages / RemovedPages).
type PageChange struct {
	PageName string
	Added    []Button
	Removed  []Button
	Modified []ButtonChange
}

// ButtonChange records a single logical edit: same ButtonKey, different fingerprint.
type ButtonChange struct {
	Key    ButtonKey
	Before Button
	After  Button
}
