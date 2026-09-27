# Word-Form Changes in TouchChat AAC Vocabulary

## What Is a "Word Form" in AAC?

In natural language, one word has many **inflected forms** — variations that carry grammatical meaning:

```
go → goes / went / going / gone
good → better / best
cat → cats / cat's
```

In an AAC (Augmentative and Alternative Communication) device, the person taps buttons to speak. The challenge: you can't put every inflected form as its own button — there simply isn't enough space, and the motor cost is too high.

**WordPower's solution:** one button per *core word*, with a linked set of inflected-form buttons that change what appears based on context.

---

## Why Not Just Use a Regular Button?

Here's the core trade-off:

| Approach | Buttons needed for "go" | Problem |
|---|---|---|
| One button per form | 5 (`go`, `goes`, `went`, `going`, `gone`) | For 200 core words × 5 forms = 1,000 buttons |
| One core button + modifiers | 1 + shared grammar buttons | Scales; preserves motor planning |

**Motor planning** is a critical AAC principle: users with motor impairments develop muscle memory for button locations. If "go" is always in the same spot, the user can internalize that location like a touch-typist. Splitting it into 5 separate buttons breaks that stability.

**Core vocabulary:** a small set of high-frequency *core* words is commonly cited as covering most of what people say day to day (figures around 80% are often quoted; *source not verified here*). Morphology (grammar endings) multiplies each word's expressive power without multiplying button count.

---

## How WordPower Implements This Technically

Inside a TouchChat `.c4v` vocabulary file (a SQLite database), the relevant tables are below. The form indexes and labels come from a real WordPower vocabulary (only a few rows of each set are shown); the `id` and `rid` values are illustrative.

```
┌──────────────────────────────────────────────────────────────────┐
│                         resources                                │
│  id │ rid (stable GUID)  │ name                                  │
│  1  │ "abc-123"           │ "go"                                 │
│  2  │ "def-456"           │ "eat"                                │
│  3  │ "ghi-789"           │ "eat"   ← same name, different set   │
└──────────────────────────────────────────────────────────────────┘
                │
                │ (one resource = one button set)
                │
                ▼
┌──────────────────────────────────────────────────────────────────┐
│                         button_sets                              │
│  id │ resource_id                                                │
│  1  │ 1  (→ "go")                                                │
│  2  │ 2  (→ "eat")                                               │
│  3  │ 3  (→ "eat")                                               │
└──────────────────────────────────────────────────────────────────┘
                │
                │ (one button_set = one family of forms)
                │
                ▼
┌──────────────────────────────────────────────────────────────────┐
│                     button_set_modifiers                         │
│  id │ button_set_id │ button_id │ modifier (form index)          │
│  1  │ 1             │ 10        │ 0    ← base form: "go"         │
│  2  │ 1             │ 11        │ 6    ← form #6: "going"        │
│  3  │ 1             │ 12        │ 21   ← form #21: "gone"        │
│  4  │ 2             │ 20        │ 0    ← base form: "eat"        │
│  5  │ 2             │ 21        │ 6    ← form #6: "eating"       │
│  6  │ 3             │ 30        │ 0    ← base form: "eat"        │
└──────────────────────────────────────────────────────────────────┘
                │
                │ (button_id → buttons.id)
                │
                ▼
┌──────────────────────────────────────────────────────────────────┐
│                           buttons                                │
│  id │ label     │ message    │ visible │ pronunciation │ actions  │
│  10 │ "go"      │ "go"       │ true    │               │ [...]    │
│  11 │ "going"   │ "going"    │ true    │               │ [...]    │
│  12 │ "gone"    │ "gone"     │ true    │               │ [...]    │
│  20 │ "eat"     │ "eat"      │ true    │               │ [...]    │
│  21 │ "eating"  │ "eating"   │ true    │               │ [...]    │
│  30 │ "eat"     │ "eat"      │ true    │               │ [...]    │
└──────────────────────────────────────────────────────────────────┘
```

**Key points about this schema:**

- `resources.rid` is a **stable GUID** — it survives vocabulary updates and identifies the same button set across versions. This is what the diff tool uses as the stable identity key, not the label.
- **One word can have many button sets.** In a real WordPower vocabulary, "eat" has 15 separate button sets, and over 200 word names are shared by more than one set. The name is a display label, not an identity.
- `button_set_modifiers.modifier` is an **integer index** (0 = base form, other numbers = other forms). The numbers are internal to TouchChat and **don't map to a fixed grammatical meaning**: "eating" appears at forms 6, 20, 43, 64, 66 and 67 across sets, and some forms aren't grammatical at all (form 13 of "good" is 👍). Indexes also skip numbers — the "go" set above has forms 0, 6, 8, 13, 20, 21, 29, 64, 66 and 67.
- The `button_set_modifiers` table is the **join table** that maps "which form of which button set connects to which actual button."

