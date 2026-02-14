APP := sweetshop

include ../../Makefile.go.mk

.PHONY: test
test: migrate-test test-default

.PHONY: migrate-test
migrate-test:
	APP_ENVIRONMENT=testing go run . migrate up

.PHONY: migrate
migrate:
	APP_ENVIRONMENT=development go run . migrate up

.PHONY: migrate-reset
migrate-reset:
	APP_ENVIRONMENT=development go run . migrate reset

.PHONY: migrate-verify
migrate-verify:
	APP_ENVIRONMENT=development go run . migrate verify

.PHONY: migrate-create
migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=create_users TYPE=sql)
endif
ifndef TYPE
	$(error TYPE is required. Usage: make migrate-create NAME=create_users TYPE=sql)
endif
	go run . migrate create $(NAME) $(TYPE)

.PHONY: run
run:
	APP_ENVIRONMENT=development go run . server
