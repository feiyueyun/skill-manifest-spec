#!/bin/bash
# list.sh — List messages in a conversation.
#
# Usage: list.sh <conversation-id> [limit]
#
# Arguments:
#   conversation-id  The conversation to list messages from
#   limit            (Optional) Maximum number of messages to return (default: 10)

set -euo pipefail

CONVERSATION_ID="${1:?Usage: list.sh <conversation-id> [limit]}"
LIMIT="${2:-10}"

fyy message list --conversation "$CONVERSATION_ID" --limit "$LIMIT"
