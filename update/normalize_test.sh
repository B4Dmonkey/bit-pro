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

test_task_filenames_are_uppercased
test_id_frontmatter_is_uppercased_and_nothing_else
echo "ok"
