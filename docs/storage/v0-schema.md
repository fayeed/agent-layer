# V0 Storage Schema

This repo now includes a first Postgres migration at `db/migrations/0001_v0_core.sql`.

The schema matches the current V0 runtime surfaces:

- bootstrap and local config:
  - `organizations`
  - `agents`
  - `inboxes`
- inbound processing and replay:
  - `inbound_receipts`
  - `messages`
  - `threads`
  - `contacts`
  - `message_attachments`
- webhook delivery and replay:
  - `webhook_deliveries`
- outbound suppression and provider config:
  - `suppressed_addresses`
  - `provider_configs`
- communication memory and auditability:
  - `contact_memory`
  - `audit_log`

Important constraints and indexes:

- `messages(inbox_id, message_id_header)` is unique when `message_id_header` is present.
  This supports inbound deduplication by inbox plus RFC `Message-ID`.
- `messages(provider_message_id)` is unique when present.
  This supports delivery callback lookups.
- `threads(organization_id, inbox_id, contact_id, subject_normalized, last_activity_at DESC)` supports subject-plus-recency thread matching.
- `inbound_receipts(received_at DESC)` supports admin/debug receipt listing.

Notes:

- `webhook_deliveries.request_headers` and provider/audit payload fields are stored as `JSONB`.
- `references` headers are stored as `TEXT[]` in `messages.references_headers`.
- raw MIME bytes and attachment blobs remain object-storage concerns; the database stores pointers and durable receipt metadata.
