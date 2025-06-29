# syntax=docker/dockerfile:1.7
FROM mcr.microsoft.com/devcontainers/go:1.24-bookworm

# Single ENV layer for all environment variables
ENV GOVERSION=go1.24 \
    BUILT_BY=devcontainer \
    GOPATH=/home/vscode/go \
    GOMODCACHE=/home/vscode/go/pkg/mod \
    ZSH="/home/vscode/.oh-my-zsh" \
    SHELL="/usr/bin/zsh" \
    DEBIAN_FRONTEND=noninteractive \
    PATH="/home/vscode/go/bin:/usr/local/bin:$PATH"

# Single RUN layer for all system setup to minimize layers
RUN set -eux && \
    # Update package lists
    apt-get update && \
    # Install core system dependencies in one go
    apt-get install -y --no-install-recommends \
        make curl git unzip jq ca-certificates gnupg zsh \
        software-properties-common lsb-release sudo \
        sshpass netcat-openbsd telnet iputils-ping nmap \
        openssh-client dnsutils net-tools && \
    # Get the codename for Docker repository
    CODENAME=$(lsb_release -cs) && \
    # Install Docker CLI and Docker Compose
    curl -fsSL https://download.docker.com/linux/debian/gpg \
        | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=arm64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian ${CODENAME} stable" \
        > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y docker-ce-cli docker-compose-plugin && \
    # Install Node.js LTS
    curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs && \
    # Get the codename for Terraform repository
    CODENAME=$(lsb_release -cs) && \
    # Install Terraform
    curl -fsSL https://apt.releases.hashicorp.com/gpg \
        | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com ${CODENAME} main" \
        > /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && \
    apt-get install -y terraform && \
    # Install Docker-in-Docker (DinD)
    curl -fsSL https://get.docker.com -o get-docker.sh && \
    sh get-docker.sh && \
    rm get-docker.sh && \
    # Setup Docker daemon configuration for DinD
    mkdir -p /etc/docker && \
    echo '{"hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"], "tls": false}' > /etc/docker/daemon.json && \
    # Install Bazel (using binary release for ARM64)
    curl -fsSL "https://github.com/bazelbuild/bazel/releases/download/7.4.1/bazel-7.4.1-linux-arm64" \
        -o /usr/local/bin/bazel && \
    chmod +x /usr/local/bin/bazel && \
    # Setup Go directories first
    mkdir -p /home/vscode/go/bin /home/vscode/go/pkg/mod /home/vscode/go/src && \
    chown -R vscode:vscode /home/vscode/go && \
    # Install npm packages
    npm install -g npm prettier && \
    # Install Go tools as vscode user
    su vscode -c 'go install gotest.tools/gotestsum@v1.12.2' && \
    su vscode -c 'go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6' && \
    su vscode -c 'go install github.com/goreleaser/goreleaser/v2@v2.10.2' && \
    # Install Bazel tools as vscode user
    su vscode -c 'go install github.com/bazelbuild/buildtools/buildifier@latest' && \
    su vscode -c 'go install github.com/bazelbuild/buildtools/buildozer@latest' && \
    # Install Oh My Zsh for vscode user
    if [ ! -d "/home/vscode/.oh-my-zsh" ]; then \
        curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -o /tmp/install-oh-my-zsh.sh && \
        chmod +x /tmp/install-oh-my-zsh.sh && \
        su vscode -c 'env ZSH="/home/vscode/.oh-my-zsh" RUNZSH=no CHSH=no KEEP_ZSHRC=yes bash /tmp/install-oh-my-zsh.sh' && \
        rm -f /tmp/install-oh-my-zsh.sh; \
    fi && \
    # Install Oh My Zsh for root user
    if [ ! -d "/root/.oh-my-zsh" ]; then \
        curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -o /tmp/install-oh-my-zsh-root.sh && \
        chmod +x /tmp/install-oh-my-zsh-root.sh && \
        env ZSH="/root/.oh-my-zsh" RUNZSH=no CHSH=no KEEP_ZSHRC=yes bash /tmp/install-oh-my-zsh-root.sh && \
        rm -f /tmp/install-oh-my-zsh-root.sh; \
    fi && \
    # Install powerlevel10k for vscode user
    if [ ! -d "/home/vscode/.oh-my-zsh/custom/themes/powerlevel10k" ]; then \
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git \
            /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k && \
        chown -R vscode:vscode /home/vscode/.oh-my-zsh; \
    fi && \
    # Install powerlevel10k for root user
    if [ ! -d "/root/.oh-my-zsh/custom/themes/powerlevel10k" ]; then \
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git \
            /root/.oh-my-zsh/custom/themes/powerlevel10k; \
    fi && \
    chsh -s /usr/bin/zsh vscode && \
    chsh -s /usr/bin/zsh root && \
    # Add vscode user to docker group (will be created by docker feature)
    usermod -aG docker vscode || true && \
    # Final ownership fix
    chown -R vscode:vscode /home/vscode && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy config files in a single layer
COPY --chown=vscode:vscode ./zshrc /home/vscode/.zshrc
COPY --chown=vscode:vscode ./.p10k.zsh /home/vscode/.p10k.zsh

# Copy config files for root user
COPY --chown=root:root ./zshrc /root/.zshrc
COPY --chown=root:root ./.p10k.zsh /root/.p10k.zsh

# Set the default user
USER vscode
WORKDIR /workspace
