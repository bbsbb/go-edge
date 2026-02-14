#!/usr/bin/env bash
set -euo pipefail

# Generates docs/generated/db-schema.md by snapshotting the actual database schema.
# Requires a running PostgreSQL instance with migrations applied.
#
# Iterates over all apps in apps/ and reads the database name from each app's
# development.yaml configuration (database.database field).
#
# Uses the root/migration user (not the app user) to capture the full schema
# including RLS policies and grants.
#
# Environment variables (with defaults for local dev):
#   PGHOST     (default: localhost)
#   PGPORT     (default: 5432)
#   PGUSER     (default: root)
#   PGPASSWORD (default: root)

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUTPUT="$REPO_ROOT/docs/generated/db-schema.md"

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-root}"
export PGHOST PGPORT PGUSER PGPASSWORD="${PGPASSWORD:-root}"

# Collect database names from app configs
DATABASES=()
for app_dir in "$REPO_ROOT"/apps/*/; do
    [ -d "$app_dir" ] || continue
    config="$app_dir/resources/config/development.yaml"
    [ -f "$config" ] || continue

    db_name=$(grep -A5 '^database:' "$config" | grep '^\s*database:' | head -1 | sed 's/.*database:\s*//' | tr -d '"' | tr -d "'" | xargs)
    if [ -n "$db_name" ]; then
        DATABASES+=("$db_name")
    fi
done

if [ ${#DATABASES[@]} -eq 0 ]; then
    echo "No app database configurations found in apps/*/resources/config/development.yaml"
    echo "Generating placeholder."
    {
        echo "<!-- GENERATED FILE — do not edit manually. Run 'make docs-schema' to regenerate. -->"
        echo "# Database Schema"
        echo ""
        echo "Run \`make docs-schema\` with a local PostgreSQL instance to generate this file."
    } > "$OUTPUT"
    exit 0
fi

if ! pg_isready -q 2>/dev/null; then
    echo "Error: PostgreSQL is not reachable at $PGHOST:$PGPORT"
    echo "Start a local database or set PGHOST/PGPORT/PGUSER/PGPASSWORD."
    exit 1
fi

{
    echo "<!-- GENERATED FILE — do not edit manually. Run 'make docs-schema' to regenerate. -->"
    echo "# Database Schema"
    echo ""
    echo "Snapshot of the \`app\` schema from local databases with all migrations applied."
} > "$OUTPUT"

for db in "${DATABASES[@]}"; do
    export PGDATABASE="$db"

    SCHEMA_DDL="$(pg_dump --schema-only --schema=app --no-owner --no-privileges --no-comments 2>/dev/null || true)"

    if [ -z "$SCHEMA_DDL" ]; then
        echo "Warning: could not dump schema 'app' from database '$db'. Are migrations applied?"
        continue
    fi

    {
        echo ""
        echo "## $db"
        echo ""
        echo '```sql'
        echo "$SCHEMA_DDL"
        echo '```'
    } >> "$OUTPUT"
done

echo "Generated $OUTPUT"
