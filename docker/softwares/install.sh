#!/bin/zsh
set -e

SOFT_DIR="$(cd "$(dirname "$0")" && pwd)"
STACKS_FILE="/opt/.stacks.json"

if [ ! -f "$STACKS_FILE" ]; then
    echo "No stacks.json found, skipping software installation."
    exit 0
fi

sudo chown ubuntu:ubuntu "$STACKS_FILE"

DESIRED=$(jq -r '.stacks[]' "$STACKS_FILE" 2>/dev/null || echo "")

for stack in $DESIRED; do
    SCRIPT="$SOFT_DIR/$stack.sh"
    if [ ! -f "$SCRIPT" ]; then
        echo "Unknown stack: $stack (no $SCRIPT found)"
        continue
    fi

    # Source the software file to load is_installed, install, and setup_env functions
    . "$SCRIPT"

    # Always ensure env file has PATH entries, even if software is already installed
    setup_env

    if is_installed; then
        echo "Stack '$stack' already installed, skipping."
        continue
    fi

    echo "Installing stack: $stack"
    install
done
