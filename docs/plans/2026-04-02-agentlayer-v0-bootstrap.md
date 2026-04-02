# AgentLayer V0 Bootstrap Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bootstrap an empty repository into a minimal, testable AgentLayer V0 foundation that can grow into the inbound email core loop one modular commit at a time.

**Architecture:** Start with the narrowest stable foundation first: Go module setup, internal package layout, and immutable core domain contracts. Build inward-out from domain and interfaces before adapters like SMTP, SES, Redis, and webhooks. Each task is intentionally small so the user can commit after every meaningful slice.

**Tech Stack:** Go, PostgreSQL, Redis, S3/MinIO-compatible blob storage, Amazon SES, HTTP webhooks

---

### Task 1: Bootstrap Go module and repository layout

**Files:**
- Create: `go.mod`
- Create: `README.md`
- Create: `cmd/agentlayer/.gitkeep`
- Create: `internal/.gitkeep`
- Create: `pkg/.gitkeep`

**Step 1: Write the failing test**

No test for this scaffold-only task.

**Step 2: Run test to verify it fails**

No-op for this scaffold-only task.

**Step 3: Write minimal implementation**

- Initialize the module as `github.com/agentlayer/agentlayer`
- Add a short README describing the V0 goal and first milestones
- Create the top-level directories needed for modular growth

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS with no packages or trivial success

**Step 5: Commit**

```bash
git add go.mod README.md cmd/agentlayer/.gitkeep internal/.gitkeep pkg/.gitkeep
git commit -m "chore: bootstrap go module and repo layout"
```

### Task 2: Add core domain types and enums

**Files:**
- Create: `internal/domain/types.go`
- Create: `internal/domain/types_test.go`

**Step 1: Write the failing test**

Add tests that assert the V0 enums and helper validation functions accept:
- thread states: `ACTIVE`, `WAITING`, `ESCALATED`, `RESOLVED`, `DORMANT`
- directions: `inbound`, `outbound`
- agent statuses: `active`, `paused`, `archived`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/...`
Expected: FAIL because the domain package does not exist yet

**Step 3: Write minimal implementation**

- Define the string-backed enum types
- Define minimal structs for `Organization`, `Agent`, `Inbox`, `Contact`, `Thread`, `Message`, `WebhookDelivery`, and `ProviderConfig`
- Add validation helpers for enum-backed fields

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/types.go internal/domain/types_test.go
git commit -m "feat: add core domain contracts"
```

### Task 3: Define service boundary interfaces

**Files:**
- Create: `internal/core/interfaces.go`
- Create: `internal/core/interfaces_test.go`

**Step 1: Write the failing test**

Add compile-oriented tests that assert these interfaces exist:
- `InboundTransport`
- `MessageParser`
- `ThreadResolver`
- `ContactResolver`
- `WebhookDispatcher`
- `EmailProvider`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/core/...`
Expected: FAIL because the interfaces do not exist yet

**Step 3: Write minimal implementation**

- Define request/response structs shared by the interfaces
- Keep the interfaces transport-agnostic and small
- Import domain types instead of duplicating shapes

**Step 4: Run test to verify it passes**

Run: `go test ./internal/core/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/core/interfaces.go internal/core/interfaces_test.go
git commit -m "feat: define core service boundaries"
```

### Task 4: Add correlation and idempotency primitives

**Files:**
- Create: `internal/platform/idempotency/keys.go`
- Create: `internal/platform/idempotency/keys_test.go`
- Create: `internal/platform/correlation/context.go`
- Create: `internal/platform/correlation/context_test.go`

**Step 1: Write the failing test**

Add tests for:
- deterministic inbound idempotency keys
- deterministic reply idempotency keys
- correlation ID insertion and retrieval from context

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/...`
Expected: FAIL because the packages do not exist yet

**Step 3: Write minimal implementation**

- Add pure helpers only; no Redis dependency yet
- Favor small stateless functions

**Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/idempotency/keys.go internal/platform/idempotency/keys_test.go internal/platform/correlation/context.go internal/platform/correlation/context_test.go
git commit -m "feat: add correlation and idempotency helpers"
```

### Task 5: Add persistence model skeletons for Postgres-backed storage

**Files:**
- Create: `internal/store/models.go`
- Create: `internal/store/models_test.go`

**Step 1: Write the failing test**

Add tests that ensure DB model structs exist for:
- organizations
- agents
- inboxes
- contacts
- threads
- messages
- message attachments
- contact memory
- webhook deliveries
- suppressed addresses
- provider configs

**Step 2: Run test to verify it fails**

Run: `go test ./internal/store/...`
Expected: FAIL because the package does not exist yet

**Step 3: Write minimal implementation**

- Add store models only, no DB driver integration yet
- Include comments for immutable-vs-mutable fields where it matters

**Step 4: Run test to verify it passes**

Run: `go test ./internal/store/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/store/models.go internal/store/models_test.go
git commit -m "feat: add persistence model skeletons"
```

### Task 6: Add the first executable binary shell

**Files:**
- Create: `cmd/agentlayer/main.go`

**Step 1: Write the failing test**

No dedicated test; verify the binary compiles.

**Step 2: Run test to verify it fails**

Run: `go build ./cmd/agentlayer`
Expected: FAIL because no main package exists

**Step 3: Write minimal implementation**

- Add a tiny `main` function
- Print a startup banner or version placeholder

**Step 4: Run test to verify it passes**

Run: `go build ./cmd/agentlayer`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/agentlayer/main.go
git commit -m "feat: add initial agentlayer binary"
```

### Task 7: Begin inbound pipeline with durable receipt contract only

**Files:**
- Create: `internal/inbound/receipt.go`
- Create: `internal/inbound/receipt_test.go`

**Step 1: Write the failing test**

Add tests for a receipt request that captures:
- smtp session ID
- envelope sender
- envelope recipients
- raw message blob key
- receipt timestamp

**Step 2: Run test to verify it fails**

Run: `go test ./internal/inbound/...`
Expected: FAIL because the inbound package does not exist yet

**Step 3: Write minimal implementation**

- Define the durable handoff contract from SMTP edge into the async pipeline
- Keep it independent of `go-smtp`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/inbound/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/inbound/receipt.go internal/inbound/receipt_test.go
git commit -m "feat: add inbound durable receipt contract"
```
