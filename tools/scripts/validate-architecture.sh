#!/usr/bin/env bash
set -euo pipefail

# Validates that ARCHITECTURE.md stays in sync with the actual codebase:
# 1. Every package under apps/<app>/internal/ is listed in ARCHITECTURE.md
# 2. Every FX module in core/fx/ is referenced in ARCHITECTURE.md
# 3. Forbidden import matrix consistency: ARCHITECTURE.md → depguard + architecture_test.go
# 4. Denied external packages are real dependencies (no phantom guards)
# 5. Reverse validation: depguard → ARCHITECTURE.md (no undocumented enforcement)
# 6. Depguard app-specific paths cover all discovered apps
# 7. QUALITY.md references valid packages
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

# --- Check 3: Forbidden import matrix cross-validation ---
# ARCHITECTURE.md is the source of truth. Verify depguard and architecture_test.go match.

echo "==> Checking forbidden import matrix consistency..."

# Map short names in ARCHITECTURE.md to depguard pkg patterns, test substrings,
# and go.mod search patterns. When adding a new forbidden package, add it to
# ARCHITECTURE.md and extend these maps.
declare -A PKG_TO_DEPGUARD=(
    ["pgx"]="jackc/pgx"
    ["database/sql"]="database/sql"
    ["net/http"]="net/http"
    ["chi"]="go-chi/chi"
    ["infrastructure/"]="internal/infrastructure"
    ["transport/"]="internal/transport"
    ["service/"]="internal/service"
    ["config/"]="internal/config"
)

declare -A PKG_TO_TEST=(
    ["pgx"]="jackc/pgx"
    ["database/sql"]="database/sql"
    ["net/http"]="net/http"
    ["chi"]="go-chi/chi"
    ["infrastructure/"]="internal/infrastructure"
    ["transport/"]="internal/transport"
    ["service/"]="internal/service"
    ["config/"]="internal/config"
)

# Packages that are stdlib or internal paths don't need go.mod validation.
# External packages must appear in at least one go.mod to be denied.
declare -A PKG_GOMOD_PATTERN=(
    ["pgx"]="jackc/pgx"
    ["chi"]="go-chi/chi"
)

