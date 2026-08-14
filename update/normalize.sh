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

main() {
    local root
    for root in "$@"; do
        normalize_task_filenames "$root/.bit/tasks"
    done
}

main "$@"
