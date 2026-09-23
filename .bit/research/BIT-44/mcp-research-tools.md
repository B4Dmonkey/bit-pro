# mcp-research-tools (Verse 1, Go side)

**Checked:** the Verse 1 Decisions on `research_write`/`research_read` against the code, and whether `task/feedback.go` is the pattern they follow.

## Where it lives
- `cmd/serve_mcp.go`:
  - tool name consts :26-27
  - descriptions :100-114
  - in/out structs :178-196
  - registration in `runMCPServer` :257-264
  - `researchWriteHandler` :466-481
  - `researchReadHandler` :483-507 (an empty topic calls `ResearchTopics`; otherwise it calls `ReadResearch`)
- `task/research.go`:
  - `researchDir` :16
  - `researchPath` :20
  - `validateTopic` :24
  - `WriteResearch` :37
  - `ReadResearch` :59
  - `ResearchTopics` :77
- Tests: only `cmd/serve_mcp_research_test.go` (it uses the MCP harness). There is no `task/research_test.go`.
- There is no `bp research` CLI, and `cmd/init.go` is unchanged. Both are **confirmed**.

## Decision checks
- **Track must exist (active/completed/archived): confirmed.** The research code reuses `trackExists` from `task/feedback.go:34-42`, which checks `Path`, `completedPath` and `archivePath`.
- **Overwrite, no delete, dir created on first write: confirmed.** `os.MkdirAll` runs only after the track and topic checks pass (:47). The write is `os.WriteFile` (:52).
- **"Topic names have no required format; bp makes any name safe instead of refusing it": CORRECTED.** `validateTopic` **refuses** four kinds of topic:
  - empty
  - dots and spaces only
  - any `..` (including `a..b`)
  - any `/` or `\`

  Everything else goes through `pathologize.Clean`, then `strings.TrimLeft(..., ".")`, then `+ ".md"`. `bit/skills/analyze/SKILL.md` already documents the refusal. The track's Decision text is what's stale.
- **Feedback as the pattern: partly true.** Feedback file names are always `<TRACK>-NNN.md`, built with `pathologize.Join` (`task/feedback.go:18-19`), so feedback never sanitizes free text. The only thing research copies from it is `pathologize.Join` + `trackExists`.

## Bugs / surprises
1. **The track ID can escape `.bit/research/` (CONFIRMED by reading).**
   - `researchDir` is `filepath.Join(s.root, "research", track)` with the raw track, and `NormalizeID` only uppercases it.
   - `trackExists` goes through `Path`, which is `pathologize.Join`. pathologize v1.1.0's `Join` doc says it drops `..` segments.
   - So with `track="../../BIT-44"`, the existence check looks at `.bit/tasks/BIT-44.md`, which exists, so it passes.
   - The write then goes to `.bit/research/../../BIT-44/<topic>.md`, which is `<project>/BIT-44/`, outside `.bit/`.
   - Reads and lists escape the same way.
   - Fix: build `researchDir` with `pathologize.Join`, or validate the track. Not reproduced with a test yet.
2. **A bar ID is accepted as a track.** Bars live in `.bit/tasks/` too, so `track=BIT-44.1` passes `trackExists` and creates `.bit/research/BIT-44.1/`. The tool descriptions say a track has no dot, but nothing enforces it.
3. **Topics that are different strings can collide on one file and silently overwrite each other.**
   - `a:b` and `ab` both become `ab.md`.
   - `Store` and `store` are the same file on macOS.
   - `?` becomes pathologize's default name, `file.md`.
   - The topic `index.md` becomes `index.md.md`.
   - The listing returns the sanitized file name, not the original topic.
4. **An empty list is `{}`, not `{"topics":[]}`**, because `Topics` has `json:"topics,omitempty"` (serve_mcp.go:194). The behaviour matches the Decision ("empty, not an error"), but the key disappears.
5. **Listing an unknown track returns an error.** A missing topic returns a wrapped `os.ReadFile` error that includes the absolute path.

## Test gaps
None of these are tested:
- an archived track
- a completed track on read or list
- an unknown track in list mode
- a lowercase track ID
- a bar ID used as the track
- track traversal (bug 1)
- topic collisions
- a folder with non-`.md` entries

## Status
BIT-44.1 to .4 are done. BIT-44.5 (the analyze skill) is `doing`. `bit/skills/analyze/` is **untracked** and not committed. The plugin picks skills up from `bit/skills/` automatically, so no manifest edit is needed. Per memory, the plugin installs from GitHub, so analyze won't reach other projects until it's pushed.
