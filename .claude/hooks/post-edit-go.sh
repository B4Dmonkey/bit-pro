#!/usr/bin/env bash
export PATH=/opt/homebrew/bin:/usr/local/bin:$HOME/go/bin:$PATH

fp=$(jq -r '.tool_input.file_path // empty')
case "$fp" in
  *.go) ;;
  *) exit 0 ;;
esac

fmt_output=$(gofmt -w "$fp" 2>&1)
fmt_status=$?

lint_output=$(golangci-lint run 2>&1)
lint_status=$?

if [ $fmt_status -eq 0 ] && [ $lint_status -eq 0 ]; then
  echo '{"systemMessage":"Go fmt+lint passed"}'
  exit 0
fi

msg="Go checks failed after editing $fp — fix before continuing."
if [ $fmt_status -ne 0 ]; then
  msg="$msg"$'\n\n'"gofmt output:"$'\n'"$fmt_output"
fi
if [ $lint_status -ne 0 ]; then
  msg="$msg"$'\n\n'"golangci-lint output:"$'\n'"$lint_output"
fi

jq -n --arg msg "$msg" '{"systemMessage":"Go fmt+lint FAILED","hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":$msg}}'
