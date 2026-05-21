#!/bin/sh
set -e

WORKSPACE="/workspaces"

# Ensure dirs are owned by coder
sudo chown coder:coder /home/coder/.local /home/coder/.local/share /home/coder/.local/share/opencode

opencode upgrade || true

# Configure Git SSH key if present
if [ -f /home/coder/.ssh/id_ed25519 ]; then
  sudo chown coder:coder /home/coder/.ssh/id_ed25519
  sudo chmod 600 /home/coder/.ssh/id_ed25519
  echo "Git SSH key configured."
fi

# Configure Git user identity if provided and not already set
if [ -n "$GIT_USER_NAME" ] && [ -z "$(git config --global user.name)" ]; then
  git config --global user.name "$GIT_USER_NAME"
  echo "Git user.name configured."
fi

if [ -n "$GIT_USER_EMAIL" ] && [ -z "$(git config --global user.email)" ]; then
  git config --global user.email "$GIT_USER_EMAIL"
  echo "Git user.email configured."
fi

# Configure GPG signing if key is provided
if [ -f /home/coder/.gnupg/private.key ]; then
  sudo chown -R coder:coder /home/coder/.gnupg
  chmod 700 /home/coder/.gnupg
  chmod 600 /home/coder/.gnupg/private.key

  # Import the key
  gpg --import /home/coder/.gnupg/private.key 2>/dev/null || true

  # Configure git to use GPG signing
  if [ -n "$GIT_GPG_KEY_ID" ] && [ -z "$(git config --global user.signingkey)" ]; then
    git config --global user.signingkey "$GIT_GPG_KEY_ID"
    git config --global commit.gpgsign true
    echo "GPG signing configured."
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

cd "$WORKSPACE"

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

  cd "$REPO_DIR"
fi

exec "$@"
