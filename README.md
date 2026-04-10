# AgentLayer

AgentLayer is an open-source email communication layer for external AI agents.

V0 focuses on one strict core loop:

1. Receive inbound email through the SMTP edge.
2. Persist the raw message durably.
3. Parse and resolve inbox, contact, and thread context.
4. Deliver a signed webhook to an external agent runtime.
5. Accept a reply action and send the threaded response through the outbound provider.
6. Track delivery and bounce lifecycle state.

The repository is being built in small, commit-friendly slices so core contracts, replayability, idempotency, and threading correctness stay stable as the system grows.

## Local Dev

Go version:

- Go `1.24+` is now required because the SES provider uses the latest `aws-sdk-go-v2` packages.

Run the current V0 server:

```bash
go run ./cmd/agentlayer
```

Default listeners:

- HTTP: `:8080`
- SMTP: `localhost:2525`

Optional env vars:

- `AGENTLAYER_ADDR`
- `AGENTLAYER_SMTP_ADDR`
- `AGENTLAYER_SMTP_DOMAIN`
- `AGENTLAYER_WEBHOOK_URL`
- `AGENTLAYER_WEBHOOK_SECRET`
- `AGENTLAYER_DATABASE_URL`
- `AGENTLAYER_RAW_STORE`
- `AGENTLAYER_RAW_DATA_DIR`
- `AGENTLAYER_S3_BUCKET`
- `AGENTLAYER_S3_ENDPOINT`
- `AGENTLAYER_S3_PATH_STYLE`
- `AGENTLAYER_S3_ACCESS_KEY_ID`
- `AGENTLAYER_S3_SECRET_ACCESS_KEY`
- `AGENTLAYER_AUTO_MIGRATE`
- `AGENTLAYER_EMAIL_PROVIDER`
- `AGENTLAYER_WEBHOOK_RETRY_ENABLED`
- `AGENTLAYER_WEBHOOK_RETRY_INTERVAL`
- `AGENTLAYER_WEBHOOK_RETRY_LIMIT`
- `AWS_REGION`

Helper files in the repo:

- [`compose.yaml`](/home/fayeed/dev/agent-layer/compose.yaml)
- [`.env.example`](/home/fayeed/dev/agent-layer/.env.example)
- [`Makefile`](/home/fayeed/dev/agent-layer/Makefile)
- [`cmd/agentlayerctl/main.go`](/home/fayeed/dev/agent-layer/cmd/agentlayerctl/main.go)
- [`scripts/bootstrap_local_runtime.sh`](/home/fayeed/dev/agent-layer/scripts/bootstrap_local_runtime.sh)
- [`scripts/show_local_runtime.sh`](/home/fayeed/dev/agent-layer/scripts/show_local_runtime.sh)

## Postgres Mode

The server can now run with Postgres-backed structured state and either:

- local filesystem raw MIME storage
- S3/MinIO-compatible raw MIME storage

Example:

```bash
export AGENTLAYER_DATABASE_URL='postgres://agentlayer:agentlayer@localhost:5432/agentlayer?sslmode=disable'
export AGENTLAYER_RAW_DATA_DIR='.agentlayer-data/raw'
export AGENTLAYER_AUTO_MIGRATE='true'

go run ./cmd/agentlayer
```

Or use the local helpers:

```bash
make postgres-up
make run-postgres
```

For local MinIO-backed raw storage:

```bash
make runtime-up
make run-postgres-s3
```

Notes:

- `AGENTLAYER_AUTO_MIGRATE=true` applies the embedded `db/migrations/0001_v0_core.sql` schema on startup.
- `AGENTLAYER_RAW_STORE=fs` writes raw MIME under `AGENTLAYER_RAW_DATA_DIR`.
- `AGENTLAYER_RAW_STORE=s3` writes raw MIME to `AGENTLAYER_S3_BUCKET`.
- `AGENTLAYER_S3_ENDPOINT` and `AGENTLAYER_S3_PATH_STYLE=true` make local MinIO work cleanly.
- If `AGENTLAYER_DATABASE_URL` is unset, the server falls back to the in-memory runtime store.
- The bundled Postgres container uses database `agentlayer` with user/password `agentlayer`.
- The bundled MinIO container uses `minioadmin` / `minioadmin` and exposes:
  - S3 API: `http://localhost:9000`
  - Console: `http://localhost:9001`

