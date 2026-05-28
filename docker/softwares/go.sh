#!/bin/sh

is_installed() {
    [ -d /opt/go/bin ]
}

install() {
    GO_VERSION=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1 | sed 's/^go//')
    echo "Downloading Go ${GO_VERSION}..."
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
    mkdir -p /opt/go-extract
    tar -C /opt/go-extract -xzf /tmp/go.tar.gz
    mv /opt/go-extract/go /opt/go
    rm -rf /opt/go-extract
    rm /tmp/go.tar.gz

    export GOPATH=/opt/go-tools
    go install golang.org/x/tools/gopls@latest
    go install mvdan.cc/gofumpt@latest

    ln -sf /opt/go-tools/bin/gopls /opt/go/bin/gopls
    ln -sf /opt/go-tools/bin/gofumpt /opt/go/bin/gofumpt
    echo "Go stack installed."
}
