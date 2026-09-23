# scope-skill (Verse 2)

**Checked:** that `bit/skills/scope/SKILL.md` is the file to edit so scope reads and cites research, and the Why's claim that "bit:plan says outright that it isn't a discovery phase".

## What the code does
- The file ships: `.claude-plugin/marketplace.json` has one plugin, `bit`, with source `./bit`. `bit/.claude-plugin/plugin.json` doesn't list skills or agents; it picks them up from the `bit/skills/` and `bit/agents/` folders. The only other scope copies aren't shipped and don't need editing: a stale worktree at `.claude/worktrees/agent-a6848e37a1646d392/` and the user-level `~/.claude/skills/bit-scope/` (the old pipeline).
- No skill or agent under `bit/` has `allowed-tools:`/`tools:` frontmatter. Tools are named only in prose, so adding `mcp__bit__research_read` means editing prose (line 20 lists task_create/task_read/task_update).
- The research guidance is at scope SKILL.md:116-121, under "Gathering context (new scopes)": "Then do *light* research ... but not a deep dive (that's bit_plan's job)". "Don't over-research. If you find yourself reading function bodies..."
- scope SKILL.md:100 says "**bit_plan is not a discovery phase.**", which contradicts line 116 in the same file.
- plan SKILL.md:97 says "Then research the codebase (a scope's light "touches" pointers are a starting point, not a substitute — go deeper here)" and :107 says "Don't start drafting until you've done this research." The plan skill never uses the word "discovery".
- Body template (scope :135-192): Why, Summary, Visual aid, Risks & unknowns, Decisions, Verses, References. "Touches:" is a line under each verse, not a section of its own.
- References (:186-191) is limited to "external artifacts the user explicitly pointed to as authorities". Research notes have no slot to be cited in.
- The refine checklist (:198-214) has no step that reads research.
- No skill or agent other than analyze mentions `.bit/research`, `research_read`, or analyze.
- Evals exist only for analyze and feedback. Scope has no evals dir.

## Verdict
- Touches pointer **confirmed**: `bit/skills/scope/SKILL.md` is the only file Verse 2 needs.
- The Why's claim is **corrected**. "Not a discovery phase" lives in the scope skill (:100). The plan skill tells itself to go deeper (:97). Scope :116 hands the deep dive to plan, and :100 says plan doesn't do discovery. The real gap is that the two skills contradict each other and nothing does deep research on purpose.
- Edits Verse 2 needs:
  - Line 20 and :116-121: list topics, read `index` first, and replace "(that's bit_plan's job)" with a pointer to bit:analyze/research.
  - A citation slot for research notes: widen References or add a separate line.
  - The refine checklist.
- **Open, outside this track's verses:** should plan :97 also read research before its own deep pass? analyze/SKILL.md already lists bit_plan as a reader, but no verse touches `bit/skills/plan/SKILL.md`. This is a scope decision.
