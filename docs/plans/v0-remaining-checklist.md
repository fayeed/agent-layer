# V0 Remaining Checklist

This checklist tracks the main work still needed before AgentLayer V0 should be considered complete.

## Must Have Before V0

- Real Postgres-mode end-to-end verification
  - Run the current server against Postgres-backed structured state, not only repository-level tests.
  - Verify bootstrap, inbound processing, webhook delivery, reply send, callback handling, replay, and reads in one real runtime path.

- Production outbound provider
  - Add the SES provider implementation for send.
  - Make SES the real default production path instead of only the local dev provider.

- Real delivery callback path
  - Validate real SES callback payload handling for delivered, bounce, and complaint events.
  - Confirm callbacks update message delivery state and suppression state correctly in the runtime path.

- Webhook retry queue
  - Move webhook retries/backoff/dead-letter behavior onto a real async queue flow.
  - Keep delivery attempts replayable and persisted.

- Reply idempotency
  - Add a real idempotency contract on `POST /threads/{id}/reply`.
  - Ensure agent retries cannot double-send outbound mail.

- Suppression enforcement before send
  - Block outbound sends to suppressed addresses in the real send path.

- Runtime auth story
  - Add minimal authentication/authorization for agent actions and admin/debug endpoints.
  - Keep webhook signature verification as the runtime-to-agent trust mechanism.

- SMTP acceptance hardening
  - Validate listener behavior under malformed mail, large payloads, invalid recipients, and durable-persistence failures.
  - Confirm acceptance semantics match the V0 rule: durable raw persistence before SMTP success.

- Deployment and runbook docs
  - Document Postgres mode setup.
  - Document raw MIME storage expectations.
  - Document SES setup and callback configuration.
  - Document local-to-first-reply walkthrough.

## Strongly Recommended Before V0

- Object storage production path
  - Add S3/MinIO-compatible raw MIME and attachment storage for production.
  - Keep local filesystem storage as the dev path.

- Async inbound processing path
  - Move stored-message handling behind a replayable worker/queue flow.
  - Preserve current replay and duplicate-protection semantics.

- End-to-end automated tests
  - Cover bootstrap -> inbound receive -> webhook delivery -> reply -> callback -> final state.
  - Cover duplicate inbound handling and replay behavior in the real runtime path.

- Startup/config validation
  - Fail fast on missing or invalid required config.
  - Improve readiness/health semantics for DB, SMTP, and provider dependencies.

## Cleanup Before Freezing V0

- Remove stale placeholder or not-implemented scaffolding that is no longer needed.
- Normalize API validation and error mapping across all handlers.
- Tighten naming consistency between routes, services, and runtime adapters.
- Review README and local helper commands against the actual happy path.

## Already In Good Shape

- Core domain model and service boundaries
- Inbound parse/contact/thread/store flow
- Webhook build/sign/deliver/record/replay flow
- Outbound reply assembly/send/status flow
- Minimal read/write/admin API surfaces
- Postgres repository layer
- Local Postgres/dev runtime helpers
- Inbound receipt replay/debug surface

## Shortest Path To Usable V0

1. Add SES provider.
2. Add real webhook retry queue.
3. Prove the full server in Postgres mode end to end.
4. Add reply idempotency.
5. Finish deploy and operator docs.
