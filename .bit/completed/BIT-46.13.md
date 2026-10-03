---
id: BIT-46.13
title: Task records carry their project and created/updated timestamps
status: done
approved: true
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Every record carries `project`, `created_at` and `updated_at`. Timestamps are RFC 3339 UTC, `updated_at` changes on every write, and a rewrite keeps `created_at`. A save-then-update test can't pass with a single fixed timestamp, so it forces the real rule.

## Scope
- `task/task.go`: `Task` gains `Project string`, `CreatedAt time.Time`, `UpdatedAt time.Time`, all tagged `yaml:"-"`, so `Bytes()` and the v1 round trip are unaffected.
- `task/store.go`:
  - `Store` gains an unexported `now func() time.Time` (`NewProject` sets `time.Now`). Package tests override it.
  - `Save`: `t.Project = s.code`; `ts := s.now().UTC().Truncate(time.Second)`; if `t.CreatedAt.IsZero()`, set it to `ts`; always set `t.UpdatedAt = ts`. Every rewrite path (`Update`, `SetApproved`, order changes, `removeFromOrder`) already `Load`s first, so `created_at` carries through.
  - `Load` fills the three fields from the record.
- `task/record.go`: `taskRecord` gains `project`, `created_at`, `updated_at` (`time.Time` marshals as RFC 3339, e.g. `"2026-10-02T14:03:00Z"`).
- `TestStoreSave/writes a json record beside the body` (BIT-46.12) asserts the exact unmarshalled record: add `project`, `created_at` and `updated_at` to its expected value (give that store a fixed clock).
- **Seam for BIT-49:** migrate builds its store with `task.NewProject(dir, code)`, so records get the code, and its "migration time" timestamps are simply `now`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreSave/stamps project and timestamps` (subtest in `TestStoreSave`, created in BIT-46.12)
     - **Behavior:** a new record says which project wrote it and when.
     - **Setup:** `s := NewProject(dir, "BIT")`; `s.now = func() time.Time { return time.Date(2026, 10, 2, 14, 3, 0, 500, time.FixedZone("EDT", -4*3600)) }`; `s.Create(CreateParams{Title: "T"})`.
     - **Assertions:** the raw `.json` has `"project": "BIT"`, `"created_at": "2026-10-02T18:03:00Z"`, `"updated_at": "2026-10-02T18:03:00Z"`.
     - **Boundary:** a non-UTC clock with sub-second nanos, which proves the UTC conversion and the truncation.
   - [ ] Confirm fails: fields missing.

2. **Implement (GREEN):**
   - [ ] Fields, the clock, and `Save` stamping.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreUpdate/keeps created at and bumps updated at` (subtest): create at T1, set the clock to T2 = T1+1h, `Update(id, Patch{Body: &b})` → `created_at` == T1, `updated_at` == T2. *Boundary:* the second write; a body-only update still bumps.
   - [ ] the same for `SetApproved` and for the parent track when `Create` appends a bar to its `order` (the track's `updated_at` moves and its `created_at` doesn't): `TestStoreSetApproved/keeps created at and bumps updated at` (subtest of a new `TestStoreSetApproved`) and `TestStoreCreate/bumps the parent track updated at and keeps its created at` (subtest). *Boundary:* indirect rewrites.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `go test ./task -run 'Parse|Bytes'` passes unchanged

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): task records carry project and timestamps`