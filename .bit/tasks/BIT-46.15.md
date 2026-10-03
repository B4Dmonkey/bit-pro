---
id: BIT-46.15
title: A research topic is stored as a <topic>.json record beside its .md
status: done
approved: true
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Research topics get the same pair format as tasks: `research/<TRACK>/<topic>.json` beside the unchanged `<topic>.md`. They carry `project`, `track`, `topic`, timestamps and an empty `commits` list. A test reading the `.json` forces it. The `commits` append rule comes in the next bar.

## Scope
- `task/research.go`:
  - unexported `researchRecord`: `project, track, topic, created_at, updated_at, commits, content` (all written, with `commits` as `[]` here). `topic` is the sanitized file stem (`strings.TrimLeft(pathologize.Clean(topic), ".")`, as `researchPath` builds it today). `content` is `<stem>.md`.
  - new `type Commit struct { SHA string `json:"sha"`; Branch string `json:"branch"`; At time.Time `json:"at"` }` in `task/record.go`, the element type of `commits` (used empty here).
  - `WriteResearch`: write the `.md`, then the `.json`. On an existing topic, read its `.json` first and keep `created_at` and `commits`. Bump `updated_at` (`s.now`, UTC, seconds). Return the `.md` path, as today.
  - `ReadResearch`: read the `.json`, then read the body from `pathologize.Join(dir, content)`, as tasks do.
  - `ResearchTopics`: list `*.json` stems (`:95`), so a stray `.md` isn't a topic.
- Tests: `cmd/serve_mcp_research_test.go` (including `seedEscapedResearch`, `:341`; converted in BIT-46.9) seeds and asserts through the store, not raw `.md` files. Returned paths stay `.md`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreWriteResearch/writes a json record beside the body` (subtest of a new `TestStoreWriteResearch`, `task/research_test.go`, new file)
     - **Behavior:** a research note is stored as metadata plus an untouched markdown body, and the caller gets the body's path.
     - **Setup:** `s := NewProject(dir, "BIT")` with a fixed clock; a saved track `BIT-46`; `path, _ := s.WriteResearch("BIT-46", "claim-audit", "# Audit\n\n| a | b |\n")`.
     - **Assertions:** `path == research/BIT-46/claim-audit.md`; its bytes equal the body. `claim-audit.json` has `"project":"BIT","track":"BIT-46","topic":"claim-audit","commits":[],"content":"claim-audit.md"` and the two timestamps.
     - **Boundary:** a first write on a fresh topic.
   - [ ] Confirm fails: no `.json`.

2. **Implement (GREEN):**
   - [ ] `researchRecord`, `Commit`, the new write/read/list.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreWriteResearch/rewrite keeps created at` (subtest): write at T1, rewrite at T2 with a new body → `.md` holds the new body, `created_at` == T1, `updated_at` == T2. *Boundary:* the second write.
   - [ ] `TestStoreResearchTopics/lists records only` (subtest of a new `TestStoreResearchTopics`): a stray `research/BIT-46/orphan.md` plus two written topics → exactly the two names. *Boundary:* orphan body.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): store research topics as JSON metadata plus markdown`