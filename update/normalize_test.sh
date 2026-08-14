#!/usr/bin/env bash
set -uo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
normalize="$script_dir/normalize.sh"
current_test=none

fail() {
    echo "FAIL: $current_test: $1" >&2
    exit 1
}

seed_task() {
    local path="$1" id="$2"
    cat >"$path" <<EOF
---
id: $id
title: seeded task
status: todo
---
body prose mentioning $id
EOF
}

seed_project() {
    local root="$1" p="$2"
    mkdir -p "$root/.bit/tasks" "$root/.bit/completed" "$root/.bit/archive/tasks" "$root/.bit/feedback"
    cat >"$root/.bit/tasks/$p-1.md" <<EOF
---
id: $p-1
title: seeded track
status: doing
order:
    - $p-1.1
---
body prose mentioning $p-1.1
EOF
    seed_task "$root/.bit/tasks/$p-1.1.md" "$p-1.1"
    seed_task "$root/.bit/completed/$p-2.md" "$p-2"
    seed_task "$root/.bit/archive/tasks/$p-3.md" "$p-3"
    printf 'FIRST NOTE\n' >"$root/.bit/feedback/$p-1-001.md"
    cat >"$root/.bit/config.toml" <<EOF
# bit-pro project config
prefix = "$p"
EOF
}

snapshot_tree() {
    local root="$1" path
    find "$root/.bit" -type f | sort | while read -r path; do
        printf '%s  %s\n' "${path#"$root"/}" "$(cksum <"$path")"
    done
}

test_task_filenames_are_uppercased() {
    current_test=test_task_filenames_are_uppercased

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit/tasks"
    seed_task "$root/.bit/tasks/bit-1.md" "bit-1"
    seed_task "$root/.bit/tasks/bit-1.2.md" "bit-1.2"

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local listing
    listing="$(ls -1 "$root/.bit/tasks")"

    grep -qxF 'BIT-1.md' <<<"$listing" || fail "expected entry BIT-1.md, got: $listing"
    grep -qxF 'BIT-1.2.md' <<<"$listing" || fail "expected entry BIT-1.2.md, got: $listing"
    grep -qxF 'bit-1.md' <<<"$listing" && fail "entry bit-1.md still present: $listing"
    grep -qxF 'bit-1.2.md' <<<"$listing" && fail "entry bit-1.2.md still present: $listing"
    grep -q '\.MD$' <<<"$listing" && fail "extension was uppercased: $listing"

    local count
    count="$(grep -c . <<<"$listing")"
    [ "$count" -eq 2 ] || fail "expected 2 files, got $count: $listing"

    rm -rf "$root"
}

test_id_frontmatter_is_uppercased_and_nothing_else() {
    current_test=test_id_frontmatter_is_uppercased_and_nothing_else

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit/tasks"
    cat >"$root/.bit/tasks/bit-1.md" <<'EOF'
---
id: bit-1
title: bit rot in the ingest step
status: todo
---
follow-on work tracked at bit-1.2
EOF

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local path="$root/.bit/tasks/BIT-1.md" contents
    [ -f "$path" ] || fail "expected $path to exist"
    contents="$(cat "$path")"

    grep -qxF 'id: BIT-1' <<<"$contents" || fail "id was not uppercased: $contents"
    grep -qxF 'title: bit rot in the ingest step' <<<"$contents" || fail "title line changed: $contents"
    grep -qxF 'follow-on work tracked at bit-1.2' <<<"$contents" || fail "body prose changed: $contents"

    rm -rf "$root"
}

test_order_entries_are_uppercased_in_frontmatter_only() {
    current_test=test_order_entries_are_uppercased_in_frontmatter_only

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit/tasks"
    cat >"$root/.bit/tasks/bit-1.md" <<'EOF'
---
id: bit-1
title: seeded track
status: todo
order:
    - bit-1.2
    - bit-1.1
---
dropped bars:
    - bit-1.9 was dropped
EOF

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local path="$root/.bit/tasks/BIT-1.md" contents frontmatter entries
    [ -f "$path" ] || fail "expected $path to exist"
    contents="$(cat "$path")"
    frontmatter="$(awk 'NR > 1 && /^---$/ { exit } NR > 1 { print }' "$path")"
    entries="$(grep -E '^    - ' <<<"$frontmatter")"

    [ "$entries" = "    - BIT-1.2
    - BIT-1.1" ] || fail "order entries wrong or reordered: $entries"
    grep -qxF '    - bit-1.9 was dropped' <<<"$contents" || fail "body list entry changed: $contents"

    rm -rf "$root"
}

test_completed_and_archived_tasks_are_normalized() {
    current_test=test_completed_and_archived_tasks_are_normalized

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit/tasks" "$root/.bit/completed" "$root/.bit/archive/tasks"
    seed_task "$root/.bit/tasks/bit-1.md" "bit-1"
    seed_task "$root/.bit/completed/bit-2.md" "bit-2"
    seed_task "$root/.bit/archive/tasks/bit-3.md" "bit-3"

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local path
    for path in "$root/.bit/tasks/BIT-1.md" "$root/.bit/completed/BIT-2.md" "$root/.bit/archive/tasks/BIT-3.md"; do
        [ -f "$path" ] || fail "expected $path to exist"
    done

    grep -qxF 'id: BIT-1' "$root/.bit/tasks/BIT-1.md" || fail "tasks/ id not uppercased"
    grep -qxF 'id: BIT-2' "$root/.bit/completed/BIT-2.md" || fail "completed/ id not uppercased"
    grep -qxF 'id: BIT-3' "$root/.bit/archive/tasks/BIT-3.md" || fail "archive/tasks/ id not uppercased"

    local remaining
    remaining="$(find "$root/.bit" -name '*.md' -print | while read -r path; do
        stem="$(basename "$path" .md)"
        case "$stem" in *[a-z]*) echo "$stem" ;; esac
    done)"
    [ -z "$remaining" ] || fail "lowercase task filenames remain: $remaining"

    rm -rf "$root"
}

