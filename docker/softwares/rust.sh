#!/bin/zsh

is_installed() {
    [ -d /opt/.cargo/bin ]
}

install() {
    echo "Installing Rust..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y \
        --default-toolchain stable \
        --no-modify-path

    mkdir -p /opt/.cargo /opt/.rustup
    cp -r "$HOME/.cargo/"* /opt/.cargo/ 2>/dev/null || true
    cp -r "$HOME/.rustup/"* /opt/.rustup/ 2>/dev/null || true

    export RUSTUP_HOME=/opt/.rustup
    export CARGO_HOME=/opt/.cargo

    rustup component add rust-analyzer

    rm -rf "$HOME/.cargo" "$HOME/.rustup"

    echo "Rust stack installed."
}

setup_env() {
    ENV_FILE="/home/ubuntu/.opencodepod/env"
    if ! grep -qF '/opt/.cargo/bin' "$ENV_FILE" 2>/dev/null; then
        echo 'export PATH="/opt/.cargo/bin:$PATH"' >> "$ENV_FILE"
    fi
}
