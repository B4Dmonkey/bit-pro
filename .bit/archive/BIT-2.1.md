---
id: BIT-2.1
title: '`bit init --prefix` writes `config.toml`'
status: done
phase: 1
phase_label: Init wizard + create
---
## Step 1 (Phase 1 — Init wizard + create) — `bit init --prefix` writes `config.toml`
**Status:** ✅ Done — verified 2026-07-15

Adds a `--prefix` flag to the existing `init` command and, when set, writes a
`config.toml` inside `.bit/`. This is the walking skeleton's first half — nothing
else in the scope has a prefix to build task IDs from until this exists. Forced by
nothing prior; it's the first new test.

**Scope:**
- `go.mod` / `go.sum` — add `github.com/BurntSushi/toml@v1.6.0`
- `cmd/config.go` — new: `Config` struct, `loadConfig()`, `saveConfig(*Config)`
- `cmd/init.go` — add `--prefix` flag; `RunE` writes config when it's set
- `cmd/init_test.go` — add new test (existing `TestInitCmd` untouched this step — it
  never sets `--prefix`, so it keeps exercising only `.bit/` creation)

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestInitCmd_WritesConfigWithPrefix` (in `cmd/init_test.go`)
     - **Behavior:** proves the non-interactive path of the init wizard — an LLM or
       scripted caller can set up a project in one command, no prompts.
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; `rootCmd := NewRootCmd()`;
       `rootCmd.SetArgs([]string{"init", "--prefix", "BIT"})`; `rootCmd.Execute()`.
     - **Assertions:** `err` is `nil`; `os.ReadFile(filepath.Join(dir, ".bit", "config.toml"))`
       succeeds and its content, decoded via `toml.Decode`, gives `Config{Prefix: "BIT"}`.
     - **Boundary:** `--prefix` supplied (non-empty) — the explicit-input path, contrasted
       with Step 2's flag-omitted path.
   - [x] Confirm fails: compile error — `undefined: Config` (`cmd/config.go` doesn't exist
     yet).

2. **Implement (GREEN):**
   - [x] `go get github.com/BurntSushi/toml@v1.6.0`
   - [x] `cmd/config.go`:
     ```go
     package cmd

     import (
         "fmt"
         "os"

         "github.com/BurntSushi/toml"
     )

     const configFileName = ".bit/config.toml"

     type Config struct {
         Prefix string `toml:"prefix"`
     }

     func loadConfig() (*Config, error) {
         var cfg Config
         if _, err := toml.DecodeFile(configFileName, &cfg); err != nil {
             return nil, fmt.Errorf("reading %s: %w", configFileName, err)
         }
         return &cfg, nil
     }

     func saveConfig(cfg *Config) error {
         f, err := os.Create(configFileName)
         if err != nil {
             return fmt.Errorf("creating %s: %w", configFileName, err)
         }
         defer f.Close()
         if err := toml.NewEncoder(f).Encode(cfg); err != nil {
             return fmt.Errorf("writing %s: %w", configFileName, err)
         }
         return nil
     }
     ```
   - [x] `cmd/init.go`: add `prefix string` var, `cmd.Flags().StringVar(&prefix, "prefix", "", "task ID prefix for this project (e.g. BIT)")`; in `RunE`, after `os.MkdirAll(".bit", 0o755)`, if `prefix != ""`, call `saveConfig(&Config{Prefix: prefix})`.

**Claude verifies:**
- [x] `just test` passes
- [x] `just build` succeeds; manually running the built binary confirms `.bit/config.toml`
  is written with the right content

**User verifies:**
- [x] `config.toml` belongs inside `.bit/`, not the project root (corrected from this
  step's original draft)

**Commit (user):** `feat(task-crud): init --prefix writes config.toml to .bit/`