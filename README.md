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
