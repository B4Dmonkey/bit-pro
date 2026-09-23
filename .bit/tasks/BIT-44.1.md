---
id: BIT-44.1
title: research_write puts a topic note in the track's research folder
status: done
approved: true
phase: 1
phase_label: Analyze
---
## **Verse 1**

Walking skeleton: an agent can write a research note for a track through the MCP. It is forced by the missing tool, since nothing today can put a note under `.bit/research/`.

## Scope
- `task/research.go` (new): `researchSubdir = "research"`, `(s *Store) researchDir(track)`, `(s *Store) researchPath(track, topic)`, and `(s *Store) WriteResearch(track, topic, body) (string, error)`.
- `cmd/serve_mcp.go`: the `researchWriteTool = "research_write"` const, a description, `researchWriteInput{Track, Topic, Body}`, `researchWriteOutput{Path}`, a `researchWriteHandler`, and registration in `runMCPServer`.
- `cmd/serve_mcp_research_test.go` (new): tests. Put test string consts beside the existing ones (the `testFeedbackDir`/`testNoteBody` style in `serve_mcp_write_test.go`), and don't write comments.

## References
- `task/feedback.go` shows the pattern to mirror: `feedbackDir`, `notePath` via `pathologize.Join`, `MkdirAll(dirMode)`, `WriteFile(fileMode)`.
- The `fileflow-pathologize` skill covers how the topic becomes a single file name (untrusted input).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_ResearchWriteWritesATopic`
     - **Behavior:** an agent's research note lands in `.bit/research/<track>/<topic>.md` through the MCP, and the folder is created on the first write.
     - **Setup:** `t.TempDir()`; `seedTasks` with `{ID: testTrackID, Title: testTitle, Status: task.StatusDoing}`; no `.bit/research/` exists. Call `research_write` with `track: testTrackID`, `topic: "index"`, and `body: "## Findings\n\n- note paths are built with pathologize.Join in task/feedback.go\n- see [store](store.md)"`.
     - **Assertions:** the returned `path` has the suffix `filepath.Join("research", testTrackID, "index.md")`, and the file's contents equal the body byte for byte.
     - **Boundary:** the research folder is absent (the first write for this track), so it must be created rather than assumed.
   - [ ] Confirm it fails: the call errors because the tool `research_write` doesn't exist.

2. **Implement (GREEN):**
   - [ ] `WriteResearch`: `track = NormalizeID(track)`; `MkdirAll(s.researchDir(track), dirMode)`; `path := s.researchPath(track, topic)`, where `researchPath` is `pathologize.Join(s.researchDir(track), pathologize.Clean(topic)+".md")` so the topic is always exactly one segment; `os.WriteFile(path, body, fileMode)`; return `path`.
   - [ ] `researchDir(track)` = `filepath.Join(s.root, researchSubdir, track)`.
   - [ ] The handler calls `task.New(bitdir.ForRoot(root)).WriteResearch` and wraps the error as `writing research for %s: %w`. Register it with `mcp.AddTool`.
   - [ ] Description: say it writes one topic of a track's agent scratchpad under `.bit/research/<track>/`, that writing an existing topic replaces it, and that the topic named `index` is the conventional summary with links, which readers open first.

3. **More tests (these pin Decisions; expect them to pass on arrival):**
   - [ ] `TestServeMCPCmd_ResearchWriteOverwritesATopic`
     - **Behavior:** rewriting a topic replaces it instead of adding a second file.
     - **Setup:** the seeded track; write `index` with body A, then again with body B.
     - **Assertions:** the file holds exactly B, and `.bit/research/<track>/` contains one `.md` file.
     - **Boundary:** the same topic is written twice (the re-check path the analyze loop uses).
   - [ ] `TestServeMCPCmd_ResearchWriteKeepsAnUnsafeTopicInTheTrackFolder`
     - **Behavior:** a topic name can't escape the track's research folder.
     - **Setup:** the seeded track; write topic `"../../tasks/" + testTrackID` with any body.
     - **Assertions:** `filepath.Rel(<dir>/.bit/research/<track>, path)` has no `..` prefix and no separator, and `store.Load(testTrackID)` still returns `testTitle`.
     - **Boundary:** the topic contains traversal segments and separators (untrusted input at its worst).

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- none (deterministic)

## Commit (user)
`feat(mcp): research_write tool writes track research notes`