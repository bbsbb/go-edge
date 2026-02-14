#!/usr/bin/env bash
set -euo pipefail

# Validates that ARCHITECTURE.md stays in sync with the actual codebase:
# 1. Every package under apps/<app>/internal/ is listed in ARCHITECTURE.md
# 2. Every FX module in core/fx/ is referenced in ARCHITECTURE.md
# 3. Forbidden import matrix in ARCHITECTURE.md matches depguard rules in .golangci.yml
#
# Exit code 0 = all checks pass
# Exit code 1 = drift detected

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ARCH_MD="$REPO_ROOT/ARCHITECTURE.md"
GOLANGCI="$REPO_ROOT/.golangci.yml"
ERRORS=0

error() {
    echo "ERROR: $1"
    ERRORS=$((ERRORS + 1))
}

# --- Check 1: Every package under apps/<app>/internal/ is mentioned in ARCHITECTURE.md ---

echo "==> Checking app internal packages are documented..."

for app_dir in "$REPO_ROOT"/apps/*/; do
    [ -d "$app_dir" ] || continue
    app_name="$(basename "$app_dir")"
    internal_dir="$app_dir/internal"
    [ -d "$internal_dir" ] || continue

    while IFS= read -r dir; do
        [ -z "$dir" ] && continue
        # Get the package name relative to internal/
        pkg="${dir#"$internal_dir/"}"

        # Skip nested packages — only check top-level and one level deep
        depth=$(echo "$pkg" | tr '/' '\n' | wc -l)
        [ "$depth" -gt 2 ] && continue

        # Check if this package or its parent is mentioned
        parent="${pkg%%/*}"
        if ! grep -qF "$pkg" "$ARCH_MD" && ! grep -qF "$parent/" "$ARCH_MD"; then
            error "apps/$app_name/internal/$pkg/ exists but is not mentioned in ARCHITECTURE.md"
        fi
    done < <(find "$internal_dir" -type d -mindepth 1)
done

# --- Check 2: Every FX module in core/fx/ is referenced in ARCHITECTURE.md ---

echo "==> Checking core/fx/ modules are documented..."

while IFS= read -r dir; do
    [ -z "$dir" ] && continue
    module="$(basename "$dir")"

    if ! grep -qF "$module" "$ARCH_MD"; then
        error "core/fx/$module/ exists but is not mentioned in ARCHITECTURE.md"
    fi
done < <(find "$REPO_ROOT/core/fx" -type d -mindepth 1 -maxdepth 1)

# --- Check 3: Depguard layers match ARCHITECTURE.md forbidden import matrix ---

echo "==> Checking depguard rules match ARCHITECTURE.md..."

# Extract layer names from depguard config
DEPGUARD_LAYERS=$(grep -E '^\s+[a-z]+-layer:' "$GOLANGCI" | sed 's/^\s*//;s/-layer:.*//' | sort || true)

# Check each depguard layer has a corresponding section in ARCHITECTURE.md
for layer in $DEPGUARD_LAYERS; do
    case "$layer" in
        domain) expected="domain/" ;;
        service) expected="service/" ;;
        infrastructure) expected="infrastructure/" ;;
        transport) expected="transport/" ;;
        *) expected="$layer/" ;;
    esac

    # Check the forbidden import matrix mentions this layer
    if ! grep -q "| \`$expected\`" "$ARCH_MD"; then
        error "depguard has '$layer-layer' rules but ARCHITECTURE.md forbidden import matrix doesn't list '$expected'"
    fi
done

# Check ARCHITECTURE.md layers have corresponding depguard rules
for layer in domain service infrastructure transport; do
    if ! grep -q "${layer}-layer:" "$GOLANGCI"; then
        error "ARCHITECTURE.md lists '$layer/' in forbidden import matrix but .golangci.yml has no '${layer}-layer' depguard rule"
    fi
done

# --- Check 4: QUALITY.md references valid packages and tools ---

echo "==> Checking QUALITY.md for stale references..."

QUALITY_MD="$REPO_ROOT/docs/QUALITY.md"
if [ -f "$QUALITY_MD" ]; then
    # Extract package references from grade headers and table entries
    while IFS= read -r ref; do
        [ -z "$ref" ] && continue
        # Check if the referenced FX module exists
        if [[ "$ref" == *fx ]]; then
            if [ ! -d "$REPO_ROOT/core/fx/$ref" ]; then
                error "QUALITY.md references '$ref' but core/fx/$ref/ does not exist. Remove or update the grade entry."
            fi
        fi
    done < <(grep -oE '\([a-z]+fx\)' "$QUALITY_MD" | sed 's/^(//;s/)$//' || true)

    # Check core/fx/ modules have grade entries
    while IFS= read -r dir; do
        [ -z "$dir" ] && continue
        module="$(basename "$dir")"
        if ! grep -qF "$module" "$QUALITY_MD"; then
            error "core/fx/$module/ exists but has no grade entry in QUALITY.md"
        fi
    done < <(find "$REPO_ROOT/core/fx" -type d -mindepth 1 -maxdepth 1)
else
    error "docs/QUALITY.md does not exist."
fi

# --- Summary ---

if [ "$ERRORS" -gt 0 ]; then
    echo ""
    echo "Architecture drift detected: $ERRORS error(s)."
    exit 1
else
    echo ""
    echo "All architecture checks passed."
    exit 0
fi
