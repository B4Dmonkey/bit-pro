---
id: BIT-21.6
title: Config prefix flips
status: done
approved: true
phase: 1
phase_label: Migration
---
## **Verse 1**

The last carrier: the `prefix` in `.bit/config.toml`. It is what every newly minted ID is built
from, so a project whose files are all uppercase but whose prefix still reads `bit` mints a
lowercase ID on the very next `bp task create` — undoing the migration one task at a time.

## Scope
- `update/normalize.sh` — uppercase the `prefix` value in `.bit/config.toml`.
- `update/normalize_test.sh` — fixture config alongside a same-value string that must not move.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_config_prefix_is_uppercased`
     - **Behavior:** the stored prefix becomes uppercase without disturbing the rest of the file.
     - **Setup:** `.bit/config.toml` containing `prefix = "bit"` preceded by a comment line
       `# bit-pro project config`. Run the script.
     - **Assertions:** the file contains exactly `prefix = "BIT"`; the comment line is unchanged;
       the TOML quoting style is preserved (double quotes, spaces around `=`), since the Go
       side rewrites this file with `toml.Marshal` and a mangled file would fail to decode.
     - **Boundary:** a lowercase `bit` occurring twice, once as the carrier and once as prose —
       the same anchoring requirement as the `id:` field, applied to a different file format.
   - [ ] Confirm fails: `config.toml` still reads `prefix = "bit"`

2. **Implement (GREEN):**
   - [ ] Rewrite only the `prefix = "…"` value in `.bit/config.toml`, uppercasing what is inside
     the quotes. A missing `config.toml` is skipped, not an error.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): uppercase the config prefix`