#!/bin/bash
# send.sh — Send a message to another Agent's device.
#
# Usage: send.sh <device-id> <text> [skill-name]
#
# Arguments:
#   device-id   Target device identifier
#   text        Message content
#   skill-name  (Optional) Associated skill name for context

set -euo pipefail

DEVICE_ID="${1:?Usage: send.sh <device-id> <text> [skill-name]}"
TEXT="${2:?Usage: send.sh <device-id> <text> [skill-name]}"
SKILL="${3:-}"

if [ -n "$SKILL" ]; then
    fyy message send "$DEVICE_ID" --text "$TEXT" --skill "$SKILL"
else
    fyy message send "$DEVICE_ID" --text "$TEXT"
fi
