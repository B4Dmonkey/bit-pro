---
id: BIT-44.3
title: research_read returns a topic's body
status: done
approved: true
phase: 1
phase_label: Analyze
---
## **Verse 1**

Agents have to read research back through the MCP, never from `.bit/` directly. Nothing can read what bar 1 writes, which is what forces `research_read`.

## Scope
- `task/research.go`: `(s *Store) ReadResearch(track, topic) (string, error)`.
- `cmd/serve_mcp.go`: the `researchReadTool = "research_read"` const, a description, `researchReadInput{Track string; Topic string (omitempty)}`, `researchReadOutput{Topics []string \`json:"topics,omitempty"\`; Body string \`json:"body,omitempty"\`}`, a `researchReadHandler`, and registration.
- `cmd/serve_mcp_research_test.go`: tests.

## References
- https://platform.claude.com/docs/en/agents-and-tools/tool-use/memory-tool: the `view` command (a directory lists its files, a file returns its contents) is the shape this tool follows.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_ResearchReadReturnsATopic`
     - **Behavior:** an agent reads back the body of a topic it (or an earlier pass) wrote.
     - **Setup:** seed the track; `research_write` topic `store` with body `"Store.Load normalizes IDs before reading."`; then call `research_read` with `track: testTrackID`, `topic: "store"`.
     - **Assertions:** `body` equals the written body.
     - **Boundary:** a topic is given (read mode).
   - [ ] Confirm it fails: the call errors because the tool `research_read` doesn't exist.

2. **Implement (GREEN):**
   - [ ] `ReadResearch`: `NormalizeID`; `trackExists`, otherwise `track %s does not exist`; `os.ReadFile(s.researchPath(track, topic))`, wrapping the error as `reading research topic %s for %s: %w`.
   - [ ] The handler returns `researchReadOutput{Body: body}` and wraps the error. Register it.
   - [ ] Description: with a topic it returns that topic's body; say reading the `index` topic first is the convention.

3. **More tests (expect them to pass on arrival):**
   - [ ] `TestServeMCPCmd_ResearchReadRefusesAMissingTopic`
     - **Behavior:** asking for a topic that was never written is an error, not an empty body.
     - **Setup:** the seeded track, no writes; read topic `"nope"`.
     - **Assertions:** `IsError` is true.
     - **Boundary:** the topic file is absent.
   - [ ] `TestServeMCPCmd_ResearchReadRefusesAnUnknownTrack`
     - **Behavior:** reads key to a real track the same way writes do.
     - **Setup:** read `track: testUnknownTrackID`, `topic: "index"`.
     - **Assertions:** `IsError` is true.
     - **Boundary:** the track ID isn't in tasks/, completed/, or archive/.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- none (deterministic)

## Commit (user)
`feat(mcp): research_read returns a research topic`