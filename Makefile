.PHONY: postgres-up postgres-down run run-postgres test

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose down

run:
	go run ./cmd/agentlayer

run-postgres:
	set -a && . ./.env.example && set +a && go run ./cmd/agentlayer

test:
	go test ./...
