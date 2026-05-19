#!/bin/sh
set -e

WORKSPACE="/home/coder/workspace"

if [ -n "$GIT_REPO" ]; then
  echo "Cloning $GIT_REPO into $WORKSPACE ..."
  cd "$WORKSPACE"
  git clone "$GIT_REPO" .
  echo "Clone complete."
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
