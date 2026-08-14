#!/usr/bin/env bash
set -euo pipefail

uppercase() {
    printf '%s' "$1" | tr '[:lower:]' '[:upper:]'
}

rename_case_only() {
    local src="$1" dst="$2"
    if [ "$src" = "$dst" ]; then
        return 0
    fi
    local tmp="$src.normalize.$$"
    mv "$src" "$tmp"
    mv "$tmp" "$dst"
}

normalize_task_filenames() {
    local dir="$1" path base stem
    for path in "$dir"/*.md; do
        [ -e "$path" ] || continue
        base="$(basename "$path")"
        stem="${base%.md}"
        rename_case_only "$path" "$dir/$(uppercase "$stem").md"
    done
}

normalize_id_fields() {
    local dir="$1" path id
    for path in "$dir"/*.md; do
        [ -e "$path" ] || continue
        id="$(sed -n '/^id: /{s/^id: //p;q;}' "$path")"
        [ -n "$id" ] || continue
        sed -i '' "s/^id: .*/id: $(uppercase "$id")/" "$path"
    done
}

normalize_order_entries() {
    local dir="$1" path tmp
    for path in "$dir"/*.md; do
        [ -e "$path" ] || continue
        tmp="$path.normalize.$$"
        awk '
            NR == 1 && $0 == "---" { in_frontmatter = 1; print; next }
            in_frontmatter && $0 == "---" { in_frontmatter = 0; print; next }
            in_frontmatter && /^ *- / { print toupper($0); next }
            { print }
        ' "$path" >"$tmp"
        mv "$tmp" "$path"
    done
}

normalize_task_dir() {
    local dir="$1"
    [ -d "$dir" ] || return 0
    normalize_task_filenames "$dir"
    normalize_id_fields "$dir"
    normalize_order_entries "$dir"
}

main() {
    local root
    for root in "$@"; do
        normalize_task_dir "$root/.bit/tasks"
        normalize_task_dir "$root/.bit/completed"
        normalize_task_dir "$root/.bit/archive/tasks"
    done
}

main "$@"
