.PHONY: postgres-up postgres-down run run-postgres bootstrap-local show-local send-sample test

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose down

run:
	go run ./cmd/agentlayer

run-postgres:
	set -a && . ./.env.example && set +a && go run ./cmd/agentlayer

bootstrap-local:
	go run ./cmd/agentlayerctl bootstrap

show-local:
	go run ./cmd/agentlayerctl show

send-sample:
	go run ./cmd/agentlayerctl send-sample

test:
	go test ./...
