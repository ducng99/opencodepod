FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    build-essential \
    python3 \
    python3-pip \
    python3-venv \
    nodejs \
    vim \
    nano \
    ssh \
    ca-certificates \
    gnupg \
    unzip \
    zip \
    jq \
    htop \
    tree \
    sudo

# Create user 'coder' with home directory and passwordless sudo
RUN useradd -m -s /bin/bash coder \
    && usermod -aG sudo coder \
    && echo "coder ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/coder \
    && chmod 0440 /etc/sudoers.d/coder

# Ensure workspace exists and is owned by coder
RUN mkdir -p /home/coder/workspace && chown -R coder:coder /home/coder

# Install opencode as user 'coder'
USER coder
ENV PATH="/home/coder/.opencode/bin:${PATH}"
RUN curl -fsSL https://opencode.ai/install | bash
COPY --chown=coder:coder opencode.jsonc /home/coder/.config/opencode/opencode.jsonc

COPY --chown=coder:coder entrypoint.sh /home/coder/entrypoint.sh

WORKDIR /home/coder/workspace

ENTRYPOINT ["bash", "/home/coder/entrypoint.sh"]
CMD ["opencode", "web"]
