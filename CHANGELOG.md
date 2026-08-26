# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/izabelacg/aac-vocab-diff/compare/v0.1.0...v0.2.0) (2026-08-26)


### Features

* **diff,report:** show nav breadcrumbs in HTML page card header ([21e5086](https://github.com/izabelacg/aac-vocab-diff/commit/21e508643f90615512abd4756fa73e2364b1c2d0))
* **diff:** add ModifierKey, ModifierChange, ModifierSetDiff types ([a48648b](https://github.com/izabelacg/aac-vocab-diff/commit/a48648bd431eee767120324b2232657ba6e253f4))
* **diff:** add PageNavGraph, LoadPageNavGraph, AllShortestPathsFromHome ([fb6c438](https://github.com/izabelacg/aac-vocab-diff/commit/fb6c438f82dd13dc3d3223c7d8a6d324e6545af9))
* **diff:** compute diff for word-form modifier buttons ([7261283](https://github.com/izabelacg/aac-vocab-diff/commit/726128389ecadeb337e8161175c7a029b5db4f1f))
* **diff:** expand nav graph to cover action codes 8, 73, and button_set cells ([f75c4c1](https://github.com/izabelacg/aac-vocab-diff/commit/f75c4c1c5ebaa21d54e167d3471c3a9ef9c31ee6))
* **diff:** load word-form modifier buttons from button_set_modifiers ([02e488d](https://github.com/izabelacg/aac-vocab-diff/commit/02e488d40562fc8d9df63797c95f1afc747c065c))
* **diff:** populate NavPathFromOld/NavPathFromNew in CompareFiles ([9c0c459](https://github.com/izabelacg/aac-vocab-diff/commit/9c0c4593a4f3c4ab33290f02d55bab326eeaf035))
* **diff:** prefer labeled buttons in nav graph, use message as label fallback ([939e18b](https://github.com/izabelacg/aac-vocab-diff/commit/939e18b9eb3b2bb0489b4634ee31b89a5d03738f))
* **report:** improve word-form diff display with page context and grouping ([0dbc982](https://github.com/izabelacg/aac-vocab-diff/commit/0dbc9829eb9376cc0be06d76b8b38eb688abdb47))
* **report:** print word-form modifier changes in terminal output ([eba7222](https://github.com/izabelacg/aac-vocab-diff/commit/eba7222c99a932826614f6da0595f206be27e4b1))
* **report:** render nav breadcrumb in HTML page card header ([c70e67c](https://github.com/izabelacg/aac-vocab-diff/commit/c70e67c802f3249f569211f9575f2b0d77d8dd82))
* **report:** show word-form modifier changes in HTML report ([ff84fe4](https://github.com/izabelacg/aac-vocab-diff/commit/ff84fe4081021b759e63a11d729844d17f7d2e6f))


### Bug Fixes

* **diff:** disambiguate duplicate 'navigate to page' action labels ([06280b9](https://github.com/izabelacg/aac-vocab-diff/commit/06280b95bfa38af88f5b7a4cf8edafd32d9aa601))
* **diff:** key word-form modifiers by resources.rid to prevent name-collision false diffs ([d3fa28d](https://github.com/izabelacg/aac-vocab-diff/commit/d3fa28da823a287da4b42207c57b4e77b3ff6420))
* **report:** show up to 10 page names in word-form breadcrumbs ([2fd79fb](https://github.com/izabelacg/aac-vocab-diff/commit/2fd79fb85c8a6db243378b93d5d8d680cdf1a89f))

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
