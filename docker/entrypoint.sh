#!/bin/sh
set -e

WORKSPACE="/workspace"

if [ -n "$GIT_REPO" ]; then
  cd "$WORKSPACE"
  if [ ! -d ".git" ]; then
    echo "Cloning $GIT_REPO into $WORKSPACE ..."
    git clone "$GIT_REPO" .
    echo "Clone complete."
  else
    echo "Workspace already initialised, skipping clone."
  fi
fi

# If a custom opencode config was bind-mounted by the orchestrator, copy it into place.
# Using ~ here lets this work regardless of the actual username (e.g. custom images).
if [ -f /etc/opencode-config.jsonc ]; then
  mkdir -p ~/.config/opencode
  cp /etc/opencode-config.jsonc ~/.config/opencode/opencode.jsonc
fi

# Install the provided SSH public key for the coder user
if [ -n "$SSH_PUBLIC_KEY" ]; then
  mkdir -p /home/coder/.ssh
  echo "$SSH_PUBLIC_KEY" > /home/coder/.ssh/authorized_keys
  chmod 700 /home/coder/.ssh
  chmod 600 /home/coder/.ssh/authorized_keys
  echo "SSH public key installed."
fi

# Start the SSH daemon in the background
if [ -x /usr/sbin/sshd ]; then
  sudo /usr/sbin/sshd
  echo "sshd started."
fi

opencode upgrade || true

exec "$@"
