#!/usr/bin/env bash
set -euo pipefail

level="${1-}"
case "$level" in
    major|minor|patch) ;;
    *)
        echo "release: unknown level '$level' — expected one of: major, minor, patch" >&2
        exit 1
        ;;
esac

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

branch="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch" != "main" ]; then
    echo "release: on branch '$branch' — releases are cut from main" >&2
    exit 1
fi

if ! git diff-index --quiet HEAD --; then
    echo "release: working tree is dirty — commit or discard tracked changes first" >&2
    git status --porcelain -uno >&2
    exit 1
fi

current="$(jq -r '.version // empty' bit/.claude-plugin/plugin.json)"

if [ -z "$current" ]; then
    next="0.1.0"
else
    IFS=. read -r major minor patch <<<"$current"
    case "$level" in
        major) next="$((major + 1)).0.0" ;;
        minor) next="$major.$((minor + 1)).0" ;;
        patch) next="$major.$minor.$((patch + 1))" ;;
    esac
fi

tags="$(git tag --list 'bit--v*' --sort=-v:refname)"
highest="${tags%%$'\n'*}"
highest="${highest#bit--v}"
if [ -n "$highest" ]; then
    top="$(printf '%s\n%s\n' "$highest" "$next" | sort -V | tail -1)"
    if [ "$top" != "$next" ] || [ "$next" = "$highest" ]; then
        echo "release: computed $next is not above the highest existing tag $highest" >&2
        exit 1
    fi
fi

manifest="bit/.claude-plugin/plugin.json"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
jq --arg v "$next" '.version = $v' "$manifest" >"$tmp"
mv "$tmp" "$manifest"

claude plugin validate bit --strict

git commit -q -m "release: v$next" "$manifest"

claude plugin tag bit

echo "bit--v$next"
