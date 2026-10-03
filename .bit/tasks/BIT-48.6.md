---
id: BIT-48.6
title: The README documents the four claude commands as the manual setup and repair path
status: todo
approved: true
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

Repairing lost global wiring has no bp command; the operator re-runs the documented `claude` commands. This docs bar puts them in the README, matching `claude.GlobalWiring()` line for line. It's the verse's last bar, so it carries the end-to-end check.

## Scope
- `README.md`: a new `## Claude Code setup` section directly after `## Install`. Two sentences: `bp add` runs these once, the first time it registers a project, at user scope, so bit loads in every Claude session; if a step fails, or the wiring is removed later, run them by hand. Then a code block with the four lines exactly as `GlobalWiring()` prints them. Leave the Quickstart, the command table and the `bp init` mentions to BIT-49 Verse 3.
- `scripts/install.sh` is unchanged (it already runs step 1).

## Steps
- [ ] Add the section.

## Claude verifies
- [ ] each of the four lines appears verbatim in `README.md`: `grep -cF -e 'claude plugin marketplace add B4Dmonkey/bit-pro' -e 'claude plugin marketplace update bit-pro' -e 'claude plugin install bit@bit-pro --scope user' -e 'claude mcp add -s user bit -- bp serve mcp' README.md` reports 4
- [ ] `just lint` and `just test` pass

## User verifies
Whole verse, in a sandbox (never `just install`; set up as in BIT-46.11, but with this logging fake `claude`):
- [ ] Before exporting `HOME`: `SB=$(mktemp -d); mkdir -p $SB/home $SB/data $SB/bin $SB/a $SB/b`, then
  ```
  cat > $SB/bin/claude <<EOF
  #!/bin/sh
  echo "\$*" >> $SB/claude.log
  [ "\$2" = install ] && [ -n "\$FAIL_INSTALL" ] && exit 1
  exit 0
  EOF
  chmod +x $SB/bin/claude; just db-gen-queries && go build -o $SB/bin/bp .
  export HOME=$SB/home XDG_DATA_HOME=$SB/data PATH=$SB/bin:$PATH
  ```
- [ ] `cd $SB/a && printf 'aaa\n' | bp add .` prints `added AAA ...` then `Setting up bit in Claude Code (user scope)...` and exits 0. The first four lines of `$SB/claude.log` are the four README commands without the leading `claude` (later `plugin marketplace update bit-pro` lines come from bp's background refresh after every command; ignore them).
- [ ] `bp add .` again prints `already added`; `grep -c 'plugin install' $SB/claude.log` is still `1`.
- [ ] `echo '{"mcpServers":{"bit":{"command":"bp","args":["serve","mcp"]}}}' > $SB/home/.claude.json; rm $SB/claude.log; cd $SB/b && printf 'bbb\n' | bp add .` → `$SB/claude.log` has no `mcp add` line.
- [ ] `rm $SB/home/.claude.json; mkdir $SB/c && cd $SB/c && printf 'ccc\n' | FAIL_INSTALL=1 bp add .; echo $?` → prints `Run these to finish setting up bit in Claude Code:` with the four commands, `Error: claude wiring step 3 of 4: ...`, and `1`; `bp list` shows `CCC`.
- [ ] Whole slice: one `bp add` sets bit up for the machine, a re-run touches nothing, and a failure leaves the operator with the exact commands.

## Commit (user)
`docs(bit): document the user-scope Claude setup commands`