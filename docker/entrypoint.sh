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

# ─── Software Stack Installation ───────────────────────────────────────────────

get_installed_stacks() {
    local stacks=""
    [ -d /opt/node/bin ] && stacks="$stacks javascript"
    [ -d /opt/go/bin ] && stacks="$stacks go"
    [ -d /opt/python/bin ] && stacks="$stacks python"
    [ -d /opt/.cargo/bin ] && stacks="$stacks rust"
    [ -d /opt/java/bin ] && stacks="$stacks java"
    [ -d /opt/ruby/bin ] && stacks="$stacks ruby"
    [ -d /opt/php/bin ] && stacks="$stacks php"
    echo "$stacks"
}

install_stacks() {
    STACKS_FILE="/opt/.stacks.json"

    if [ ! -f "$STACKS_FILE" ]; then
        return 0
    fi

    DESIRED=$(jq -r '.stacks[]' "$STACKS_FILE" 2>/dev/null || echo "")

    for stack in $DESIRED; do
        case "$stack" in
            javascript) [ -d /opt/node/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            go)         [ -d /opt/go/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            python)     [ -d /opt/python/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            rust)       [ -d /opt/.cargo/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            java)       [ -d /opt/java/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            ruby)       [ -d /opt/ruby/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
            php)        [ -d /opt/php/bin ] && echo "Stack '$stack' already installed, skipping." && continue ;;
        esac

        echo "Installing stack: $stack"
        case "$stack" in
            javascript) install_javascript ;;
            go)         install_go ;;
            python)     install_python ;;
            rust)       install_rust ;;
            java)       install_java ;;
            ruby)       install_ruby ;;
            php)        install_php ;;
            *)          echo "Unknown stack: $stack" ;;
        esac
    done

    build_stack_path
}

build_stack_path() {
    local path_entries=""
    [ -d /opt/node/bin ] && path_entries="/opt/node/bin:$path_entries"
    [ -d /opt/go/bin ] && path_entries="/opt/go/bin:$path_entries"
    [ -d /opt/python/bin ] && path_entries="/opt/python/bin:$path_entries"
    [ -d /opt/.cargo/bin ] && path_entries="/opt/.cargo/bin:$path_entries"
    [ -d /opt/java/bin ] && path_entries="/opt/java/bin:$path_entries"
    [ -d /opt/ruby/bin ] && path_entries="/opt/ruby/bin:$path_entries"
    [ -d /opt/php/bin ] && path_entries="/opt/php/bin:$path_entries"

    if [ -n "$path_entries" ]; then
        export PATH="${path_entries}${PATH}"
        echo "export PATH=\"${path_entries}\$PATH\"" | sudo tee /etc/profile.d/opencode-stacks.sh > /dev/null
        sudo chmod +x /etc/profile.d/opencode-stacks.sh
    fi
}

# ─── Individual Stack Installers ──────────────────────────────────────────────

install_javascript() {
    NODE_VERSION="24.16.0"
    echo "Downloading Node.js ${NODE_VERSION}..."
    curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" -o /tmp/node.tar.xz
    sudo mkdir -p /opt/node
    sudo tar -xJf /tmp/node.tar.xz -C /opt/node --strip-components=1
    rm /tmp/node.tar.xz

    echo "Installing Bun..."
    curl -fsSL https://bun.sh/install | sudo BUN_INSTALL=/opt/bun bash
    sudo ln -sf /opt/bun/bin/bun /opt/node/bin/bun
    echo "JavaScript stack installed."
}

install_go() {
    GO_VERSION=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1 | sed 's/^go//')
    echo "Downloading Go ${GO_VERSION}..."
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
    sudo mkdir -p /opt/go-extract
    sudo tar -C /opt/go-extract -xzf /tmp/go.tar.gz
    sudo mv /opt/go-extract/go /opt/go
    sudo rm -rf /opt/go-extract
    rm /tmp/go.tar.gz

    export PATH="/opt/go/bin:$PATH"
    export GOPATH=/opt/go-tools
    go install golang.org/x/tools/gopls@latest
    go install mvdan.cc/gofumpt@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest

    sudo ln -sf /opt/go-tools/bin/gopls /opt/go/bin/gopls
    sudo ln -sf /opt/go-tools/bin/gofumpt /opt/go/bin/gofumpt
    sudo ln -sf /opt/go-tools/bin/staticcheck /opt/go/bin/staticcheck
    echo "Go stack installed."
}

