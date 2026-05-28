#!/bin/sh
set -e

SOFT_DIR="$(cd "$(dirname "$0")" && pwd)"
STACKS_FILE="/opt/.stacks.json"

if [ ! -f "$STACKS_FILE" ]; then
    echo "No stacks.json found, skipping software installation."
    exit 0
fi

DESIRED=$(jq -r '.stacks[]' "$STACKS_FILE" 2>/dev/null || echo "")

for stack in $DESIRED; do
    SCRIPT="$SOFT_DIR/$stack.sh"
    if [ ! -f "$SCRIPT" ]; then
        echo "Unknown stack: $stack (no $SCRIPT found)"
        continue
    fi

    # Source the software file to load is_installed and install functions
    . "$SCRIPT"

    if is_installed; then
        echo "Stack '$stack' already installed, skipping."
        continue
    fi

    echo "Installing stack: $stack"
    install
done

# Build and persist PATH for future shells
path_entries=""
[ -d /opt/uv ] && path_entries="/opt/uv:$path_entries"
[ -d /opt/node/bin ] && path_entries="/opt/node/bin:$path_entries"
[ -d /opt/go/bin ] && path_entries="/opt/go/bin:$path_entries"
[ -d /opt/python/bin ] && path_entries="/opt/python/bin:$path_entries"
[ -d /opt/.cargo/bin ] && path_entries="/opt/.cargo/bin:$path_entries"
[ -d /opt/java/bin ] && path_entries="/opt/java/bin:$path_entries"
[ -d /opt/ruby/bin ] && path_entries="/opt/ruby/bin:$path_entries"
[ -d /opt/php/bin ] && path_entries="/opt/php/bin:$path_entries"

if [ -n "$path_entries" ]; then
    echo "export PATH=\"${path_entries}\$PATH\"" | tee /etc/profile.d/opencode-stacks.sh > /dev/null
    chmod +x /etc/profile.d/opencode-stacks.sh

    # zsh doesn't source /etc/profile.d/, so write to .zshrc too
    ZSHRC="/home/ubuntu/.zshrc"
    if [ -f "$ZSHRC" ] && ! grep -qF 'opencode-stacks' "$ZSHRC" 2>/dev/null; then
        echo "" >> "$ZSHRC"
        echo "# Software stack paths" >> "$ZSHRC"
        echo "export PATH=\"${path_entries}\$PATH\"" >> "$ZSHRC"
    fi

    # Export for current shell (entrypoint bash subshell)
    . /etc/profile.d/opencode-stacks.sh
fi
