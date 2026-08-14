#!/usr/bin/env bash
set -euo pipefail

git_root=

uppercase() {
    printf '%s' "$1" | tr '[:lower:]' '[:upper:]'
}

rename_case_only() {
    local src="$1" dst="$2"
    if [ "$src" = "$dst" ]; then
        return 0
    fi
    if [ -n "$git_root" ]; then
        git -C "$git_root" mv --force "${src#"$git_root"/}" "${dst#"$git_root"/}"
        return 0
    fi
    local tmp="$src.normalize.$$"
    mv "$src" "$tmp"
    mv "$tmp" "$dst"
}

normalize_md_filenames() {
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
    normalize_md_filenames "$dir"
    normalize_id_fields "$dir"
    normalize_order_entries "$dir"
}

normalize_feedback_dir() {
    local dir="$1"
    [ -d "$dir" ] || return 0
    normalize_md_filenames "$dir"
}

normalize_config_prefix() {
    local file="$1" prefix
    [ -f "$file" ] || return 0
    prefix="$(sed -n '/^prefix = "/{s/^prefix = "\(.*\)"$/\1/p;q;}' "$file")"
    [ -n "$prefix" ] || return 0
    sed -i '' "s/^prefix = \".*\"\$/prefix = \"$(uppercase "$prefix")\"/" "$file"
}

validate_roots() {
    local root invalid=0
    if [ "$#" -eq 0 ]; then
        echo "usage: normalize.sh <project-dir>..." >&2
        return 1
    fi
    for root in "$@"; do
        if [ ! -d "$root/.bit" ]; then
            echo "normalize.sh: $root: not a bit project (no .bit directory)" >&2
            invalid=1
        fi
    done
    return "$invalid"
}

main() {
    validate_roots "$@" || exit 1

    local root
    for root in "$@"; do
        if git -C "$root" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
            git_root="$root"
        else
            git_root=
        fi
        normalize_task_dir "$root/.bit/tasks"
        normalize_task_dir "$root/.bit/completed"
        normalize_task_dir "$root/.bit/archive/tasks"
        normalize_feedback_dir "$root/.bit/feedback"
        normalize_config_prefix "$root/.bit/config.toml"
    done
}

main "$@"
