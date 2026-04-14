# Vocab Diff

A tool for parents and caregivers of AAC users to compare two TouchChat™ `.ce`
vocabulary files and see exactly what changed — pages added or removed, buttons
added, removed, or modified — without having to manually browse every page in
the app.

## How it works

A `.ce` file is a ZIP archive containing a `.c4v` file, which is a SQLite
database. The tool extracts that database, queries it, and diffs the results
between two files. Nothing is written to disk permanently.

## Usage

### CLI mode

Compare two `.ce` files and print a summary to the terminal:

```bash
aac-vocab-diff old.ce new.ce
```

Generate an HTML report as well:

```bash
aac-vocab-diff old.ce new.ce --report report.html
```

### Web UI mode on a Raspberry Pi

Start a local web server (useful on a Raspberry Pi so anyone on the home
network — or the wider internet via a tunnel — can use it from their phone or
laptop):

```bash
aac-vocab-diff serve --addr :8888
```

Open `http://raspberrypi.local:8888` in any browser, upload two `.ce` files,
and the diff report is shown directly in the browser. Files are saved to a
temporary location on the server for the duration of the request and deleted
immediately after. When using a public tunnel, uploads travel encrypted to the
server; processing and any temporary storage happen on the server only, not in
any cloud service.

**Safety guardrails (relevant when the server is publicly reachable):**

- Uploads are capped at 80 MB; anything larger is rejected with a clear error.
- Uploaded files are validated as ZIP archives before processing; non-ZIP files
  are rejected with a friendly 400.
- Requests to `/diff` are rate-limited to one per 5 seconds per IP (burst of 3)
  to prevent accidental overload.
- Internal error details are never returned to the client; they are logged
  server-side only.

## Build from source

Requires Go 1.21+.

```bash
# Local binary
make build

# ARM64 binary for Raspberry Pi
make build-arm64
```

## Deploy to Raspberry Pi

```bash
make deploy-pi                          # uses PI_HOST=raspberrypi.local by default
```

This cross-compiles the ARM64 binary, copies it and the systemd unit file to
the Pi, and enables/restarts the service.

## What is compared

For each button, the tool tracks changes to:

- **Label** — the text shown on the button
- **Spoken message** — what the device says when the button is pressed
- **Alternate pronunciation** — phonetic override for text-to-speech
- **Visible** — whether the button is shown or hidden
- **Actions** — navigate to page, play a sound, open an app, etc.

Changes are grouped into three categories:

| Symbol | Meaning |
|--------|---------|
| `+` | Added |
| `-` | Removed |
| `~` | Modified |

## Notes

- Button position, colour, and images are not compared.
- Empty placeholder cells (no label, no message) are ignored.
- `.ce` files are never modified.
- Tested with **TouchChat™ WordPower 60 Basic SS**. Other vocabulary types
  should work — if yours produces an error, please
  [send feedback](https://docs.google.com/forms/d/11YlQYQ_YzGSYQaFJxZFT8TbPX6vj2AvW1ow5ai4CDOU/viewform).

## Disclaimer

This project is not affiliated with, endorsed by, or associated with Saltillo
Corporation or PRC-Saltillo in any way. TouchChat™ and WordPower™ are trademarks
of their respective owners. Users are responsible for ensuring their use of
this tool complies with their own software licence agreements.

## License

MIT
