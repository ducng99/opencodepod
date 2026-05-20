FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive \
    PATH="/usr/local/go/bin:${PATH}"

RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    build-essential \
    htop \
    tree \
    sudo \
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
    ripgrep

# Generate SSH host keys and harden sshd config
RUN ssh-keygen -A \
    && mkdir -p /var/run/sshd \
    && sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
    && sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config \
    && echo "AllowUsers coder" >> /etc/ssh/sshd_config

# Install gh CLI
RUN mkdir -p -m 755 /etc/apt/keyrings \
	&& out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
	&& cat $out | tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
	&& chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
	&& mkdir -p -m 755 /etc/apt/sources.list.d \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
	&& apt-get update \
	&& apt-get install gh -y

# Install latest Go from official tarball
RUN GO_VERSION=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1 | sed 's/^go//') \
    && echo "Installing Go ${GO_VERSION}" \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

# Create user 'coder' with home directory and passwordless sudo
RUN useradd -m -s /bin/bash coder \
    && usermod -aG sudo coder \
    && echo "coder ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/coder \
    && chmod 0440 /etc/sudoers.d/coder

# Ensure workspace and .ssh dir exist and are owned by coder
RUN mkdir -p /workspaces /home/coder/.ssh \
    && chown -R coder:coder /home/coder /workspaces

# Install opencode as user 'coder'
USER coder
ENV PATH="/home/coder/.opencode/bin:${PATH}"
RUN curl -fsSL https://opencode.ai/install | bash

# Write SSH client config for GitHub / GitLab
RUN mkdir -p ~/.ssh \
    && { \
    echo "Host github.com gitlab.com"; \
    echo "    IdentityFile ~/.ssh/id_ed25519"; \
    echo "    StrictHostKeyChecking accept-new"; \
    echo "    AddKeysToAgent yes"; \
    } > ~/.ssh/config \
    && chmod 600 ~/.ssh/config

# Create systemd user service for ssh-agent
RUN mkdir -p ~/.config/systemd/user \
    && { \
        echo "[Unit]"; \
        echo "Description=SSH key agent"; \
        echo ""; \
        echo "[Service]"; \
        echo "Type=simple"; \
        echo "Environment=SSH_AUTH_SOCK=%t/ssh-agent.socket"; \
        echo "ExecStart=/usr/bin/ssh-agent -D -a \$SSH_AUTH_SOCK"; \
        echo ""; \
        echo "[Install]"; \
        echo "WantedBy=default.target"; \
    } > ~/.config/systemd/user/ssh-agent.service \
    && chmod 644 ~/.config/systemd/user/ssh-agent.service

COPY --chown=coder:coder opencode.jsonc ~/.config/opencode/opencode.jsonc

COPY --chown=coder:coder entrypoint.sh /home/coder/entrypoint.sh

WORKDIR /workspaces

ENTRYPOINT ["bash", "/home/coder/entrypoint.sh"]
CMD ["opencode", "web"]
