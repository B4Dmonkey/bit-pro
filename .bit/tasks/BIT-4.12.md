---
id: BIT-4.12
title: Custom delegate renders tracks, bars, verses
status: done
phase: 3
phase_label: see the shape
---
## Step 12 (Phase 3 — see the shape) — Custom delegate renders tracks, bars, verses
**Status:** ✅ Done — verified 2026-07-18 (committed `9442a7d`)
The visual grouping: tracks distinct from indented bars, each bar showing its verse. A
custom `list.ItemDelegate` using the Step 11 helpers. Rendering is visual → manual verify.

> Kept, but not settled. The delegate went single-line (denser than the default two-line
> title+description) which dropped the per-row status field, and the track/bar/verse styling
> is a first cut. The user likes it for now but isn't sure it's what they had in mind —
> logged in the README's "Cleanup & known issues" as something to possibly revisit.

**Scope:**
- `tui/delegate.go` — new: a custom delegate whose `Render` indents bars, distinguishes
  tracks, and shows the verse column via `isBar`/`verse` and Lip Gloss styles.
- `tui/model.go` — `New` uses the custom delegate instead of the default.

**Implement (GREEN):**
- [x] Custom delegate; swap it in.

**Claude verifies:**
- [x] `just build`, `just test` green, `just lint` clean

**User verifies:**
- [x] against the real ~29 records: tracks read as distinct from their bars (indent /
  styling), and each bar shows its verse column
- [x] the tree stays legible while navigating — you can tell which track a bar belongs to
  without opening it

**Commit (user):** `feat(tui): render tracks, bars, and verses in the list`