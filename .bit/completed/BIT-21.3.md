---
id: BIT-21.3
title: Order entries flip inside the frontmatter only
status: done
approved: true
phase: 1
phase_label: Migration
---
## **Verse 1**

The third carrier: a track's `order:` list. Measured format is a block list indented four
spaces (`    - BIT-10.1`), and that shape can legitimately appear in body prose too — so the
rewrite has to stop at the closing frontmatter delimiter. That constraint is what forces this
to be its own step rather than an extension of the `id:` rewrite.

## Scope
- `update/normalize.sh` — uppercase `order:` list entries, frontmatter only.
- `update/normalize_test.sh` — fixture track carrying an order list plus a body checklist that
  mimics its shape.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_order_entries_are_uppercased_in_frontmatter_only`
     - **Behavior:** every entry in a track's `order:` list is uppercased, and an identically
       shaped list in the body is not.
     - **Setup:** `.bit/tasks/bit-1.md` whose frontmatter carries

       ```
       order:
           - bit-1.2
           - bit-1.1
       ```

       and whose body contains a bullet reading `    - bit-1.9 was dropped`. Run the script.
     - **Assertions:** frontmatter reads `- BIT-1.2` then `- BIT-1.1`, in that order — the
       sequence is preserved, only the case changes; the body bullet still reads
       `    - bit-1.9 was dropped`.
     - **Boundary:** the frontmatter/body delimiter itself. A body bullet one line past the
       closing `---` is the nearest possible false positive, so this is the tightest case that
       distinguishes a delimiter-aware rewrite from a whole-file one.
   - [ ] Confirm fails: the frontmatter entries are still lowercase

2. **Implement (GREEN):**
   - [ ] Confine the rewrite to lines between the opening `---` and the second `---`, and within
     that range uppercase list entries under `order:`. `sed` range addressing (`/^---$/,/^---$/`)
     is enough; no YAML parsing.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): uppercase order list entries`