# Collect all go.mod contents once for dependency checks
ALL_GOMODS=$(cat "$REPO_ROOT"/core/go.mod "$REPO_ROOT"/apps/*/go.mod 2>/dev/null)

# Check each layer has corresponding depguard and architecture_test rules
for layer in domain service infrastructure transport; do
    if ! grep -q "${layer}-layer:" "$GOLANGCI"; then
        error "ARCHITECTURE.md lists '$layer/' but .golangci.yml has no '${layer}-layer' depguard rule"
    fi
done

# Extract only the forbidden import matrix table (between the header and the next ### heading).
FORBIDDEN_TABLE=$(sed -n '/^### Forbidden Import Matrix/,/^### /{/^### Forbidden/d;/^### /d;p}' "$ARCH_MD")

# Parse each row of the forbidden import matrix.
# Rows look like: | `domain/` | `gorm.io`, `pgx`, `database/sql`, ... |
while IFS= read -r line; do
    [ -z "$line" ] && continue
    # Skip table header and separator rows
    [[ "$line" == "| Package"* ]] && continue
    [[ "$line" == "|---"* ]] && continue

    # Extract layer name (first column, strip backticks and slash)
    layer=$(echo "$line" | sed 's/^| `//;s|/\`.*||;s|/.*||')

    # Skip the RLS row (internal/**)
    [[ "$layer" == "internal" ]] && continue

    # Extract denied packages from second column: split on comma, strip backticks and whitespace
    denied_col=$(echo "$line" | sed 's/^[^|]*| *//;s/^[^|]*| *//;s/ *|$//')
    denied=$(echo "$denied_col" | tr ',' '\n' | sed 's/^ *`//;s/`.*$//' | grep -v '^ *$' || true)

    for pkg in $denied; do
        # Skip entries that aren't package references (prose fragments)
        [[ "$pkg" == *" "* ]] && continue

        depguard_pattern="${PKG_TO_DEPGUARD[$pkg]:-}"
        test_pattern="${PKG_TO_TEST[$pkg]:-}"

        # Skip packages not in the mapping — these are prose fragments or
        # packages intentionally not enforced via depguard (e.g. psqlfx, pgxpool
        # have their own dedicated rls-enforcement rule)
        if [ -z "$depguard_pattern" ]; then
            continue
        fi

        # Check depguard: extract the layer section using awk.
        # Stop at the next depguard rule (8-space indent + name:) or a less-indented setting.
        depguard_section=$(awk "
            /^        ${layer}-layer:/ { found=1; next }
            found && /^        [a-zA-Z]/ { exit }
            found && /^    [a-zA-Z]/ { exit }
            found { print }
        " "$GOLANGCI")

        if ! echo "$depguard_section" | grep -qF "$depguard_pattern"; then
            error ".golangci.yml ${layer}-layer is missing deny for '$pkg' (expected pattern: $depguard_pattern)"
        fi

        # Check architecture_test.go in each app
        for app_dir in "$REPO_ROOT"/apps/*/; do
            [ -d "$app_dir" ] || continue
            arch_test="$app_dir/architecture_test.go"
            [ -f "$arch_test" ] || continue

            test_section=$(sed -n "/layer:.*\"$layer\"/,/},/p" "$arch_test")
            if [ -n "$test_section" ] && ! echo "$test_section" | grep -qF "$test_pattern"; then
                app_name=$(basename "$app_dir")
                error "apps/$app_name/architecture_test.go $layer denied list is missing '$test_pattern' (for ARCHITECTURE.md entry '$pkg')"
            fi
        done
    done
done <<< "$FORBIDDEN_TABLE"

# --- Check 4: Denied external packages must be real dependencies ---
# Prevents phantom guards — if a package isn't in any go.mod, don't deny it.

echo "==> Checking denied packages are real dependencies..."

for pkg in "${!PKG_GOMOD_PATTERN[@]}"; do
    gomod_pattern="${PKG_GOMOD_PATTERN[$pkg]}"
    if ! echo "$ALL_GOMODS" | grep -qF "$gomod_pattern"; then
        error "ARCHITECTURE.md forbids '$pkg' but '$gomod_pattern' is not in any go.mod. Remove the deny rule or add the dependency."
    fi
done

# --- Check 5: Depguard deny entries must be documented in ARCHITECTURE.md ---
# Prevents undocumented enforcement — every deny rule should trace back to the matrix.

echo "==> Checking depguard rules are documented in ARCHITECTURE.md..."

# Reverse mapping: depguard pattern → ARCHITECTURE.md short name
declare -A DEPGUARD_TO_PKG=()
for pkg in "${!PKG_TO_DEPGUARD[@]}"; do
    DEPGUARD_TO_PKG["${PKG_TO_DEPGUARD[$pkg]}"]="$pkg"
done

for layer in domain service infrastructure transport; do
    depguard_section=$(awk "
        /^        ${layer}-layer:/ { found=1; next }
        found && /^        [a-zA-Z]/ { exit }
        found && /^    [a-zA-Z]/ { exit }
        found { print }
    " "$GOLANGCI")

    while IFS= read -r deny_pkg; do
        [ -z "$deny_pkg" ] && continue

        # Check if this deny pattern maps back to a known ARCHITECTURE.md entry
        matched=false
        for pattern in "${!DEPGUARD_TO_PKG[@]}"; do
            if [[ "$deny_pkg" == *"$pattern"* ]]; then
                matched=true
                break
            fi
        done

        if ! $matched; then
            error ".golangci.yml ${layer}-layer denies '$deny_pkg' but it is not mapped in the forbidden import matrix"
        fi
    done < <(echo "$depguard_section" | sed -n 's/.*pkg: "\(.*\)"/\1/p')
done

# --- Check 6: Depguard app-specific paths cover all apps ---
# When a new app is added, depguard layer rules must include deny entries for it.

echo "==> Checking depguard rules cover all apps..."

for app_dir in "$REPO_ROOT"/apps/*/; do
    [ -d "$app_dir" ] || continue
    app_name=$(basename "$app_dir")

    for layer in domain service infrastructure transport; do
        depguard_section=$(awk "
            /^        ${layer}-layer:/ { found=1; next }
            found && /^        [a-zA-Z]/ { exit }
            found && /^    [a-zA-Z]/ { exit }
            found { print }
        " "$GOLANGCI")

        # Check if any internal/ deny entry references this app
        has_internal_deny=false
        if echo "$depguard_section" | grep -qF "internal/" ; then
            if echo "$depguard_section" | grep -qF "$app_name/internal/"; then
                has_internal_deny=true
            fi
        fi

        # Only flag if the layer has internal/ deny entries but none for this app
        if echo "$depguard_section" | grep -qF "internal/" && ! $has_internal_deny; then
            error ".golangci.yml ${layer}-layer has internal/ deny entries but none for app '$app_name' (missing $app_name/internal/ paths)"
        fi
    done
done

# --- Check 7: QUALITY.md references valid packages and tools ---

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
