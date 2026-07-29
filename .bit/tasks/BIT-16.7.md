---
id: BIT-16.7
title: An existing settings file survives the wiring
status: todo
phase: 3
phase_label: Init keeps it current
---
## **Verse 3**

A repo that already has a `.claude/settings.json` contradicts the previous bar's fixed document —
writing ours would delete theirs. That is what forces a real merge, and it matters here more than
usual because this repo's own settings file already enables six `go-skills` plugins.

## Scope
- `claude/settings.go`
- `claude/settings_test.go`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestWriteSettings_MergesIntoExisting`
     - **Behavior:** the wiring is additive — unrelated settings and other people's plugins survive
       it — so init is safe to re-run in a configured repo rather than something you think twice
       about.
     - **Setup:** write `{"model": "opus", "enabledPlugins": {"go@go-skills": true}}` to the path
       first, then `WriteSettings(path)`.
     - **Assertions:** `model` is still `"opus"`; `enabledPlugins` has two entries, both `true`,
       one of them `go@go-skills`; `extraKnownMarketplaces["bit-pro"]` is present.
     - **Boundary:** `enabledPlugins` count == 1 going in and 2 coming out — one above the empty
       case the previous bar covered, which is the smallest input that can detect a clobber.
   - [ ] Confirm fails: `model` is gone and `go@go-skills` is gone, because the fixed document
     replaced the file wholesale.

2. **Implement (GREEN):**
   - [ ] Decode the top level into `map[string]json.RawMessage`, so every key we do not touch is
     re-emitted as the exact bytes it came in as. Decode only `extraKnownMarketplaces` and
     `enabledPlugins` further — also into `map[string]json.RawMessage` — set our two entries, and
     re-marshal with `json.MarshalIndent(m, "", "  ")` plus a trailing newline to match the
     existing file's formatting.
   - [ ] Known and accepted consequence: `encoding/json` emits map keys sorted, so the two keys we
     rewrite come out alphabetised and the top-level key order changes. This repo's checked-in
     settings file will reorder once, on the first init after this lands. That is a one-time diff,
     not a bug — do not build ordering machinery to avoid it.

3. **More tests (RED → GREEN):**
   - [ ] `TestWriteSettings_RejectsUnparseableFile`
     - **Behavior:** a settings file that is not valid JSON is reported and left exactly as it was,
       rather than silently replaced. Someone else's broken config is not ours to overwrite, and a
       half-edited settings file is a normal thing to walk into.
     - **Setup:** write `{ not json` to the path, then `WriteSettings(path)`.
     - **Assertions:** the error is non-nil and names the settings path; the file's bytes are
       unchanged.
     - **Boundary:** the invalid end of the file-contents range — the one input where writing is
       the wrong move.
   - [ ] `TestWriteSettings_IsIdempotent`
     - **Behavior:** two runs produce identical bytes, so re-running init to pull a skill fix
       leaves nothing spurious to commit. This is the scope's idempotence decision, made checkable.
     - **Setup:** `WriteSettings(path)` twice against a path that starts absent, capturing the
       bytes after each.
     - **Assertions:** the two byte slices are equal.
     - **Boundary:** run count 1 vs 2 — the repeat case, which is the workflow the scope commits to
       as the supported way to update.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(claude): merge the plugin wiring instead of overwriting settings`