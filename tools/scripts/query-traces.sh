#!/usr/bin/env bash
set -euo pipefail

# Query Grafana Tempo using TraceQL
# Usage: ./query-traces.sh '{resource.service.name="env"}'
# Docs: https://grafana.com/docs/tempo/latest/traceql/

TEMPO_URL="${TEMPO_URL:-http://localhost:3200}"
QUERY="${1:?Usage: query-traces.sh \"<TraceQL query>\"}"
LIMIT="${LIMIT:-20}"

curl -sf "${TEMPO_URL}/api/search" \
  --data-urlencode "q=${QUERY}" \
  --data-urlencode "limit=${LIMIT}" | jq .
