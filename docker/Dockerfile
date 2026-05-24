FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    build-essential \
    htop \
    tree \
    sudo \
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
    ripgrep \
    zsh

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
RUN useradd -m -s /bin/zsh coder \
    && usermod -aG sudo coder \
    && echo "coder ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/coder \
    && chmod 0440 /etc/sudoers.d/coder

# Ensure workspace and .ssh dir exist and are owned by coder
RUN mkdir -p /workspaces \
    && chown -R coder:coder /workspaces

# Install opencode as user 'coder'
USER coder
ENV PATH="/home/coder/.opencode/bin:/home/coder/bin:/usr/local/go/bin:${PATH}"

# Install Oh My Zsh and plugins
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
    && git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions \
    && git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting \
    && sed -i 's/plugins=(git)/plugins=(git zsh-autosuggestions zsh-syntax-highlighting)/' ~/.zshrc

RUN curl -LsSf https://astral.sh/uv/install.sh | sh \
    && echo 'eval "$(uv generate-shell-completion zsh)"' >> ~/.zshrc
RUN curl -fsSL https://bun.sh/install | bash
RUN curl -fsSL https://opencode.ai/install | bash
COPY --chown=coder:coder opencode.jsonc /home/coder/.config/opencode/opencode.jsonc

# Write SSH client config for GitHub / GitLab
RUN mkdir -p ~/.ssh \
    && { \
    echo "Host github.com gitlab.com"; \
    echo "    IdentityFile ~/.ssh/id_ed25519"; \
    echo "    StrictHostKeyChecking accept-new"; \
    echo "    AddKeysToAgent yes"; \
    } > ~/.ssh/config \
    && chmod 600 ~/.ssh/config

# Initialise GPG home directory so GPG creates its internal structure
RUN gpg --list-keys >/dev/null 2>&1 || true

COPY --chown=coder:coder entrypoint.sh /home/coder/entrypoint.sh

WORKDIR /workspaces
EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/global/health || exit 1

ENTRYPOINT ["bash", "/home/coder/entrypoint.sh"]
CMD ["opencode", "web", "--hostname", "0.0.0.0", "--port", "8080", "--cors", "*"]
