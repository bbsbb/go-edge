#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

compute_content_hash() {
    grep -v 'last-reviewed:' "$1" | sha256sum | cut -c1-8
}

update_file() {
    local file="$1"
    local rel_path="${file#"$REPO_ROOT/"}"

    tag_line=$(grep -E '<!-- last-reviewed:.*-->' "$file" | head -1 || true)
    if [ -z "$tag_line" ]; then
        echo "SKIP: $rel_path (no last-reviewed tag)"
        return
    fi

    local hash
    hash=$(compute_content_hash "$file")

    local date
    date=$(echo "$tag_line" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -1)

    local new_tag="<!-- last-reviewed: $date content-hash: $hash -->"

    sed -i "s|$tag_line|$new_tag|" "$file"
    echo "OK: $rel_path (hash: $hash)"
}

if [ $# -gt 0 ]; then
    for arg in "$@"; do
        file="$(cd "$REPO_ROOT" && realpath "$arg")"
        update_file "$file"
    done
else
    while IFS= read -r doc; do
        [ -z "$doc" ] && continue
        rel_path="${doc#"$REPO_ROOT/"}"
        [[ "$rel_path" == docs/generated/* ]] && continue
        update_file "$doc"
    done < <(find "$REPO_ROOT/docs" "$REPO_ROOT/ARCHITECTURE.md" "$REPO_ROOT/CLAUDE.md" -name '*.md' -type f 2>/dev/null)
fi
