---
id: BIT-7.4
title: Re-running init refreshes seeded files
status: done
phase: 2
phase_label: seed skills
---
Locks the scope's core promise — re-running `init` refreshes seeded files back to the embedded content — with a test that modifies a seeded skill then re-runs `init`. If `Seed` already overwrites (via `WriteFile`), this is a green-on-arrival contract guard; if a prior bar made it skip existing files, this goes red and forces overwrite semantics.

**Scope:**
- `cmd/init_test.go` — new test.
- `assets/assets.go` — only if the test goes red (make `Seed`'s write unconditional).

**TDD cycle:**

1. **Write test (RED or guard):**
   - [ ] `TestInitCmd_ReseedRefreshes`
     - **Behavior:** a second `init` restores a locally-modified seeded file to the embedded content.
     - **Setup:** `t.Chdir(t.TempDir())`; `mustRun("init","--prefix","BIT")`; overwrite `.claude/skills/bit_do/SKILL.md` with `"stale\n"`; `mustRun("init","--prefix","BIT")` again.
     - **Assertions:** `.claude/skills/bit_do/SKILL.md` bytes equal the embedded copy again.
     - **Boundary:** destination file already exists with different content — the overwrite/refresh case (vs. 7.2's create-from-nothing).
   - [ ] Confirm result: passes if `Seed` overwrites; if it fails, `Seed` was skipping existing files — change it to always write.

2. **Implement (GREEN):**
   - [ ] Only if red: make `Seed`'s file write unconditional (`os.WriteFile` truncates). No change if already green.

**Claude verifies:**
- [ ] `just test` green (all three seed tests).
- [ ] `just lint`, `just build` green.

**User verifies:**
- [ ] Comfortable that "re-run `init` to refresh" is the documented, tested refresh mechanism (version-aware refresh stays deferred).

**Commit (user):** `test(init): lock idempotent skill refresh on re-init`