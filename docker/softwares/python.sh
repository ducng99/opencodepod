#!/bin/sh

is_installed() {
    [ -d /opt/python/bin ]
}

fetch_latest_python_version() {
    local version
    version=$(curl -fsSL 'https://endoflife.date/api/python.json' | jq -r '.[0].latest' 2>/dev/null)
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "3.14.3"
    else
        echo "$version"
    fi
}

install() {
    if ! command -v uv >/dev/null 2>&1; then
        echo "Installing uv..."
        mkdir -p /opt/uv
        curl -LsSf https://astral.sh/uv/install.sh | env UV_INSTALL_DIR=/opt/uv sh
        UV_BIN="/opt/uv/uv"
    else
        UV_BIN="$(which uv)"
    fi

    PYTHON_FULL_VERSION=$(fetch_latest_python_version)
    PYTHON_MINOR="${PYTHON_FULL_VERSION%.*}"
    echo "Installing Python ${PYTHON_MINOR} via uv..."
    "$UV_BIN" python install "${PYTHON_MINOR}" --install-dir /opt/python

    mkdir -p /opt/python/bin
    ln -sf "/opt/python/bin/python${PYTHON_MINOR}" /opt/python/bin/python3
    ln -sf "/opt/python/bin/python${PYTHON_MINOR}" /opt/python/bin/python

    /opt/python/bin/python3 -m ensurepip --upgrade || true
    ln -sf /opt/python/bin/pip3 /opt/python/bin/pip 2>/dev/null || true
    echo "Python stack installed."
}
