# BIT-49 research index

Most of this track's evidence lives in BIT-45's and BIT-46's notes, written before the split. Where those notes disagree with the track body, the track body wins.

- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): a claim-by-claim audit (58 confirmed, 6 wrong, 7 stale), plus the build detail for each verse: call sites and tests, the full `.bit/` inventory and v1 layout history, the complete sweep line list, and a bar order. It **corrects the review's flat-archive history**: 95aec1f moved those files to `completed/`, not `archive/tasks/`.
- [review-2026-10-02](review-2026-10-02.md): an independent review against the code and this repo's real `.bit/`. **Read before scoping or planning.**
  - Confirms the track's file:line claims.
  - **Corrects:** Verse 1 must also rewrite the learn body (:8, :16) and retro body (:10, :102, :122), not just the descriptions. There's a third staging line at `do/SKILL.md:75`. The README sweep also covers :82 and :103.
  - Lists tracked docs outside Touches (`hierarchy.md` and others) and the BIT-47/BIT-50 line overlaps.
  - Inventory of bit-pro's `.bit/`: no unknown files, 6 non-`done` records in `completed/`, `order` lists, 52 untracked files.
  - The legacy flat `archive/<ID>.md` layout in older v1 projects would stall migrate.
  - Under-specified points with suggested defaults: the `feedback_read` scope, the `retro_list` scope and replace semantics, retro naming, worktree detection after `bitdir` goes, the subfolder walk-up, and the never-hand-edit rule's wording.
- [decisions](decisions.md): the operator decisions of 2026-09-30 and 2026-10-01, and the readiness-pass Claude defaults. Read first.
- BIT-45 topic `skills` (Verse 1 and Verse 3): which skills and agents do direct file I/O on `.bit/`, the staging lines to delete (it misses do :75, see review), and the MCP tools retro and learn need.
- BIT-45 topic `migrate` (Verse 2): target layout, staged build, verify, commit, idempotency and rollback. It still says `bit.db`, lowercase dirs, git-log timestamps and copying unknown files, all superseded.
- BIT-45 topics `json-schema` and `history-anchors`: the record format and git facts. Read their superseded banners first.
- BIT-46 topic `file-inventory` (Verse 3): every file that names `.bit/`, grouped by owning track and verse. The review adds `README.md:82`, `do/SKILL.md:75` and the tracked design docs.
- BIT-46 topic `plugin-dir-testing`: how Verses 1 and 3 are tested in a dev session.