---

## What Happens at Runtime (User Perspective)

When the user taps a word button that has word forms:

```
User taps "go" (base form)    →    "go" appears in message bar
                                         │
                              ┌──────────▼──────────┐
                              │  Grammar page pops  │
                              │  up or modifier     │
                              │  buttons light up:  │
                              │                     │
                              │  [goes] [went]      │
                              │  [going] [gone]     │
                              └──────────┬──────────┘
                                         │
                              User taps "went"
                                         │
                                         ▼
                              "went" replaces "go" in message bar
```

In the data, these buttons carry **action code 71** (`"word form"`, with a form index as its value). It appears on the modifier buttons themselves and also on regular page buttons — in a real WordPower vocabulary, about 10,000 uses are on modifier buttons and about 1,500 on page buttons, which is why reports show chips like `word form: 38`. What exactly TouchChat does with it at runtime (for example, "replace the current word with that form") is *inferred, not verified*.

---

## Why This Matters for AAC Users Clinically

Two key clinical reasons:

**1. You can't learn what you can't produce.**

> *"AAC users cannot learn grammar if their AAC system does not have grammatical forms available."* — AssistiveWare

If "walked" doesn't exist in the vocabulary, the user can never produce past tense — not just as a communication barrier, but as a language *acquisition* barrier. SLPs modeling language on the device also can't demonstrate it.

**2. Morphological awareness supports literacy.**

Evidence-based morphology instruction improves vocabulary, comprehension, spelling, and decoding. AAC users working toward literacy need to encounter and produce inflected forms, not just base words.

**The SLP design tension:**

| Automatic grammar | Popup/explicit forms |
|---|---|
| System auto-applies endings | User selects the form they want |
| Easier for immediate communication | Builds rule internalization over time |
| Can prevent the user from *learning* the rule | Supports language development |

WordPower's popup-style approach (offering choices rather than auto-correcting) is generally preferred by SLPs for users still developing language.

---

## What This Tool Tracks as "Word-Form Changes"

When comparing two versions of a vocabulary file, the diff tool tracks changes to `button_set_modifiers` rows as `WordFormChanges`, separately from regular page button changes. A word-form change is:

- **Added** — a form that exists in the new vocabulary but not the old (e.g., a new form #21 "gone" was added to a "go" set)
- **Removed** — a form that existed in the old vocabulary but was deleted
- **Modified** — same `(ButtonSetRID, FormIndex)` key, but one of the five tracked fields changed: `Label`, `Message`, `Visible`, `Pronunciation`, or `Actions`

The identity key is `(rid, form_index)` — **not the label** — so if "going" is renamed to "GOING" on form #6 of a "go" set, that's a *modification*, not a remove+add. This is important for accurate diffing across vocabulary updates.

In the HTML report, word-form changes have their own section, **collapsed by default** so it stays out of printouts unless the reader expands it. Changes are grouped into one card **per word name** (sketch; page names and values are illustrative):

```
WORD-FORM CHANGES   Show   Hidden from this printout (17 words).

┌─ go ──────────────────────────────────────────────────────────────┐
│  Core → Actions                          ← pages the sets are on   │
│  Button label       Spoken message  Pronunciation  Visible  Actions│
│  ~ go (base)        go              — → "go go"    Yes      …      │
│  + gone (form #21)  gone            —              Yes      …      │
└────────────────────────────────────────────────────────────────────┘
```

Because cards are grouped by name and one name can belong to many button sets (see above), a single card can merge forms from several sets. That's why a card may list the same form number more than once, e.g. several "(form #28)" rows under one word. The grouping keeps everything a therapist would think of as "the word *go*" in one place; the trade-off is that it doesn't show which of the word's sets each row came from.

---

## Sources

- [TouchChat Button Actions](https://touchchatapp.com/support/touchchat-button-actions) — official Saltillo support page
- [WordPower Vocabulary — PRC-Saltillo documentation](https://documentation.prc-saltillo.com/docs/wordpower-vocabulary)
- [Teaching AAC Users Grammar — AssistiveWare](https://www.assistiveware.com/learn-aac/teach-grammar)
- [Find Different Forms of a Word — AssistiveWare Support](https://www.assistiveware.com/support/proloquo2go/basics/pop-up)
- [Grammar Support — AssistiveWare Blog](https://www.assistiveware.com/blog/grammar-support)
- [Teach Me to Fish: Building Grammar Skills — AssistiveWare Blog](https://www.assistiveware.com/blog/teach-me-to-fish-building-grammar-skills)
- [Features of WordPower — Smartbox Hub](https://hub.thinksmartbox.com/knowledgebase/features-of-wordpower/)
