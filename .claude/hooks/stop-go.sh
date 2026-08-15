#!/usr/bin/env bash
export PATH=/opt/homebrew/bin:/usr/local/bin:$HOME/go/bin:$PATH

git status --porcelain -- '*.go' 2>/dev/null | grep -q . || exit 0

lint_output=$(just lint 2>&1)
lint_status=$?

test_output=$(just test 2>&1)
test_status=$?

if [ $lint_status -eq 0 ] && [ $test_status -eq 0 ]; then
  echo '{"systemMessage":"Go: lint clean, tests passed"}'
  exit 0
fi

reason="Go checks failed — do not stop until these are fixed."
if [ $lint_status -ne 0 ]; then
  reason="$reason"$'\n\n'"lint output:"$'\n'"$lint_output"
fi
if [ $test_status -ne 0 ]; then
  reason="$reason"$'\n\n'"go test output:"$'\n'"$test_output"
fi

jq -n --arg reason "$reason" '{"decision":"block","reason":$reason}'