test_feedback_note_filenames_are_uppercased_contents_intact() {
    current_test=test_feedback_note_filenames_are_uppercased_contents_intact

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit/feedback"
    printf 'FIRST NOTE\n' >"$root/.bit/feedback/bit-1-001.md"
    printf 'SECOND NOTE\n' >"$root/.bit/feedback/bit-1-002.md"

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local listing
    listing="$(ls -1 "$root/.bit/feedback")"

    grep -qxF 'BIT-1-001.md' <<<"$listing" || fail "expected entry BIT-1-001.md, got: $listing"
    grep -qxF 'BIT-1-002.md' <<<"$listing" || fail "expected entry BIT-1-002.md, got: $listing"

    local count
    count="$(grep -c . <<<"$listing")"
    [ "$count" -eq 2 ] || fail "expected 2 files, got $count: $listing"

    local first second
    first="$(cat "$root/.bit/feedback/BIT-1-001.md")"
    second="$(cat "$root/.bit/feedback/BIT-1-002.md")"
    [ "$first" = "FIRST NOTE" ] || fail "first note contents changed: $first"
    [ "$second" = "SECOND NOTE" ] || fail "second note contents changed: $second"

    rm -rf "$root"
}

test_config_prefix_is_uppercased() {
    current_test=test_config_prefix_is_uppercased

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    mkdir -p "$root/.bit"
    cat >"$root/.bit/config.toml" <<'EOF'
# bit-pro project config
prefix = "bit"
EOF

    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"

    local contents
    contents="$(cat "$root/.bit/config.toml")"

    grep -qxF 'prefix = "BIT"' <<<"$contents" || fail "prefix not uppercased: $contents"
    grep -qxF '# bit-pro project config' <<<"$contents" || fail "comment line changed: $contents"

    rm -rf "$root"
}

test_second_run_changes_nothing() {
    current_test=test_second_run_changes_nothing

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    seed_project "$root" bit

    bash "$normalize" "$root" || fail "first normalize.sh run exited non-zero"
    local first second
    first="$(snapshot_tree "$root")"

    bash "$normalize" "$root" || fail "second normalize.sh run exited non-zero"
    second="$(snapshot_tree "$root")"

    [ "$first" = "$second" ] || fail "second run changed the tree:
after one run:
$first
after two runs:
$second"

    rm -rf "$root"
}

test_already_uppercase_project_is_untouched() {
    current_test=test_already_uppercase_project_is_untouched

    local root
    root="$(mktemp -d)" || fail "could not create temp root"
    seed_project "$root" BIT

    local before after
    before="$(snapshot_tree "$root")"
    bash "$normalize" "$root" || fail "normalize.sh exited non-zero"
    after="$(snapshot_tree "$root")"

    [ "$before" = "$after" ] || fail "uppercase project was changed:
before:
$before
after:
$after"

    rm -rf "$root"
}

test_directory_without_bit_is_an_error() {
    current_test=test_directory_without_bit_is_an_error

    local root empty
    root="$(mktemp -d)" || fail "could not create temp root"
    empty="$(mktemp -d)" || fail "could not create temp root"
    seed_project "$root" bit

    local before after status stderr
    before="$(snapshot_tree "$root")"

    stderr="$(bash "$normalize" "$empty" 2>&1 >/dev/null)"
    status=$?
    [ "$status" -ne 0 ] || fail "expected non-zero exit for a root without .bit/"
    grep -qF "$empty" <<<"$stderr" || fail "stderr does not name the offending directory: $stderr"

    stderr="$(bash "$normalize" "$root" "$empty" 2>&1 >/dev/null)"
    status=$?
    [ "$status" -ne 0 ] || fail "expected non-zero exit when one root of several lacks .bit/"
    grep -qF "$empty" <<<"$stderr" || fail "stderr does not name the offending directory: $stderr"

    after="$(snapshot_tree "$root")"
    [ "$before" = "$after" ] || fail "valid project was modified before validation failed:
before:
$before
after:
$after"

    rm -rf "$root" "$empty"
}

test_no_arguments_is_an_error() {
    current_test=test_no_arguments_is_an_error

    local status stderr
    stderr="$(bash "$normalize" 2>&1 >/dev/null)"
    status=$?
    [ "$status" -ne 0 ] || fail "expected non-zero exit with no arguments"
    grep -qF '<project-dir>...' <<<"$stderr" || fail "stderr does not show a usage line: $stderr"
}

test_task_filenames_are_uppercased
test_id_frontmatter_is_uppercased_and_nothing_else
test_order_entries_are_uppercased_in_frontmatter_only
test_completed_and_archived_tasks_are_normalized
test_feedback_note_filenames_are_uppercased_contents_intact
test_config_prefix_is_uppercased
test_second_run_changes_nothing
test_already_uppercase_project_is_untouched
test_directory_without_bit_is_an_error
test_no_arguments_is_an_error
echo "ok"
