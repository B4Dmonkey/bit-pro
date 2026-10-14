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

git update-index -q --refresh
if ! git diff-index --quiet HEAD --; then
    echo "release: working tree is dirty — commit or discard tracked changes first" >&2
    git status --porcelain -uno >&2
    exit 1
fi

tags="$(git tag --list 'v*' --sort=-v:refname)"
current="${tags%%$'\n'*}"
current="${current#v}"

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

git tag -a "v$next" -m "bit $next"

echo "v$next"
