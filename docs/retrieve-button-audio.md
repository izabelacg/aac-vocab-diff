# Retrieving audio attached to a TouchChat button

Like photos, a button's audio isn't stored next to the button itself — it's
reached through the `actions` table. Unlike photos, the audio bytes live
**inside the same `.c4v` file**, not a separate archive member — there's no
`Audio.c4s` sidecar the way there's an `Images.c4s`.

## Background

Audio comes from the button's **"play sound" action** (action `code = 4`):

- `buttons.resource_id → actions.resource_id` (where `actions.code = 4`)
- `actions.id → action_data.action_id`, where:
  - `key = 0` → the `rid` of the sound resource
  - `key = 1` → a human-readable label (e.g. `Recording: 2024-08-23, 9:06 AM`,
    or a renamed one like `What a beautiful day!`)
- `action_data.value (key=0) → resources.rid → resources.id`
- `resources.id → sounds.resource_id → sounds.data` — the actual audio bytes

All sound resources have `resources.type = 3`, whether they're built-in
library sounds (`cat_meow_x.wav`, `Q.mp3`, …) or user recordings.

Audio blobs are WAV (`RIFF...WAVE`, magic bytes `52 49 46 46`) for recordings
made in-app, though the library also ships some `.mp3` sounds under the same
mechanism.

## Steps

### 1. Unzip the `.ce` export

```bash
unzip "MyVocab - 2026-05-30.ce" -d some-folder
cd some-folder
```

### 2. Find the button and its sound action

```sql
sqlite3 "MyVocab - 2026-05-30.c4v" "
SELECT b.id, b.label, b.message, ad0.value AS sound_rid, ad1.value AS sound_label
FROM buttons b
JOIN actions a       ON a.resource_id = b.resource_id AND a.code = 4
JOIN action_data ad0 ON ad0.action_id = a.id AND ad0.key = 0
LEFT JOIN action_data ad1 ON ad1.action_id = a.id AND ad1.key = 1
WHERE b.label LIKE '%dive in%';
"
```

If the button has no `code = 4` action, it doesn't play a custom/recorded
sound (it may just speak the label via TTS instead).

### 3. Write the blob out to a file

No `ATTACH` needed — the sound lives in the same database as the button:

```sql
sqlite3 "MyVocab - 2026-05-30.c4v" "
SELECT writefile('lets-dive-in.wav', s.data)
FROM action_data ad
JOIN resources r ON r.rid = ad.value
JOIN sounds s     ON s.resource_id = r.id
WHERE ad.key = 0 AND ad.value = '{7D05E711-F8990F94-D9BD3324-15243013}';
"
```

## One-liner (find + extract in a single command)

```bash
sqlite3 "MyVocab - 2026-05-30.c4v" "
SELECT writefile(b.label || '.wav', s.data)
FROM buttons b
JOIN actions a       ON a.resource_id = b.resource_id AND a.code = 4
JOIN action_data ad  ON ad.action_id = a.id AND ad.key = 0
JOIN resources r     ON r.rid = ad.value
JOIN sounds s        ON s.resource_id = r.id
WHERE b.label LIKE '%dive in%';
"
```

## Telling custom audio apart from pre-existing vocabulary audio

There's no explicit "is_custom" flag — TouchChat leaves a naming convention
instead, so this is approximate, not exact:

- **Auto-named recordings** — TouchChat's own recorder names them
  `Recording: <date>, <time>` (visible in `action_data.value` at `key = 1`,
  and mirrored in `resources.name`). Any resource matching that pattern is
  almost certainly something recorded on-device, not shipped vocabulary.
- **Renamed recordings** — a user can rename a recording after making it
  (e.g. `FestaDePalavras`, `What a beautiful day!`), which loses the
  `Recording:` prefix. These won't match the naming heuristic above.
- **Library sounds** — ship named like real asset filenames, e.g.
  `cat_meow_x.wav`, `duck.wav`, `Q.mp3`.

The naming heuristic catches the common case cheaply:

```sql
sqlite3 "MyVocab - 2026-05-30.c4v" "
SELECT r.name, length(s.data)
FROM resources r
JOIN sounds s ON s.resource_id = r.id
WHERE r.type = 3 AND r.name LIKE 'Recording:%';
"
```

But it **won't catch renamed recordings**, and can't distinguish those from
stock sounds by name alone. The only authoritative way to tell is to diff
the sound resource set (`rid`s where `resources.type = 3`) between:

- the unmodified base vocab export (e.g. one of the `WordPower60 Basic SS_*.ce`
  files under `diff/testdata/`) and your customized `.ce` — anything present
  only in the customized file was added by you, or
- two of your own dated snapshots (`tmp/vocabs/*.ce`) — the earliest snapshot
  a given `rid` appears in tells you roughly when it was added.

That's a natural extension of what this repo's diff tooling already does for
buttons and pages, just not yet wired up for the sound resource set
specifically — see [diff/load.go](../diff/load.go#L362-L372), which already
resolves a "play sound" action's target resource name for display.
