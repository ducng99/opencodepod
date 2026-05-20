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

# Install the provided SSH public key for the coder user
if [ -n "$SSH_PUBLIC_KEY" ]; then
  mkdir -p /home/coder/.ssh
  AUTH_KEYS="/home/coder/.ssh/authorized_keys"

  if [ ! -f "$AUTH_KEYS" ] || ! grep -qF "$SSH_PUBLIC_KEY" "$AUTH_KEYS" 2>/dev/null; then
    echo "$SSH_PUBLIC_KEY" >> "$AUTH_KEYS"
    echo "SSH public key installed."
  fi

  chmod 700 /home/coder/.ssh
  chmod 600 "$AUTH_KEYS"
fi

# Start the SSH daemon in the background
if [ -x /usr/sbin/sshd ]; then
  sudo /usr/sbin/sshd
  echo "sshd started."
fi

opencode upgrade || true

exec "$@"
