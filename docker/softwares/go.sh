#!/bin/sh

is_installed() {
    [ -d /opt/go/bin ]
}

install() {
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

    export PATH="/opt/go/bin:$PATH"
}
