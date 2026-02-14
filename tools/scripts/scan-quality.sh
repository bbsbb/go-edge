#!/usr/bin/env bash
set -euo pipefail

# Scans for quality deviations across the codebase.
# Run on-demand to identify areas needing attention.
#
# Checks:
# 1. Go packages without test files
# 2. Go files exceeding 500 lines (revive backup)
# 3. Public packages missing package-level doc comments
# 4. FX modules missing quality grades in QUALITY.md
#
# Exit code 0 = no issues found
# Exit code 1 = issues detected (informational, not blocking)

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
QUALITY_MD="$ROOT/docs/QUALITY.md"
ISSUES=0

issue() {
    echo "ISSUE: $1"
    ISSUES=$((ISSUES + 1))
}

# --- Check 1: Go packages without test files ---

echo "==> Checking for packages without tests..."

while IFS= read -r dir; do
    [ -z "$dir" ] && continue

    # Skip directories that shouldn't have tests
    basename_dir="$(basename "$dir")"
    case "$basename_dir" in
        mocks|sqlcgen|migrations|versions|cmd|testdata|testing) continue ;;
    esac

    # Skip packages that are test-exempt by design
    rel_check="${dir#"$ROOT/"}"
    case "$rel_check" in
        core/fx/bootfx) continue ;;    # composition root, no testable logic
        core/secretstore) continue ;;   # interface-only package
    esac

    # Skip app packages covered by handler integration tests
    case "$rel_check" in
        apps/*/internal/domain) continue ;;
        apps/*/internal/service) continue ;;
        apps/*/internal/config) continue ;;
        apps/*/internal/infrastructure/persistence) continue ;;
        apps/*/internal/migrations) continue ;;
        apps/*/internal/transport/http) continue ;;
        apps/*/internal/transport/http/dto) continue ;;
    esac

    # Skip if directory has no .go files
    go_files=$(find "$dir" -maxdepth 1 -name '*.go' ! -name '*_test.go' ! -name 'doc.go' ! -name 'tools.go' 2>/dev/null | head -1)
    [ -z "$go_files" ] && continue

    # Check for test files
    test_files=$(find "$dir" -maxdepth 1 -name '*_test.go' 2>/dev/null | head -1)
    if [ -z "$test_files" ]; then
        rel="${dir#"$ROOT/"}"
        issue "$rel/ has Go source files but no tests"
    fi
done < <(find "$ROOT/core" "$ROOT/apps" -type d 2>/dev/null)

# --- Check 2: Go files exceeding 500 lines ---

echo "==> Checking for oversized Go files (>500 lines)..."

while IFS= read -r file; do
    [ -z "$file" ] && continue
    lines=$(wc -l < "$file")
    if [ "$lines" -gt 500 ]; then
        rel="${file#"$ROOT/"}"
        issue "$rel is $lines lines (limit: 500)"
    fi
done < <(find "$ROOT/core" "$ROOT/apps" -name '*.go' \
    -not -path '*/mocks/*' \
    -not -path '*/sqlcgen/*' \
    -not -path '*/vendor/*' \
    -not -path '*/.git/*' \
    2>/dev/null)

# --- Check 3: Public packages missing doc comments ---

echo "==> Checking for public packages missing doc comments..."

while IFS= read -r dir; do
    [ -z "$dir" ] && continue

    # Skip internal, test, generated, and mock packages
    case "$dir" in
        */internal/*|*/mocks/*|*/sqlcgen/*|*/testdata/*|*/tests/*) continue ;;
    esac

    # Find any non-test Go file
    first_go=$(find "$dir" -maxdepth 1 -name '*.go' ! -name '*_test.go' ! -name 'doc.go' ! -name 'tools.go' 2>/dev/null | head -1)
    [ -z "$first_go" ] && continue

    # Check for doc.go or package comment in any file
    has_doc=false
    if [ -f "$dir/doc.go" ]; then
        has_doc=true
    else
        for f in "$dir"/*.go; do
            [[ "$f" == *_test.go ]] && continue
            if head -20 "$f" | grep -q '^// Package '; then
                has_doc=true
                break
            fi
        done
    fi

    if [ "$has_doc" = false ]; then
        rel="${dir#"$ROOT/"}"
        issue "$rel/ is a public package without a package doc comment"
    fi
done < <(find "$ROOT/core" -type d -not -path '*/internal/*' 2>/dev/null)

# --- Check 4: FX modules missing quality grades ---

echo "==> Checking FX modules have quality grades..."

if [ -f "$QUALITY_MD" ]; then
    while IFS= read -r dir; do
        [ -z "$dir" ] && continue
        module="$(basename "$dir")"
        if ! grep -qF "$module" "$QUALITY_MD"; then
            issue "core/fx/$module/ has no quality grade in QUALITY.md"
        fi
    done < <(find "$ROOT/core/fx" -type d -mindepth 1 -maxdepth 1 2>/dev/null)
fi

# --- Summary ---

echo ""
if [ "$ISSUES" -gt 0 ]; then
    echo "Quality scan found $ISSUES issue(s)."
    exit 1
else
    echo "No quality issues found."
    exit 0
fi
