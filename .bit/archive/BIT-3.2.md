---
id: BIT-3.2
title: Two-digit IDs contradict the lexical sort
status: done
phase: 1
phase_label: Newest first
---
## Step 2 (Phase 1 — Newest first) — Two-digit IDs contradict the lexical sort

**Status:** ✅ Done — verified 2026-07-17

`BIT-10` sorts lexically between `BIT-1` and `BIT-2`, so reversing lexical order gets
`BIT-9` before `BIT-10`. Only parsing the number satisfies both this test and Step 1's.
This closes the README's "Known limitations" entry.

**Scope:**
- `task/store.go` — `List`: sort parsed tasks, not filename strings; new unexported
  `compareIDs(a, b string) int`
- `task/store_test.go` — new table test for `compareIDs`
- `README.md` — remove the lexical-sort entry from **Known limitations**

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskListCmd_OrdersNumericallyNotLexically`
     - **Behavior:** ordering follows the ID's number, so a two-digit task doesn't wedge
       itself among the single-digit ones.
     - **Setup:** `initProject(t, "BIT")`, then create 10 tasks named `"T1"…"T10"`,
       yielding BIT-1 … BIT-10.
     - **Assertions:** the ID column, in order, is exactly `BIT-10, BIT-9, BIT-8, BIT-7,
       BIT-6, BIT-5, BIT-4, BIT-3, BIT-2, BIT-1`.
     - **Boundary:** ID number crosses 9→10, the digit-count boundary where lexical and
       numeric ordering first disagree.
   - [x] Confirm fails: reverse-lexical yields `BIT-9, BIT-8 … BIT-2, BIT-10, BIT-1`

2. **Implement (GREEN):**
   - [x] Add `compareIDs(a, b string) int`: take the substring after the last `-`,
         `strconv.Atoi` it, return the comparison **descending**. Unparseable suffix
         sorts last — it can only come from a hand-edited file.
   - [x] In `List`, drop `slices.Sort(matches)` / `slices.Reverse(matches)`; after the
         parse loop, `slices.SortFunc(tasks, func(a, b *Task) int { return compareIDs(a.ID, b.ID) })`

3. **More tests (RED → GREEN):**
   - [x] `TestCompareIDs` (table-driven, `t.Parallel()`, matching `store_test.go`'s style)
     - **Behavior:** the comparison is a correct total ordering in isolation, so the
       hierarchy work in Step 5 has a pinned starting point to change.
     - **Setup:** pairs — `("BIT-2","BIT-1")`, `("BIT-1","BIT-2")`, `("BIT-1","BIT-1")`,
       `("BIT-10","BIT-9")`, `("BIT-abc","BIT-1")`.
     - **Assertions:** `<0`, `>0`, `==0`, `<0`, `>0` respectively.
     - **Boundary:** equal IDs — the reflexive case a comparator must return 0 for, and
       the one a naive `a > b` implementation gets wrong.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] `just build && ./bin/bit task list` in this repo reads right against real records

**Commit (user):** `fix(list): order tasks numerically, not lexically`