#!/usr/bin/env bash
set -euo pipefail

echo "Scanning tracked repository files for likely OpenAI API keys (pattern: sk-)"

# Look only in files tracked by git (avoids scanning binary deps)
# Regex: sk- followed by 20+ allowed characters
pattern='sk-[A-Za-z0-9_-]{20,}'

matches=$(git ls-files -z | xargs -0 grep -nE --color=never "$pattern" || true)

if [ -z "$matches" ]; then
  echo "No likely OpenAI keys found in tracked files."
  exit 0
fi

echo "Potential matches found:" 
echo
printf "%s\n" "$matches"

echo
echo "If you see an exposed key, immediately revoke it at https://platform.openai.com/account/api-keys and rotate to a new key."

exit 0
