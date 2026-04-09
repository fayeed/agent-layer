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
- `AGENTLAYER_RAW_DATA_DIR`
- `AGENTLAYER_AUTO_MIGRATE`

Helper files in the repo:

- [`compose.yaml`](/home/fayeed/dev/agent-layer/compose.yaml)
- [`.env.example`](/home/fayeed/dev/agent-layer/.env.example)
- [`Makefile`](/home/fayeed/dev/agent-layer/Makefile)

## Postgres Mode

The server can now run with Postgres-backed structured state and local filesystem raw MIME storage.

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

Notes:

- `AGENTLAYER_AUTO_MIGRATE=true` applies the embedded `db/migrations/0001_v0_core.sql` schema on startup.
- Raw MIME files are written under `AGENTLAYER_RAW_DATA_DIR`.
- If `AGENTLAYER_DATABASE_URL` is unset, the server falls back to the in-memory runtime store.
- The bundled Postgres container uses database `agentlayer` with user/password `agentlayer`.

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

Current local admin/runtime endpoints:

- `GET /healthz`
- `GET /bootstrap`
- `POST /bootstrap`
- `GET /inbound/receipts`
- `GET /inbound/receipts/list`
- `POST /inbound/reprocess`
- `GET /webhooks/deliveries`
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
