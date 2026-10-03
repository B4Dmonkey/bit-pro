# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default".

## 2026-09-30
- **Cutover:** the operator alone decides when v2 is ready and merges the branch. No track gates cutover, and none of BIT-45, 46, 47, 48, 49 or 50 is a precondition for it.
- **The track split:** the original BIT-46 was split into three tracks:
  - BIT-46: registry, resolver, central store and the record format.
  - BIT-48: global Claude wiring and removing `bp init`. (2026-10-01: `bp init` is deleted in BIT-46 instead.)
  - BIT-49: shared feedback and retro, `bp migrate` and the `.bit/` sweep.
  Each track has its own `decisions` topic.
- **Order:** BIT-45, then BIT-46, then BIT-48, then BIT-49, then BIT-47. (2026-10-01: BIT-50 now follows BIT-47.)
  - BIT-49's migrate reuses BIT-48's wiring step.
  - BIT-47 extends BIT-49's git helper.

## 2026-10-01
- **BIT-47 is split.** BIT-47 keeps the commit side: the commit skill, the reorder in do and bot-dev, the MCP git inputs, and git capture on research, feedback and retro writes. Merge-aware completion moves to the new track BIT-50.
- **Order:** BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50. BIT-50 relies on the bar hashes BIT-47 records.

## 2026-10-01 readiness pass (Claude defaults; the operator can veto these)
- **Track shape and order unchanged.** Each track is a coherent unit that lands on its own, and the dependencies run one way (45 → 46 → 48 → 49 → 47 → 50).
- **Dev runs never use `just install` on `v2`, in every track,** with `XDG_DATA_HOME` and `HOME` sandboxed. `just install` replaces the daily `bp` that every daily session runs as its MCP server. This overrides the memory `bit-pro-just-install-after-changes` for v2 work, so the rule is now stated in every track body, not only BIT-46.
- **Each bar leaves the build and tests green.** Callers move before the packages they import are deleted (`daemon/`, `claude/dispatch.go` here; `bitdir`, `task.Config` in BIT-46).