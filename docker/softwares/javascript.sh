#!/bin/sh

is_installed() {
    [ -d /opt/node/bin ]
}

fetch_latest_node_version() {
    local version
    version=$(curl -fsSL 'https://nodejs.org/dist/index.json' | jq -r '.[0].version' 2>/dev/null | sed 's/^v//')
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "24.16.0"
    else
        echo "$version"
    fi
}

install() {
    NODE_VERSION=$(fetch_latest_node_version)
    echo "Downloading Node.js ${NODE_VERSION}..."
    curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" -o /tmp/node.tar.xz
    sudo mkdir -p /opt/node
    sudo tar -xJf /tmp/node.tar.xz -C /opt/node --strip-components=1
    rm /tmp/node.tar.xz

    echo "Installing Bun..."
    curl -fsSL https://bun.sh/install | sudo BUN_INSTALL=/opt/bun bash
    sudo ln -sf /opt/bun/bin/bun /opt/node/bin/bun
    echo "JavaScript stack installed."

    export PATH="/opt/node/bin:$PATH"
}
