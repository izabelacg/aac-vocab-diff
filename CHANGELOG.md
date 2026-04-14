# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.1.0 (2026-04-14)


### ✈️ Features

- **CLI diff** — Compare two TouchChat `.ce` vocabulary files and print a summary to the terminal.
- **HTML report (CLI)** — Optional `--report <file>` to write the same report as static HTML.
- **Flexible `--report` placement** — Flag may appear before or after the two file paths.
- **Web UI (`serve`)** — HTTP server with an upload form; diff result is returned as HTML in the browser. Uploaded files are written to temporary storage for the request only and removed afterward.
- **Diff coverage** — Detects added/removed pages; added/removed buttons; and per-button changes to label, spoken message, alternate pronunciation, visibility, and actions (e.g. navigation, media, external app).
- **Operational endpoints** — `GET /health` (JSON with build metadata when stamped at link time) and `GET /metrics` (plain-text counters for page views and diff success/error counts).
- **Optional analytics log** — `serve --analytics-log <file>` appends timestamped events (`page_view`, `diff_ok`, `diff_err`) with no personal data; survives process restarts (unlike in-memory `/metrics`).
- **Deployment-oriented guardrails (serve)** — 80 MB upload cap, ZIP validation before processing, per-IP rate limit on `POST /diff` (one request per five seconds, burst of three), and internal errors logged server-side only.


### Known limitations

- **Not compared** — Button position, colour, and images are not part of the diff.
- **Ignored rows** — Empty placeholder cells (no label and no message) are skipped.
- **Vocab coverage** — Primarily exercised with **TouchChat WordPower 60 Basic SS**; other vocabularies should work but are less validated.
- **In-memory metrics** — `/metrics` counters reset when the server process restarts; use the analytics log (or external scraping) for totals across deploys.
- **`.ce` files** — Read-only; the tool does not modify vocabulary files.

[0.1.0]: https://github.com/izabelacg/aac-vocab-diff/releases/tag/v0.1.0
