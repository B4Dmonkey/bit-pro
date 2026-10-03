---
id: BIT-46.16
title: A research write appends the caller's commit only when HEAD has moved
status: done
approved: true
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Research topics carry a `commits` list. Its first entry is HEAD when the record is created, and a later write appends an entry when HEAD has moved. The store takes the commit from its caller (empty in this track, filled by BIT-47 and BIT-49). Two writes at the same SHA versus different SHAs can't both pass without the real rule.

## Scope
- `task/record.go`: unexported `appendCommit(list []Commit, head Commit) []Commit`. If `head.SHA == ""`, return `list` unchanged. If `list` is empty or `list[len(list)-1].SHA != head.SHA`, append `head` with `At` normalized to UTC, truncated to seconds. Otherwise return `list` unchanged. **Seam for BIT-49:** its retro store (also in `task/`) reuses it, and so does feedback creation in the next bar.
- `task/research.go`: `WriteResearch(track, topic, body string, head Commit) (string, error)`, where `commits = appendCommit(existing, head)`.
- callers pass `task.Commit{}`: `cmd/serve_mcp.go:474` (research_write handler) and the tests.
- **Seam (exported):** `(*Store).WriteResearch(track, topic, body string, head task.Commit) (string, error)`, `task.Commit{SHA, Branch string; At time.Time}`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreWriteResearch/records the creating commit` (subtest in `TestStoreWriteResearch`, created in BIT-46.15)
     - **Behavior:** a new topic remembers the commit it was written at.
     - **Setup:** store with track `BIT-46`; `WriteResearch("BIT-46", "index", "x", Commit{SHA: "aaa111", Branch: "v2", At: <2026-10-02T10:00:00-04:00>})`.
     - **Assertions:** `.json` `commits` == `[{"sha":"aaa111","branch":"v2","at":"2026-10-02T14:00:00Z"}]`.
     - **Boundary:** an empty list grows to 1.
   - [ ] Confirm fails: `WriteResearch` takes 3 arguments (compile error at the call), then `commits` is `[]`.

2. **Implement (GREEN):**
   - [ ] The new parameter; `appendCommit`.

3. **More tests (RED → GREEN):** (`TestStoreWriteResearch/commit history`, a subtest holding a table; each row is a second write after the one above, except the last)
   - [ ] row `same sha keeps one entry`: same SHA `aaa111` → still 1 entry. *Boundary:* HEAD unmoved.
   - [ ] row `same sha new branch keeps one entry`: SHA `aaa111`, branch `main` → still 1 entry, still branch `v2`. *Boundary:* only the SHA decides whether HEAD moved; a branch change at the same commit adds nothing.
   - [ ] row `moved sha appends`: SHA `bbb222` → 2 entries, in order. *Boundary:* HEAD moved.
   - [ ] row `empty commit leaves the history`: empty `Commit{}` → unchanged (1 entry), body still rewritten. *Boundary:* no git (e.g. `acme/`), which today is every call.
   - [ ] row `first write without a commit`: first write with `Commit{}` → `"commits": []`, not `null`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): research records keep a commit history from caller-supplied HEAD`