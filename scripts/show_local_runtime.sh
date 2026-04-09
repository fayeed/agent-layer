#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${AGENTLAYER_BASE_URL:-http://localhost:8080}"
WEBHOOK_LIMIT="${AGENTLAYER_WEBHOOK_LIMIT:-5}"
RECEIPT_LIMIT="${AGENTLAYER_RECEIPT_LIMIT:-5}"

printf '\n== bootstrap ==\n'
curl --fail-with-body "${BASE_URL}/bootstrap"
printf '\n\n== webhook deliveries ==\n'
curl --fail-with-body "${BASE_URL}/webhooks/deliveries?limit=${WEBHOOK_LIMIT}"
printf '\n\n== inbound receipts ==\n'
curl --fail-with-body "${BASE_URL}/inbound/receipts/list?limit=${RECEIPT_LIMIT}"
printf '\n'
