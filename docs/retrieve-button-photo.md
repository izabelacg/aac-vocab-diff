# Retrieving a photo attached to a TouchChat button

TouchChat vocab exports (`.ce` files) are zip archives. A button's photo
isn't stored next to the button itself — it's split across two SQLite
databases inside the archive, linked by a GUID (`rid`). This walks through
pulling one out by hand with the `sqlite3` CLI.

## Background

- `<name>.c4v` — the vocabulary DB (pages, buttons, actions, styles…).
- `Images.c4s` — a separate DB holding embedded image blobs.

Buttons get an image one of two ways:

- **Built-in SymbolStix icons** (the vast majority of buttons) —
  `buttons.resource_id → resources.rid`. This `rid` resolves against the
  symbol library bundled with the TouchChat app itself; it is **not**
  present in `Images.c4s`.
- **Custom photos** (assigned by a user, e.g. a person's face on a name
  button) — `buttons.symbol_link_id → symbol_links.rid →
  Images.c4s.symbols.data`. The actual PNG bytes live in `Images.c4s`.

So: if the button you want a photo for was populated with a real photo
(not a stock icon), it'll be reachable through `symbol_link_id`.

## Steps

### 1. Unzip the `.ce` export

```bash
unzip "MyVocab - 2026-05-30.ce" -d some-folder
cd some-folder
ls
# Images.c4s  MyVocab - 2026-05-30.c4v  Manifest.c4i  version.txt
```

### 2. Find the button

Search `buttons` by label, message, or pronunciation in the `.c4v` file:

```bash
sqlite3 "MyVocab - 2026-05-30.c4v" \
  "SELECT id, label, message, pronunciation, symbol_link_id
   FROM buttons
   WHERE label LIKE '%hands on top%';"
```

Note the `symbol_link_id` — that's what chains to the image. If it's
`NULL` or `0`, the button is using a built-in icon, not a custom photo,
and there's nothing to extract from `Images.c4s`.

### 3. Resolve the image through `symbol_links`

`Images.c4s` is a separate file, so attach it to the same `sqlite3`
session and join across the two databases:

```sql
ATTACH DATABASE 'Images.c4s' AS imgs;

SELECT s.width, s.height, length(s.data)
FROM symbol_links sl
JOIN imgs.symbols s ON lower(sl.rid) = lower(s.rid)
WHERE sl.id = 39150;   -- the symbol_link_id from step 2
```

(`lower()` on both sides guards against GUID case-formatting
differences between the two DBs.)

### 4. Write the blob out to a file

`writefile()` is a built-in SQLite CLI function — no scripting needed:

```sql
ATTACH DATABASE 'Images.c4s' AS imgs;

SELECT writefile('hands-on-top.png', s.data)
FROM symbol_links sl
JOIN imgs.symbols s ON lower(sl.rid) = lower(s.rid)
WHERE sl.id = 39150;
```

This prints the byte count written and drops `hands-on-top.png` (or
whatever path you give) next to wherever you're running `sqlite3` from.
The images in `Images.c4s.symbols` are PNGs (magic bytes `89 50 4E 47`).

## One-liner (find + extract in a single command)

Combining steps 2–4 for a known label:

```bash
sqlite3 "MyVocab - 2026-05-30.c4v" "
ATTACH DATABASE 'Images.c4s' AS imgs;
SELECT writefile(b.label || '.png', s.data)
FROM buttons b
JOIN symbol_links sl ON b.symbol_link_id = sl.id
JOIN imgs.symbols s ON lower(sl.rid) = lower(s.rid)
WHERE b.label LIKE '%hands on top%';
"
```
