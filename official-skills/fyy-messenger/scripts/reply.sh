#!/bin/bash
# reply.sh — Reply to a specific message.
#
# Usage: reply.sh <message-id> <text>
#
# Arguments:
#   message-id  The ID of the message being replied to
#   text        Reply content

set -euo pipefail

MESSAGE_ID="${1:?Usage: reply.sh <message-id> <text>}"
TEXT="${2:?Usage: reply.sh <message-id> <text>}"

fyy message reply "$MESSAGE_ID" --text "$TEXT"
