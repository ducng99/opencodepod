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
    openssh-server \
    ca-certificates \
    gnupg \
    unzip \
    zip \
    jq \
    htop \
    tree \
    sudo

# Generate SSH host keys and harden sshd config
RUN ssh-keygen -A \
    && mkdir -p /var/run/sshd \
    && sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
    && sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config \
    && echo "AllowUsers coder" >> /etc/ssh/sshd_config

# Create user 'coder' with home directory and passwordless sudo
RUN useradd -m -s /bin/bash coder \
    && usermod -aG sudo coder \
    && echo "coder ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/coder \
    && chmod 0440 /etc/sudoers.d/coder

# Ensure workspace and .ssh dir exist and are owned by coder
RUN mkdir -p /home/coder/workspace /home/coder/.ssh \
    && chown -R coder:coder /home/coder

# Install opencode as user 'coder'
USER coder
ENV PATH="/home/coder/.opencode/bin:${PATH}"
RUN curl -fsSL https://opencode.ai/install | bash
COPY --chown=coder:coder opencode.jsonc /home/coder/.config/opencode/opencode.jsonc

COPY --chown=coder:coder entrypoint.sh /home/coder/entrypoint.sh

WORKDIR /home/coder/workspace

ENTRYPOINT ["bash", "/home/coder/entrypoint.sh"]
CMD ["opencode", "web"]
