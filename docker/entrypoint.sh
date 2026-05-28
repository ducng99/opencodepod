#!/bin/sh
set -e

eval "$(ssh-agent -s)"

WORKSPACE="/workspaces"

# Ensure dirs are owned by ubuntu
sudo chown ubuntu:ubuntu /home/ubuntu/.local /home/ubuntu/.local/share /home/ubuntu/.local/share/opencode /home/ubuntu/.config /home/ubuntu/.config/opencode || echo "Warning: failed to set ownership on config directories"

opencode upgrade || true

# Configure GPG signing if key is provided
if [ -f /home/ubuntu/.gnupg/private.key ]; then
  sudo chown -R ubuntu:ubuntu /home/ubuntu/.gnupg
  chmod 700 /home/ubuntu/.gnupg
  chmod 600 /home/ubuntu/.gnupg/private.key

  GPG_PASSPHRASE_FILE=""
  if [ -f /home/ubuntu/.gnupg/gpg_passphrase.key ]; then
    chmod 600 /home/ubuntu/.gnupg/gpg_passphrase.key
    GPG_PASSPHRASE_FILE="/home/ubuntu/.gnupg/gpg_passphrase.key"
  fi

  # Allow loopback pinentry for unattended signing
  mkdir -p /home/ubuntu/.gnupg
  echo "allow-loopback-pinentry" >> /home/ubuntu/.gnupg/gpg-agent.conf
  gpgconf --kill gpg-agent 2>/dev/null || true

  # Import the key
  if [ -n "$GPG_PASSPHRASE_FILE" ]; then
    gpg --batch --yes --pinentry-mode loopback --passphrase-file "$GPG_PASSPHRASE_FILE" --armor --import /home/ubuntu/.gnupg/private.key
  else
    gpg --armor --import /home/ubuntu/.gnupg/private.key 2>/dev/null || true
  fi

  # Create a wrapper so git commit -S is non-interactive
  if [ -n "$GPG_PASSPHRASE_FILE" ]; then
    sudo tee /usr/local/bin/gpg-auto > /dev/null <<'EOF'
#!/bin/sh
exec /usr/bin/gpg --batch --yes --pinentry-mode loopback --passphrase-file /home/ubuntu/.gnupg/gpg_passphrase.key "$@"
EOF
    sudo chmod +x /usr/local/bin/gpg-auto
    git config --global gpg.program /usr/local/bin/gpg-auto
  fi

  # Configure git to use GPG signing
  if [ -n "$GIT_GPG_KEY_ID" ] && [ -z "$(git config --global user.signingkey)" ]; then
    git config --global user.signingkey "$GIT_GPG_KEY_ID"
    git config --global commit.gpgsign true
    echo "GPG signing configured."
  fi
fi

# Configure Git HTTP credentials if provided
if [ -f /home/ubuntu/.git-credentials ]; then
  sudo chown ubuntu:ubuntu /home/ubuntu/.git-credentials
  chmod 600 /home/ubuntu/.git-credentials
  if [ -z "$(git config --global credential.helper)" ]; then
    git config --global credential.helper store
    echo "Git credential helper configured."
  fi
fi

# Install the provided SSH public key for the ubuntu user
if [ -n "$SSH_PUBLIC_KEY" ]; then
  mkdir -p /home/ubuntu/.ssh
  AUTH_KEYS="/home/ubuntu/.ssh/authorized_keys"

  if [ ! -f "$AUTH_KEYS" ] || ! grep -qF "$SSH_PUBLIC_KEY" "$AUTH_KEYS" 2>/dev/null; then
    echo "$SSH_PUBLIC_KEY" >> "$AUTH_KEYS"
    echo "SSH public key installed."
  fi

  sudo chown -R ubuntu:ubuntu /home/ubuntu/.ssh
  chmod 700 /home/ubuntu/.ssh
  chmod 600 "$AUTH_KEYS"
fi

# Configure Git SSH key if present
if [ -f /home/ubuntu/.ssh/id_ed25519 ]; then
  sudo chown ubuntu:ubuntu /home/ubuntu/.ssh/id_ed25519
  sudo chmod 600 /home/ubuntu/.ssh/id_ed25519
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

# Start the SSH daemon in the background
if [ -x /usr/sbin/sshd ]; then
  sudo /usr/sbin/sshd
  echo "sshd started."
fi

# Software Stack Installation
/home/ubuntu/.opencodepod/softwares/install.sh

cd "$WORKSPACE"

if [ -n "$GIT_REPO" ]; then
  REPO_NAME=$(basename "$GIT_REPO" .git)
  REPO_DIR="$WORKSPACE/$REPO_NAME"

  if [ ! -d "$REPO_DIR/.git" ]; then
    echo "Cloning $GIT_REPO into $REPO_DIR ..."
    CLONE_ARGS=""
    if [ -n "$GIT_BRANCH" ]; then
      CLONE_ARGS="$CLONE_ARGS --branch $GIT_BRANCH"
    fi
    if [ -n "$GIT_DEPTH" ]; then
      CLONE_ARGS="$CLONE_ARGS --depth $GIT_DEPTH"
    fi
    # shellcheck disable=SC2086
    git clone $CLONE_ARGS "$GIT_REPO" "$REPO_DIR"
    echo "Clone complete."
  else
    echo "$REPO_NAME already cloned, skipping clone."
  fi

  cd "$REPO_DIR"
fi

exec "$@"
