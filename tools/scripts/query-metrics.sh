#!/usr/bin/env bash
set -euo pipefail

# Query Victoria Metrics using PromQL
# Usage: ./query-metrics.sh "http_server_request_duration_seconds_count"
# Docs: https://docs.victoriametrics.com/keyconcepts/#instant-query

VICTORIA_METRICS_URL="${VICTORIA_METRICS_URL:-http://localhost:8428}"
QUERY="${1:?Usage: query-metrics.sh \"<PromQL query>\"}"

curl -sf "${VICTORIA_METRICS_URL}/api/v1/query" \
  --data-urlencode "query=${QUERY}" | jq .
