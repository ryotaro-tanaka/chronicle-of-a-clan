#!/usr/bin/env bash
set -euo pipefail

MESSAGE_FILE=${1:?"message file is required"}

if [[ ! -s "$MESSAGE_FILE" ]]; then
  echo "Message file is empty. Skip Threads post."
  exit 0
fi

: "${THREADS_USER_ID:?THREADS_USER_ID is required}"
: "${THREADS_ACCESS_TOKEN:?THREADS_ACCESS_TOKEN is required}"

TEXT=$(cat "$MESSAGE_FILE")

CREATE_RESPONSE=$(curl -sS -X POST "https://graph.threads.net/v1.0/${THREADS_USER_ID}/threads" \
  -F "media_type=TEXT" \
  -F "text=${TEXT}" \
  -F "access_token=${THREADS_ACCESS_TOKEN}")

CREATION_ID=$(echo "$CREATE_RESPONSE" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
if [[ -z "$CREATION_ID" ]]; then
  echo "Failed to create Threads container: $CREATE_RESPONSE"
  exit 1
fi

PUBLISH_RESPONSE=$(curl -sS -X POST "https://graph.threads.net/v1.0/${THREADS_USER_ID}/threads_publish" \
  -F "creation_id=${CREATION_ID}" \
  -F "access_token=${THREADS_ACCESS_TOKEN}")

if ! echo "$PUBLISH_RESPONSE" | grep -q '"id"'; then
  echo "Failed to publish Threads post: $PUBLISH_RESPONSE"
  exit 1
fi

echo "Threads post published."
