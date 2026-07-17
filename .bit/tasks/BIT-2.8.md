---
id: BIT-2.8
title: 'Refactor: consolidate frontmatter parsing'
status: done
phase: 2
phase_label: List & read
---
## Step 8 (Phase 2 — List & read) — Refactor: consolidate frontmatter parsing
**Status:** ✅ Done — verified 2026-07-15
**Not test-driven** — no new behavior, tests stay green throughout. By this point three
places touch the same file format: `task_create.go` writes it, `task_list.go` parses it
partially, `task_read.go` parses it fully. Phase 3's `update` is about to become a fourth
touchpoint that both parses *and* rewrites — worth consolidating before that duplication
compounds.

**Scope:**
- `cmd/task_model.go` — new: shared `Task` struct + `loadTask`/`(*Task).save` /
  `parseTask` helpers
- `cmd/task_create.go`, `cmd/task_list.go`, `cmd/task_read.go` — switch to the shared
  helpers, removing their local ad-hoc parsing/writing

**What to look for / criteria:** the same three-line pattern — read file, split on
`---`, `yaml.Unmarshal` — appearing independently in `task_list.go` and `task_read.go`,
plus `task_create.go`'s hand-assembled frontmatter string. If a step's implementation
diverged enough that consolidating isn't a clean lift (e.g. list's partial-parse struct
genuinely needs different fields), keep them separate rather than forcing a bad
abstraction — but expect it to be clean here since both already unmarshal the same
`id`/`title`/`status` shape.

**Implementation:**
- [x] `cmd/task_model.go`:
  ```go
  package cmd

  type Task struct {
      ID     string `yaml:"id"`
      Title  string `yaml:"title"`
      Status string `yaml:"status"`
      Body   string `yaml:"-"`
  }

  func taskPath(id string) string {
      return filepath.Join(tasksDir, id+".md")
  }

  func loadTask(id string) (*Task, error) {
      data, err := os.ReadFile(taskPath(id))
      if err != nil {
          return nil, fmt.Errorf("loading task %s: %w", id, err)
      }
      return parseTask(data)
  }

  func parseTask(data []byte) (*Task, error) {
      s := string(data)
      if !strings.HasPrefix(s, "---\n") {
          return nil, fmt.Errorf("task file missing frontmatter delimiter")
      }
      rest := s[len("---\n"):]
      idx := strings.Index(rest, "\n---\n")
      if idx == -1 {
          return nil, fmt.Errorf("task file missing closing frontmatter delimiter")
      }
      var t Task
      if err := yaml.Unmarshal([]byte(rest[:idx+1]), &t); err != nil {
          return nil, fmt.Errorf("parsing task frontmatter: %w", err)
      }
      t.Body = rest[idx+len("\n---\n"):]
      return &t, nil
  }

  func (t *Task) save() error {
      header, err := yaml.Marshal(t)
      if err != nil {
          return fmt.Errorf("marshaling task %s: %w", t.ID, err)
      }
      content := "---\n" + string(header) + "---\n" + t.Body
      if err := os.WriteFile(taskPath(t.ID), []byte(content), 0o644); err != nil {
          return fmt.Errorf("writing task %s: %w", t.ID, err)
      }
      return nil
  }
  ```
- [x] `task_create.go`'s `RunE` builds a `Task{...}` and calls `.save()` instead of
  hand-writing frontmatter.
- [x] `task_list.go`'s `RunE` calls `loadTask` (or `parseTask` on each glob match)
  instead of its own local unmarshal.
- [x] `task_read.go`'s `RunE` calls `loadTask(args[0])` instead of its own parsing.

**Claude verifies:**
- [x] `go test ./...` passes — every existing test from Steps 4–7 stays green
- [x] `go vet ./...` passes

**User verifies:**
- [x] none — internal refactor, no behavior change

**Commit (user):** `refactor(task-crud): consolidate frontmatter parsing into a shared Task type`