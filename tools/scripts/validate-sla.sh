#!/usr/bin/env bash
set -euo pipefail

# Validates SLA thresholds against the local observability stack.
# Requires: app running + observability stack (make observability-up)
#
# Thresholds are defined in docs/RELIABILITY.md.
# Exit code 0 = all checks pass
# Exit code 1 = SLA violation detected

VICTORIA_METRICS_URL="${VICTORIA_METRICS_URL:-http://localhost:8428}"
TEMPO_URL="${TEMPO_URL:-http://localhost:3200}"
SERVICE="${SERVICE:-}"
ERRORS=0

error() {
    echo "FAIL: $1"
    ERRORS=$((ERRORS + 1))
}

pass() {
    echo "PASS: $1"
}

skip() {
    echo "SKIP: $1"
}

# --- Check connectivity ---

if ! curl -sf "${VICTORIA_METRICS_URL}/api/v1/query" --data-urlencode "query=up" >/dev/null 2>&1; then
    echo "ERROR: Victoria Metrics not reachable at ${VICTORIA_METRICS_URL}"
    echo "Start the observability stack with: make observability-up"
    exit 1
fi

if ! curl -sf "${TEMPO_URL}/ready" >/dev/null 2>&1; then
    echo "ERROR: Grafana Tempo not reachable at ${TEMPO_URL}"
    echo "Start the observability stack with: make observability-up"
    exit 1
fi

if [ -z "$SERVICE" ]; then
    echo "ERROR: SERVICE environment variable required (the OTel service_name)"
    echo "Usage: SERVICE=myapp ./validate-sla.sh"
    exit 1
fi

echo "Validating SLA thresholds for service: $SERVICE"
echo ""

# --- Check 1: HTTP request p99 latency < 500ms ---

echo "==> HTTP request p99 latency (threshold: 500ms)..."

P99_RESULT=$(curl -sf "${VICTORIA_METRICS_URL}/api/v1/query" \
    --data-urlencode "query=histogram_quantile(0.99, rate(http_server_request_duration_seconds_bucket{service_name=\"${SERVICE}\"}[5m]))" \
    | jq -r '.data.result[0].value[1] // empty' 2>/dev/null || true)

if [ -z "$P99_RESULT" ]; then
    skip "No HTTP latency data available yet"
else
    P99_MS=$(echo "$P99_RESULT * 1000" | bc -l 2>/dev/null | cut -d. -f1 || echo "0")
    if [ "$P99_MS" -gt 500 ]; then
        error "HTTP p99 latency is ${P99_MS}ms (threshold: 500ms)"
    else
        pass "HTTP p99 latency is ${P99_MS}ms"
    fi
fi

# --- Check 2: Max span duration < 2s ---

echo "==> Max span duration (threshold: 2s)..."

LONG_SPANS=$(curl -sf "${TEMPO_URL}/api/search" \
    --data-urlencode "q={resource.service.name=\"${SERVICE}\" && duration > 2s}" \
    --data-urlencode "limit=5" \
    | jq -r '.traces | length' 2>/dev/null || echo "")

if [ -z "$LONG_SPANS" ]; then
    skip "No trace data available yet"
elif [ "$LONG_SPANS" -gt 0 ]; then
    error "Found ${LONG_SPANS} traces with spans exceeding 2s"
else
    pass "No spans exceeding 2s"
fi

# --- Check 3: Health probe response < 100ms ---

echo "==> Health probe response time (threshold: 100ms)..."

HEALTH_RESULT=$(curl -sf "${VICTORIA_METRICS_URL}/api/v1/query" \
    --data-urlencode "query=histogram_quantile(0.99, rate(http_server_request_duration_seconds_bucket{service_name=\"${SERVICE}\", http_route=~\"/healthz|/readyz\"}[5m]))" \
    | jq -r '.data.result[0].value[1] // empty' 2>/dev/null || true)

if [ -z "$HEALTH_RESULT" ]; then
    skip "No health probe latency data available yet"
else
    HEALTH_MS=$(echo "$HEALTH_RESULT * 1000" | bc -l 2>/dev/null | cut -d. -f1 || echo "0")
    if [ "$HEALTH_MS" -gt 100 ]; then
        error "Health probe p99 latency is ${HEALTH_MS}ms (threshold: 100ms)"
    else
        pass "Health probe p99 latency is ${HEALTH_MS}ms"
    fi
fi

# --- Summary ---

echo ""
if [ "$ERRORS" -gt 0 ]; then
    echo "SLA validation failed: $ERRORS threshold(s) exceeded."
    exit 1
else
    echo "All SLA thresholds within limits."
    exit 0
fi
