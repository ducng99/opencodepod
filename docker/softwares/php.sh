#!/bin/sh

is_installed() {
    [ -d /opt/php/bin ]
}

fetch_latest_php_version() {
    local version
    version=$(curl -fsSL 'https://api.github.com/repos/shivammathur/php-builder/releases/latest' | jq -r '.tag_name' 2>/dev/null | sed 's/^php-//')
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "8.5"
    else
        echo "$version"
    fi
}

install() {
    PHP_VERSION=$(fetch_latest_php_version)
    echo "Downloading PHP ${PHP_VERSION}..."
    curl -fsSL "https://github.com/shivammathur/php-builder/releases/latest/download/php-${PHP_VERSION}-linux-x86_64.tar.xz" -o /tmp/php.tar.xz
    sudo mkdir -p /opt/php
    sudo tar -xJf /tmp/php.tar.xz -C /opt/php --strip-components=1
    rm /tmp/php.tar.xz

    curl -sS https://getcomposer.org/installer | sudo php -- --install-dir=/opt/php/bin --filename=composer
    echo "PHP stack installed."

    export PATH="/opt/php/bin:$PATH"
}
