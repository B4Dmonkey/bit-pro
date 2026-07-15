# Agent Instructions

## Critical: Go skills during bit_plan and bit_do

This is a Go project. Whenever `bit_plan` (planning) or `bit_do` (implementing) is run,
consult the relevant Go skills as part of the process itself — not as a separate
after-the-fact review pass:

- **go** — idiomatic Go for any `.go` file: package design, error handling, interfaces,
  concurrency, testing patterns. Applies to every step that writes or edits Go code.
- **cobra-viper** — whenever a step adds or changes CLI commands, subcommands, flags, or
  configuration binding.
- **fileflow-pathologize** — whenever a step moves, copies, renames, or generates files
  on disk, or builds a path from untrusted input (task IDs, CLI args, anything not
  hardcoded).
- **wails** — if a step touches a desktop/webview frontend-to-Go bridge.
- **go-spec-reviewer** — if a plan step's detail reads more like a spec than settled
  code, e.g. introducing a new subsystem.
- **go-release** — only for steps touching `go.mod` versioning, tags, or exported API
  surface of a published module.

Not every step needs every skill — apply the ones the step's scope files actually call
for. The point is to write plans and code that already follow these conventions, so
review doesn't have to catch it after the fact.
