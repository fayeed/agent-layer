#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${AGENTLAYER_BASE_URL:-http://localhost:8080}"
WEBHOOK_URL="${AGENTLAYER_WEBHOOK_URL:-http://localhost:3000/webhook}"
WEBHOOK_SECRET="${AGENTLAYER_WEBHOOK_SECRET:-dev-secret}"
ORG_NAME="${AGENTLAYER_ORG_NAME:-Acme Support}"
AGENT_NAME="${AGENTLAYER_AGENT_NAME:-Acme Agent}"
AGENT_STATUS="${AGENTLAYER_AGENT_STATUS:-active}"
INBOX_ADDRESS="${AGENTLAYER_INBOX_ADDRESS:-agent@localhost}"
INBOX_DOMAIN="${AGENTLAYER_INBOX_DOMAIN:-localhost}"
INBOX_DISPLAY_NAME="${AGENTLAYER_INBOX_DISPLAY_NAME:-Acme Inbox}"

curl --fail-with-body -X POST "${BASE_URL}/bootstrap" \
  -H 'Content-Type: application/json' \
  -d "{
    \"organization_name\":\"${ORG_NAME}\",
    \"agent_name\":\"${AGENT_NAME}\",
    \"agent_status\":\"${AGENT_STATUS}\",
    \"webhook_url\":\"${WEBHOOK_URL}\",
    \"webhook_secret\":\"${WEBHOOK_SECRET}\",
    \"inbox_address\":\"${INBOX_ADDRESS}\",
    \"inbox_domain\":\"${INBOX_DOMAIN}\",
    \"inbox_display_name\":\"${INBOX_DISPLAY_NAME}\"
  }"

printf '\n'
