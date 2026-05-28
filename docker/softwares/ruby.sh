#!/bin/sh

is_installed() {
    [ -d /opt/ruby/bin ]
}

fetch_latest_ruby_version() {
    local version
    version=$(curl -fsSL 'https://endoflife.date/api/ruby.json' | jq -r '.[0].latest' 2>/dev/null)
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "4.0.5"
    else
        echo "$version"
    fi
}

install() {
    RUBY_VERSION=$(fetch_latest_ruby_version)
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

    export PATH="/opt/ruby/bin:$PATH"
}
