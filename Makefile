.PHONY: postgres-up postgres-down runtime-up runtime-down run run-postgres run-postgres-s3 bootstrap-local ready-local show-local retry-webhooks-local send-sample test

postgres-up:
	docker compose up -d postgres

runtime-up:
	docker compose up -d postgres minio

postgres-down:
	docker compose down

runtime-down:
	docker compose down

run:
	go run ./cmd/agentlayer

run-postgres:
	set -a && . ./.env.example && set +a && go run ./cmd/agentlayer

run-postgres-s3:
	set -a && . ./.env.example && set +a && AGENTLAYER_RAW_STORE=s3 go run ./cmd/agentlayer

bootstrap-local:
	go run ./cmd/agentlayerctl bootstrap

ready-local:
	go run ./cmd/agentlayerctl ready

show-local:
	go run ./cmd/agentlayerctl show

retry-webhooks-local:
	go run ./cmd/agentlayerctl retry-webhooks

send-sample:
	go run ./cmd/agentlayerctl send-sample

test:
	go test ./...