install_python() {
    echo "Installing Python via uv..."
    UV_BIN="$(which uv 2>/dev/null || echo /home/ubuntu/.local/bin/uv)"
    "$UV_BIN" python install 3.14 --python-dir /opt/python

    sudo mkdir -p /opt/python/bin
    sudo ln -sf /opt/python/bin/python3.14 /opt/python/bin/python3
    sudo ln -sf /opt/python/bin/python3.14 /opt/python/bin/python

    /opt/python/bin/python3 -m ensurepip --upgrade || true
    sudo ln -sf /opt/python/bin/pip3 /opt/python/bin/pip 2>/dev/null || true
    echo "Python stack installed."
}

install_rust() {
    echo "Installing Rust..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y \
        --default-toolchain stable \
        --no-modify-path

    sudo mkdir -p /opt/.cargo /opt/.rustup
    sudo cp -r "$HOME/.cargo/"* /opt/.cargo/ 2>/dev/null || true
    sudo cp -r "$HOME/.rustup/"* /opt/.rustup/ 2>/dev/null || true

    export RUSTUP_HOME=/opt/.rustup
    export CARGO_HOME=/opt/.cargo
    export PATH="/opt/.cargo/bin:$PATH"

    rustup component add rust-analyzer

    rm -rf "$HOME/.cargo" "$HOME/.rustup"
    echo "Rust stack installed."
}

install_java() {
    JAVA_VERSION="25"
    echo "Downloading JDK ${JAVA_VERSION}..."
    curl -fsSL "https://api.adoptium.net/v3/binary/latest/${JAVA_VERSION}/ga/linux/x64/jdk/hotspot/normal/eclipse" -o /tmp/java.tar.gz
    sudo mkdir -p /opt/java-extract
    sudo tar -xzf /tmp/java.tar.gz -C /opt/java-extract
    sudo mv /opt/java-extract/jdk-* /opt/java
    sudo rm -rf /opt/java-extract
    rm /tmp/java.tar.gz

    MAVEN_VERSION="3.9.16"
    echo "Downloading Maven ${MAVEN_VERSION}..."
    curl -fsSL "https://dlcdn.apache.org/maven/maven-3/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz" -o /tmp/maven.tar.gz
    sudo mkdir -p /opt/maven
    sudo tar -xzf /tmp/maven.tar.gz -C /opt/maven --strip-components=1
    rm /tmp/maven.tar.gz

    GRADLE_VERSION="9.5.1"
    echo "Downloading Gradle ${GRADLE_VERSION}..."
    curl -fsSL "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" -o /tmp/gradle.zip
    sudo mkdir -p /opt/gradle-extract
    sudo unzip -q /tmp/gradle.zip -d /opt/gradle-extract
    sudo mv /opt/gradle-extract/gradle-${GRADLE_VERSION} /opt/gradle
    sudo rm -rf /opt/gradle-extract
    rm /tmp/gradle.zip

    sudo ln -sf /opt/maven/bin/mvn /opt/java/bin/mvn
    sudo ln -sf /opt/gradle/bin/gradle /opt/java/bin/gradle
    echo "Java stack installed."
}

install_ruby() {
    RUBY_VERSION="4.0.5"
    echo "Downloading Ruby ${RUBY_VERSION}..."
    curl -fsSL "https://cache.ruby-lang.org/pub/ruby/${RUBY_VERSION%.*}/ruby-${RUBY_VERSION}.tar.gz" -o /tmp/ruby.tar.gz
    sudo mkdir -p /opt/ruby-build
    sudo tar -xzf /tmp/ruby.tar.gz -C /opt/ruby-build

    cd /opt/ruby-build/ruby-${RUBY_VERSION}
    ./configure --prefix=/opt/ruby --disable-install-doc
    make -j$(nproc)
    sudo make install
    cd /
    sudo rm -rf /opt/ruby-build
    rm /tmp/ruby.tar.gz

    /opt/ruby/bin/gem install bundler
    echo "Ruby stack installed."
}

install_php() {
    PHP_VERSION="8.5"
    echo "Downloading PHP ${PHP_VERSION}..."
    curl -fsSL "https://github.com/shivammathur/php-builder/releases/latest/download/php-${PHP_VERSION}-linux-x86_64.tar.xz" -o /tmp/php.tar.xz
    sudo mkdir -p /opt/php
    sudo tar -xJf /tmp/php.tar.xz -C /opt/php --strip-components=1
    rm /tmp/php.tar.xz

    curl -sS https://getcomposer.org/installer | sudo php -- --install-dir=/opt/php/bin --filename=composer
    echo "PHP stack installed."
}

# Install software stacks on first launch
install_stacks

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
