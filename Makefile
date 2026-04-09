.PHONY: postgres-up postgres-down run run-postgres bootstrap-local show-local test

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose down

run:
	go run ./cmd/agentlayer

run-postgres:
	set -a && . ./.env.example && set +a && go run ./cmd/agentlayer

bootstrap-local:
	./scripts/bootstrap_local_runtime.sh

show-local:
	./scripts/show_local_runtime.sh

test:
	go test ./...
