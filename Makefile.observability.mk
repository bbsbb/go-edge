.PHONY: observability-up
observability-up:
	docker compose -f development/docker-compose.observability.yml up -d

.PHONY: observability-down
observability-down:
	docker compose -f development/docker-compose.observability.yml down

.PHONY: observability-logs
observability-logs:
	docker compose -f development/docker-compose.observability.yml logs -f vector

.PHONY: query-logs
query-logs:
	./tools/scripts/query-logs.sh "$(Q)"

.PHONY: query-metrics
query-metrics:
	./tools/scripts/query-metrics.sh "$(Q)"

.PHONY: query-traces
query-traces:
	./tools/scripts/query-traces.sh "$(Q)"
