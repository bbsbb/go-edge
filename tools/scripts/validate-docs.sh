#!/usr/bin/env bash
set -euo pipefail

# Validates the docs/ knowledge base structure:
# 1. Every .md file in docs/ must be reachable from CLAUDE.md
#    (directly linked, or inside a linked directory, or transitively via a linked file)
# 2. docs/design-docs/index.md must reference every .md in docs/design-docs/
# 3. Internal markdown links must resolve to existing files
#
# Exit code 0 = all checks pass
# Exit code 1 = validation errors found

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CLAUDE_MD="$REPO_ROOT/CLAUDE.md"
ERRORS=0

error() {
    echo "ERROR: $1"
    ERRORS=$((ERRORS + 1))
}

warn() {
    echo "WARN: $1"
}

# --- Check 1: Every .md in docs/ is reachable from CLAUDE.md ---

echo "==> Checking docs/ reachability from CLAUDE.md..."

# Extract all docs/ paths referenced in CLAUDE.md (from markdown links)
CLAUDE_REFS=$(grep -oE '\./docs/[^)]*' "$CLAUDE_MD" | sed 's|^\./||' || true)

# For each .md file in docs/, check if it's reachable
while IFS= read -r doc; do
    [ -z "$doc" ] && continue
    rel_path="${doc#"$REPO_ROOT/"}"

    reachable=false

    # Check direct link
    if echo "$CLAUDE_REFS" | grep -qF "$rel_path"; then
        reachable=true
    fi

    # Check if file is inside a linked directory
    if [ "$reachable" = false ]; then
        while IFS= read -r ref; do
            [ -z "$ref" ] && continue
            # If ref ends with /, it's a directory reference
            if [[ "$ref" == */ ]] && [[ "$rel_path" == "$ref"* ]]; then
                reachable=true
                break
            fi
        done <<< "$CLAUDE_REFS"
    fi

    # Check if file is referenced transitively from a linked .md file
    if [ "$reachable" = false ]; then
        while IFS= read -r ref; do
            [ -z "$ref" ] && continue
            ref_file="$REPO_ROOT/$ref"
            if [ -f "$ref_file" ] && grep -qF "$(basename "$rel_path")" "$ref_file" 2>/dev/null; then
                reachable=true
                break
            fi
        done <<< "$CLAUDE_REFS"
    fi

    if [ "$reachable" = false ]; then
        error "$rel_path is not reachable from CLAUDE.md. Add a direct link or ensure it's inside a linked directory."
    fi
done < <(find "$REPO_ROOT/docs" -name '*.md' -type f)

# --- Check 2: design-docs/index.md references all .md files in design-docs/ ---

echo "==> Checking design-docs/index.md completeness..."

DESIGN_INDEX="$REPO_ROOT/docs/design-docs/index.md"
if [ -f "$DESIGN_INDEX" ]; then
    while IFS= read -r doc; do
        [ -z "$doc" ] && continue
        filename="$(basename "$doc")"
        # Skip index.md itself
        [ "$filename" = "index.md" ] && continue

        if ! grep -qF "$filename" "$DESIGN_INDEX"; then
            error "docs/design-docs/$filename exists but is not listed in docs/design-docs/index.md. Add it to the catalogue."
        fi
    done < <(find "$REPO_ROOT/docs/design-docs" -name '*.md' -type f)
else
    error "docs/design-docs/index.md does not exist."
fi

# --- Check 2b: product-specs/index.md references all .md files in product-specs/ ---

echo "==> Checking product-specs/index.md completeness..."

PRODUCT_INDEX="$REPO_ROOT/docs/product-specs/index.md"
if [ -f "$PRODUCT_INDEX" ]; then
    while IFS= read -r doc; do
        [ -z "$doc" ] && continue
        filename="$(basename "$doc")"
        [ "$filename" = "index.md" ] && continue

        if ! grep -qF "$filename" "$PRODUCT_INDEX"; then
            error "docs/product-specs/$filename exists but is not listed in docs/product-specs/index.md. Add it to the catalogue."
        fi
    done < <(find "$REPO_ROOT/docs/product-specs" -name '*.md' -type f)
else
    error "docs/product-specs/index.md does not exist."
fi

# --- Check 3: Internal markdown links resolve to existing files ---

echo "==> Checking internal markdown links..."

while IFS= read -r doc; do
    [ -z "$doc" ] && continue
    doc_dir="$(dirname "$doc")"

    # Extract relative markdown links (not http/https)
    while IFS= read -r link; do
        [ -z "$link" ] && continue

        # Skip external links
        [[ "$link" == http* ]] && continue
        [[ "$link" == "#"* ]] && continue

        # Strip anchor
        link="${link%%#*}"
        [ -z "$link" ] && continue

        # Resolve relative to the document's directory
        target="$doc_dir/$link"
        if [ ! -e "$target" ]; then
            rel_doc="${doc#"$REPO_ROOT/"}"
            error "$rel_doc contains broken link to '$link' (resolved: ${target#"$REPO_ROOT/"})"
        fi
    done < <(grep -oE '\]\([^)]+' "$doc" | sed 's/^](//' || true)
done < <(find "$REPO_ROOT/docs" "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/ARCHITECTURE.md" -name '*.md' -type f 2>/dev/null)

# --- Check 4: Doc freshness (reviewed date must be >= last commit date) ---

echo "==> Checking doc freshness..."

while IFS= read -r doc; do
    [ -z "$doc" ] && continue
    rel_path="${doc#"$REPO_ROOT/"}"

    # Skip generated docs
    [[ "$rel_path" == docs/generated/* ]] && continue

    reviewed_date=$(grep -oE 'last-reviewed: [0-9]{4}-[0-9]{2}-[0-9]{2}' "$doc" | head -1 | sed 's/last-reviewed: //' || true)

    if [ -z "$reviewed_date" ]; then
        error "$rel_path is missing a <!-- last-reviewed: YYYY-MM-DD --> tag."
        continue
    fi

    last_commit_date=$(git -C "$REPO_ROOT" log -1 --format='%as' -- "$rel_path" 2>/dev/null || true)
    if [ -n "$last_commit_date" ] && [ "$reviewed_date" \< "$last_commit_date" ]; then
        error "$rel_path was modified ($last_commit_date) after its last-reviewed date ($reviewed_date). Update the last-reviewed tag."
    fi
done < <(find "$REPO_ROOT/docs" "$REPO_ROOT/ARCHITECTURE.md" "$REPO_ROOT/CLAUDE.md" -name '*.md' -type f 2>/dev/null)

# --- Summary ---

if [ "$ERRORS" -gt 0 ]; then
    echo ""
    echo "Validation failed with $ERRORS error(s)."
    exit 1
else
    echo ""
    echo "All docs validation checks passed."
    exit 0
fi
