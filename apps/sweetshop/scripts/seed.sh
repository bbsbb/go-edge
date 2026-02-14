#!/usr/bin/env bash
set -euo pipefail

# Seeds a development organization into the database.
# Usage: ./scripts/seed.sh

CONTAINER="${CONTAINER:-development-postgres-1}"

docker exec "$CONTAINER" psql -U root -d app_sweetshop -v ON_ERROR_STOP=1 -c "
INSERT INTO app_sweetshop.organizations (id, name, slug)
VALUES ('01961a1a-0000-7000-8000-000000000001', 'Dev Shop', 'dev-shop')
ON CONFLICT (slug) DO NOTHING;
"

echo "Seeded organization: dev-shop"
