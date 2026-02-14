#!/usr/bin/env bash
set -euo pipefail

# Query Victoria Logs using LogsQL
# Usage: ./query-logs.sh "service:env AND error"
# Docs: https://docs.victoriametrics.com/victorialogs/logsql/

VICTORIA_LOGS_URL="${VICTORIA_LOGS_URL:-http://localhost:9428}"
QUERY="${1:?Usage: query-logs.sh \"<LogsQL query>\"}"
LIMIT="${LIMIT:-100}"

curl -sf "${VICTORIA_LOGS_URL}/select/logsql/query" \
  --data-urlencode "query=${QUERY}" \
  --data-urlencode "limit=${LIMIT}" | jq .
