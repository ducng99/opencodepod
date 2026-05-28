#!/bin/sh

is_installed() {
    [ -d /opt/java/bin ]
}

fetch_latest_jdk_version() {
    local version
    version=$(curl -fsSL 'https://api.adoptium.net/v3/info/available_releases' | jq -r '.releases[0]' 2>/dev/null)
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "25"
    else
        echo "$version"
    fi
}

fetch_latest_maven_version() {
    local version
    version=$(curl -fsSL 'https://api.github.com/repos/apache/maven/releases/latest' | jq -r '.tag_name' 2>/dev/null | sed 's/^maven-//')
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "3.9.16"
    else
        echo "$version"
    fi
}

fetch_latest_gradle_version() {
    local version
    version=$(curl -fsSL 'https://services.gradle.org/versions/current' | jq -r '.version' 2>/dev/null)
    if [ -z "$version" ] || [ "$version" = "null" ]; then
        echo "9.5.1"
    else
        echo "$version"
    fi
}

install() {
    JAVA_VERSION=$(fetch_latest_jdk_version)
    echo "Downloading JDK ${JAVA_VERSION}..."
    curl -fsSL "https://api.adoptium.net/v3/binary/latest/${JAVA_VERSION}/ga/linux/x64/jdk/hotspot/normal/eclipse" -o /tmp/java.tar.gz
    sudo mkdir -p /opt/java-extract
    sudo tar -xzf /tmp/java.tar.gz -C /opt/java-extract
    sudo mv /opt/java-extract/jdk-* /opt/java
    sudo rm -rf /opt/java-extract
    rm /tmp/java.tar.gz

    MAVEN_VERSION=$(fetch_latest_maven_version)
    echo "Downloading Maven ${MAVEN_VERSION}..."
    curl -fsSL "https://dlcdn.apache.org/maven/maven-3/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz" -o /tmp/maven.tar.gz
    sudo mkdir -p /opt/maven
    sudo tar -xzf /tmp/maven.tar.gz -C /opt/maven --strip-components=1
    rm /tmp/maven.tar.gz

    GRADLE_VERSION=$(fetch_latest_gradle_version)
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

    export PATH="/opt/java/bin:$PATH"
}
