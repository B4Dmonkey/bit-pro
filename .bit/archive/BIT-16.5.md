---
id: BIT-16.5
title: The four real skills load from the plugin
status: done
phase: 2
phase_label: The plugin ships the skills
---
## **Verse 2**

The four real skills move into the plugin and the throwaway probe comes out, so the plugin exposes
exactly the pipeline and nothing else. This is a content move, not code — there is no test to
drive it, which is why its checks are greps and a real session rather than a red-green cycle.

## Scope
- `bit/skills/scope/SKILL.md` — content from `assets/skills/bit_scope/SKILL.md`
- `bit/skills/plan/SKILL.md` — content from `assets/skills/bit_plan/SKILL.md`
- `bit/skills/do/SKILL.md` — content from `assets/skills/bit_do/SKILL.md`
- `bit/skills/check/SKILL.md` — content from `assets/skills/bit_check/SKILL.md`
- `bit/skills/ping/SKILL.md` — deleted; the spike is over and its probe was always throwaway.
- `assets/skills/` — **left alone.** Two copies exist on purpose from here until Verse 4, which is
  the verse that makes one of them impossible.

Exactly two edits per ported file, per the scope's "ports the skill text unchanged" decision:

1. The directory drops its `bit_` prefix — that is the path change above, not a content change.
   Plugin skills under `skills/` are namespaced `/<plugin>:<skill>`, so `bit_scope` would read
   `/bit:bit_scope`.
2. The one contract-pointer line changes from reading a path to running a command. Current
   locations: `bit_scope/SKILL.md:20`, `bit_plan/SKILL.md:10`, `bit_do/SKILL.md:15`,
   `bit_check/SKILL.md:12`. In each, **read `.claude/bit-cli.md`** becomes **run `bp
   instructions`**; the trailing parenthetical describing what the contract covers stays as
   written, because it differs per skill and is still accurate.

Nothing else changes — not the prose that names the skills `bit_scope`/`bit_plan`/`bit_do`, even
though invocation is about to become `/bit:scope`. Improving the skill text is separate work.

## References
- `https://code.claude.com/docs/en/plugins` — confirms `skills/` sits at the plugin root beside
  `.claude-plugin/`, how `/<plugin>:<skill>` namespacing derives from the directory name, and what
  `--plugin-dir` loads.
- `https://code.claude.com/docs/en/skills` — the supporting-files pattern, and that
  `${CLAUDE_SKILL_DIR}` resolves to the skill's own subdirectory rather than the plugin root. Read
  it only if a ported skill turns out to need a file beside it; none should.

## Claude verifies
- [ ] `grep -rn "bit-cli.md" bit/skills/` prints nothing — no ported skill still points at a path
- [ ] `grep -rln "bp instructions" bit/skills/` lists exactly the four `SKILL.md` files
- [ ] `diff assets/skills/bit_scope/SKILL.md bit/skills/scope/SKILL.md` shows exactly one changed
      line, and the same for plan, do, and check — proof the port was a move, not a rewrite
- [ ] `git status` shows the four new files and `bit/skills/ping/SKILL.md` deleted
- [ ] `claude plugin validate ./bit` exits 0; the missing-`author` warning is expected and is not
      a failure
- [ ] `just test` and `just lint` stay green — no Go changed, but the port must not have broken
      the build
- [ ] `just install`, so the `bp` on PATH has the `instructions` command the skills now call

## User verifies
- [ ] Start `claude --plugin-dir ./bit` in this repo. `/bit:scope`, `/bit:plan`, `/bit:do`, and
      `/bit:check` are all offered, and `/bit:ping` is gone.
- [ ] Invoke `/bit:scope` on something small. It runs `bp instructions` to learn the contract —
      it does not try to read `.claude/bit-cli.md`.
- [ ] Edit one identifiable word in `bit/skills/do/SKILL.md`, run `/reload-plugins`, then invoke
      `/bit:do`: the edited word is there, with no `just install` and no `bp init`. This is the
      authoring loop the Why promises, and this bar is where it first exists.
- [ ] Whole slice: the pipeline you actually use now comes from the plugin. Note that this repo's
      `.claude/skills/bit_*` still exist and still answer to `/bit_scope`, so invoke the
      **namespaced** commands — otherwise you are testing the old copy and learning nothing. Verse
      4 removes that ambiguity.

## Commit (user)
`feat(plugin): ship the four bit skills through the plugin`