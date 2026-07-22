---
id: BIT-7.1
title: just install puts bit on your bin dir
status: done
phase: 1
phase_label: bit runs anywhere
---
Adds a `just install` recipe that builds `bit` into the user's Go bin dir so it runs from any directory — the walking skeleton for "bit runs anywhere." No Go test here: this is a build-tooling recipe, not application logic, so the check is running it and confirming the binary lands and runs (a unit test would exercise `just`/`go`, not our code — YAGNI).

**Scope:**
- `Justfile` — add an `install` target.
- `README.md` — add a short Install note (run `just install`; `$(go env GOPATH)/bin` must be on `PATH`).

**Implement (GREEN):**
- [ ] Add an `install` recipe. Reuse the same ldflags version injection as `build:`, output named `bit` into the Go bin dir. Use a shell body so the dir is computed right:
  `dir="$(go env GOBIN)"; [ -n "$dir" ] || dir="$(go env GOPATH)/bin"; go build -ldflags="-X 'github.com/B4Dmonkey/bit-pro/cmd.version={{version}}'" -o "$dir/bit" .`
  Do **not** use `go install .` — it names the binary `bit-pro` (the module's final path element), not `bit`.
- [ ] README: add an Install section — `just install`, then ensure the Go bin dir is on `PATH`.

**Claude verifies:**
- [ ] `just install` succeeds; the resulting `"$dir/bit" --help` runs and prints usage.
- [ ] `just build`, `just test`, `just lint` still green.

**User verifies:**
- [ ] From a directory outside this repo, `bit --help` works (confirms your Go bin dir is on `PATH`).

**Commit (user):** `feat(install): add just install target to put bit on PATH`