---
id: BIT-33.2
title: Play-prompt yes calls enqueue stub
status: todo
approved: true
phase: 2
phase_label: Popup yes enqueues
---
## **Verse 2**

TUI test: pressing `"y"` at the play prompt calls no enqueue stub → fails (field doesn't exist) → add `enqueue` field + `WithEnqueue` + call in `handlePlayPrompt`. A contradicting test proves `"n"` must not call enqueue, ruling out "always call on close."

Wire in `cmd/tui.go`: resolve `project_id` from DB at startup, pass closure as `WithEnqueue`.

## Scope
- `tui/model.go` — add `enqueue func(targetID, targetTyp string) error` field; `WithEnqueue(fn func(string, string) error) Option`; in `handlePlayPrompt`: `case "y": m.playPromptOpen = false; if m.enqueue != nil { _ = m.enqueue(selectedTaskID, targetTyp) }`; determine `targetTyp` via `isBar(id)` → `"bar"` else `"track"`
- `cmd/tui.go` — `db.Open()` + `orm.New(sqlDB).GetProjectByPath(ctx, cwd)` → `project.ID`; if project not found (unregistered): proceed without setting `WithEnqueue` (no-op per Decision); pass `WithEnqueue(func(sid, kind string) error { return orm.New(sqlDB).EnqueueTask(ctx, orm.EnqueueTaskParams{ProjectID: project.ID, SubjectID: sid, SubjectKind: kind}) })`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestHandlePlayPrompt_Yes_CallsEnqueue`
     - **Behavior:** pressing `y` at the play prompt invokes the enqueue callback with the selected task's ID and kind
     - **Setup:** construct `model` with `playPromptOpen = true`; selected task ID `"BIT-33"` (a track — no dot); inject `WithEnqueue` stub that records `(targetID, targetTyp)` args
     - **Assertions:** after `m.Update(tea.KeyPressMsg{Code: tea.KeyRunes, Runes: []rune("y")})`: stub called once; `gotSubjectID == "BIT-33"`; `gotSubjectKind == "track"`
     - **Boundary:** `"y"` key — the one key in `handlePlayPrompt` that should trigger enqueue
   - [ ] Confirm fails: `m.enqueue undefined` (field doesn't exist yet)

2. **Implement (GREEN):**
   - [ ] Add `enqueue` field + `WithEnqueue` to `tui/model.go`
   - [ ] Call `m.enqueue(id, kind)` in the `"y"` branch of `handlePlayPrompt`

3. **More tests (RED → GREEN):**
   - [ ] `TestHandlePlayPrompt_No_SkipsEnqueue`
     - **Behavior:** pressing `n` closes the prompt without calling enqueue
     - **Setup:** same as above with enqueue stub
     - **Assertions:** stub not called; `playPromptOpen == false`
     - **Boundary:** `"n"` key — the other close path; contradicts any "call enqueue on every close" implementation

## Claude verifies
- [ ] `go test ./tui/...` passes
- [ ] `go test ./cmd/...` passes (wiring in tui cmd compiles and no regressions)
- [ ] `go build ./...` passes
- [ ] `just install`

## User verifies
- [ ] In `bp tui`, approve the last bar of any track so the play prompt opens; press `y`; run `sqlite3 ~/.local/share/bit-pro/bit.db "SELECT * FROM queue;"` and confirm a row exists with the correct `target_id`

## Commit (user)
`feat(tui): enqueue on play-prompt yes`