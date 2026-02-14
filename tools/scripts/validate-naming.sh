#!/usr/bin/env bash
set -euo pipefail

# Validates file naming conventions across the repository.
# - Go source files must be snake_case.go
# - Test files must end with _test.go
# - Migration files must match NNNNN_description.sql or NNNNN_description.go
# - No uppercase letters in Go filenames

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0

error() {
  echo "ERROR: $1" >&2
  ERRORS=$((ERRORS + 1))
}

# Find all Go files, excluding vendor and generated code
GO_FILES=$(find "$ROOT" -name '*.go' \
  -not -path '*/vendor/*' \
  -not -path '*/tests/mocks/*' \
  -not -path '*/specs/*' \
  -not -path '*/.git/*')

# Check 1: No uppercase letters in Go filenames
echo "==> Check 1: No uppercase letters in Go filenames"
while IFS= read -r file; do
  basename=$(basename "$file")
  if [[ "$basename" =~ [A-Z] ]]; then
    error "$file — filename contains uppercase letters"
  fi
done <<< "$GO_FILES"

# Check 2: Go filenames are snake_case (letters, digits, underscores only before .go / _test.go)
# Migration Go files (NNNNN_*.go) are excluded — they're validated by check 3.
echo "==> Check 2: Go filenames are snake_case"
while IFS= read -r file; do
  basename=$(basename "$file")
  # Skip migration Go files (start with digits)
  if [[ "$basename" =~ ^[0-9] ]]; then
    continue
  fi
  # Strip _test.go or .go suffix
  name="${basename%_test.go}"
  if [[ "$name" == "$basename" ]]; then
    name="${basename%.go}"
  fi
  # Allow: lowercase letters, digits, underscores, dots (for foo.sql.go generated files)
  if [[ ! "$name" =~ ^[a-z][a-z0-9_.]*$ ]]; then
    error "$file — filename is not snake_case: $basename"
  fi
done <<< "$GO_FILES"

# Check 3: Migration files match NNNNN_description pattern
echo "==> Check 3: Migration file naming"
MIGRATION_DIRS=$(find "$ROOT" -type d -name 'versions' -path '*/migrations/*')
for dir in $MIGRATION_DIRS; do
  for file in "$dir"/*; do
    [[ -f "$file" ]] || continue
    basename=$(basename "$file")
    # Skip non-migration helper files (doc.go, helpers.go, etc.)
    # Migration files start with digits
    if [[ "$basename" =~ ^[0-9] ]]; then
      if [[ ! "$basename" =~ ^[0-9]{5}_[a-z][a-z0-9_]*\.(sql|go)$ ]]; then
        error "$file — migration filename doesn't match NNNNN_description.{sql,go}: $basename"
      fi
    fi
  done
done

# Summary
echo ""
if [[ $ERRORS -gt 0 ]]; then
  echo "FAIL: $ERRORS naming violation(s) found." >&2
  exit 1
else
  echo "OK: All naming conventions pass."
fi
