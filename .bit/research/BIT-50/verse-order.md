# Suggested verse order (walking skeleton first)

Proposal for bit_scope, not a decision:

1. **Skeleton: direct commits on trunk.** Completion resolves trunk (`origin/main`, else `main`), checks every bar hash with is-ancestor; all landed → records the landing commit (newest bar, first-parent rule so merges also work) on the track and files it. Anything else → "not landed yet: fetch/pull/push and retry", no filing. Move the complete skill's trigger from sign-off to after landing (do's sign-off text). Covers 304/320 real commits.
2. **Classification + operator check-ins:** partly done / not done / local-only / pushed-not-merged / can't-tell, with archive-or-keep prompts (`done-classification`).
3. **Squash rung + bar repointing:** subject matching incl. single-commit squash shape and duplicate squashes; repoint bars before filing.
4. **Ask-for-PR rung:** map `(#N)` to the trunk squash locally; fall back to operator-supplied hash.

Merge-commit landing fits inside verse 1 (one extra rev-list) rather than its own verse. Rebase-merge detection (patch-id/subject lookup) is out of the decided ladder; leave as a risk.
