#!/bin/sh
set -e

WORKSPACE="/workspaces"

# Configure Git SSH key if present
if [ -f /home/coder/.ssh/id_ed25519 ]; then
  sudo chown coder:coder /home/coder/.ssh/id_ed25519
  sudo chmod 600 /home/coder/.ssh/id_ed25519
  export GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=accept-new -i /home/coder/.ssh/id_ed25519"
  echo "Git SSH key configured."
fi

if [ -n "$GIT_REPO" ]; then
  REPO_NAME=$(basename "$GIT_REPO" .git)
  REPO_DIR="$WORKSPACE/$REPO_NAME"
  if [ ! -d "$REPO_DIR/.git" ]; then
    echo "Cloning $GIT_REPO into $REPO_DIR ..."
    git clone "$GIT_REPO" "$REPO_DIR"
    echo "Clone complete."
  else
    echo "$REPO_NAME already cloned, skipping clone."
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

  sudo chown -R coder:coder /home/coder/.ssh
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
