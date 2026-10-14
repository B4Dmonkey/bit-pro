#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

git update-index -q --refresh
if ! git diff-index --quiet HEAD --; then
    echo "release-push: working tree is dirty — commit or discard tracked changes first" >&2
    git status --porcelain -uno >&2
    exit 1
fi

tags="$(git tag --list 'v*' --sort=-v:refname)"
tag="${tags%%$'\n'*}"

if [ -z "$tag" ]; then
    echo "release-push: no v* tag — run 'just release <level>' first" >&2
    exit 1
fi

git push origin "$tag"

echo "$tag"
