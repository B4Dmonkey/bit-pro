---
id: BIT-44.4
title: 'Omitted topic contradicts read-only: list the track''s topics'
status: todo
approved: true
phase: 1
phase_label: Analyze
---
## **Verse 1**

Progressive disclosure means an agent sees which topics exist before loading any. Bar 3's handler always reads a file, so a call with no topic can't return a listing. That contradiction forces the list branch.

## Scope
- `task/research.go`: `(s *Store) ResearchTopics(track) ([]string, error)`.
- `cmd/serve_mcp.go`: `researchReadHandler` branches on `in.Topic == ""`; update the description.
- `cmd/serve_mcp_research_test.go`: tests.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_ResearchReadListsTopics`
     - **Behavior:** with only a track, `research_read` returns that track's topic names, so an agent can choose what to open.
     - **Setup:** seed the track; `research_write` topics `index`, `store`, and `mcp`; then call `research_read` with only `track: testTrackID`.
     - **Assertions:** `topics` equals `["index", "mcp", "store"]` (sorted, without `.md`), and `body` is absent or empty.
     - **Boundary:** the topic is omitted, with N = 3 topic files.
   - [ ] Confirm it fails: an error from reading a default-named file, or `topics` missing.

2. **Implement (GREEN):**
   - [ ] `ResearchTopics`: `NormalizeID`; `trackExists`, otherwise an error; `os.ReadDir(s.researchDir(track))` (already sorted by name); for each regular file ending in `.md`, append `strings.TrimSuffix(name, ".md")`.
   - [ ] Handler: if `in.Topic == ""`, return `researchReadOutput{Topics: topics}`; otherwise bar 3's path.
   - [ ] Description: add that omitting the topic lists the track's topic names.

3. **More tests (RED → GREEN):**
   - [ ] `TestServeMCPCmd_ResearchReadListsNothingForATrackWithNoResearch`
     - **Behavior:** a track with no research yet lists nothing instead of erroring, so the first call of an analyze or scope pass is safe.
     - **Setup:** seed the track, no writes; call `research_read` with only the track.
     - **Assertions:** `IsError` is false, and `topics` is absent or has length 0.
     - **Boundary:** the research folder is absent (topic count 0).
   - [ ] Confirm it fails: `ReadDir` returns `ErrNotExist`. Fix it with `if errors.Is(err, fs.ErrNotExist) { return []string{}, nil }`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install`, so the `bp serve mcp` that Claude Code launches serves the new tools.

## User verifies
- [ ] Open a fresh Claude Code session in this repo and ask for `mcp__bit__research_read` with track `BIT-44`. It returns an empty topic list, not an error, and `mcp__bit__research_write` shows up in `/mcp` tools for the `bit` server.

## Commit (user)
`feat(mcp): research_read lists a track's research topics`