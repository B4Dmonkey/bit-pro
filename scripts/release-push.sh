#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

if ! git diff-index --quiet HEAD --; then
    echo "release-push: working tree is dirty — commit or discard tracked changes first" >&2
    git status --porcelain -uno >&2
    exit 1
fi

version="$(jq -r '.version // empty' bit/.claude-plugin/plugin.json)"
tag="bit--v$version"

if ! git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
    echo "release-push: no local tag $tag — run 'just release <level>' first" >&2
    exit 1
fi

git push origin "$tag"

echo "$tag"