## Email Provider Mode

The runtime currently supports:

- `AGENTLAYER_EMAIL_PROVIDER=dev`
- `AGENTLAYER_EMAIL_PROVIDER=ses`

SES mode uses the latest `aws-sdk-go-v2` client and expects standard AWS credentials plus `AWS_REGION`.
The server now fails fast on boot if `AGENTLAYER_EMAIL_PROVIDER=ses` is selected without `AWS_REGION`.

## Webhook Retry Worker

Failed webhook deliveries now retry automatically in the background.

Relevant env vars:

- `AGENTLAYER_WEBHOOK_RETRY_ENABLED`
- `AGENTLAYER_WEBHOOK_RETRY_INTERVAL`
- `AGENTLAYER_WEBHOOK_RETRY_LIMIT`

You can still trigger a manual sweep with:

```bash
curl -X POST 'http://localhost:8080/webhooks/deliveries/retry?limit=20'
```

Example:

```bash
export AGENTLAYER_EMAIL_PROVIDER='ses'
export AWS_REGION='us-east-1'

go run ./cmd/agentlayer
```

## Local Walkthrough

Bootstrap the in-memory local runtime:

```bash
curl -X POST http://localhost:8080/bootstrap \
  -H 'Content-Type: application/json' \
  -d '{
    "organization_name":"Acme Support",
    "agent_name":"Acme Agent",
    "agent_status":"active",
    "webhook_url":"http://localhost:3000/webhook",
    "webhook_secret":"dev-secret",
    "inbox_address":"agent@localhost",
    "inbox_domain":"localhost",
    "inbox_display_name":"Acme Inbox"
  }'
```

Inspect the current bootstrap config:

```bash
curl http://localhost:8080/bootstrap
```

Or with the bundled helpers:

```bash
make bootstrap-local
make ready-local
make show-local
make retry-webhooks-local
make send-sample
```

Inspect webhook delivery activity:

```bash
curl http://localhost:8080/webhooks/deliveries
curl 'http://localhost:8080/webhooks/deliveries?limit=5'
curl http://localhost:8080/webhooks/deliveries/<delivery-id>
```

Replay a stored webhook delivery:

```bash
curl -X POST http://localhost:8080/webhooks/deliveries/<delivery-id>/replay
```

Check readiness and trigger a manual retry sweep:

```bash
curl http://localhost:8080/readyz
curl -X POST 'http://localhost:8080/webhooks/deliveries/retry?limit=20'
```

Current local admin/runtime endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /bootstrap`
- `POST /bootstrap`
- `GET /inbound/receipts`
- `GET /inbound/receipts/list`
- `POST /inbound/reprocess`
- `GET /webhooks/deliveries`
- `POST /webhooks/deliveries/retry`
- `GET /webhooks/deliveries/{deliveryID}`
- `POST /webhooks/deliveries/{deliveryID}/replay`
- `GET /threads/{threadID}`
- `GET /threads/{threadID}/messages`
- `POST /threads/{threadID}/reply`
- `POST /threads/{threadID}/escalate`
- `GET /contacts/{contactID}`
- `POST /contacts/{contactID}/memory`
- `POST /provider/callbacks/outbound`

Inbound receipt admin flow:

```bash
curl 'http://localhost:8080/inbound/receipts?object_key=raw/2026/04/09/inbox-local/example.eml'
curl 'http://localhost:8080/inbound/receipts/list?limit=10'
curl -X POST http://localhost:8080/inbound/reprocess \
  -H 'Content-Type: application/json' \
  -d '{"object_key":"raw/2026/04/09/inbox-local/example.eml"}'
```

## Schema

The first Postgres schema for the V0 core loop lives at:

- `db/migrations/0001_v0_core.sql`

Schema notes are documented in:

- `docs/storage/v0-schema.md`

Current launch-gap checklist:

- [`docs/plans/v0-remaining-checklist.md`](/home/fayeed/dev/agent-layer/docs/plans/v0-remaining-checklist.md)
