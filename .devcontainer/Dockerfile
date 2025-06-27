# syntax=docker/dockerfile:1.7
# Platform targeting ARM64 for Mac
FROM --platform=linux/arm64 mcr.microsoft.com/devcontainers/go:1.24-bullseye

# Single ENV layer for all environment variables
ENV GOVERSION=go1.24 \
    BUILT_BY=devcontainer \
    GOPATH=/home/vscode/go \
    GOMODCACHE=/home/vscode/go/pkg/mod \
    ZSH="/home/vscode/.oh-my-zsh" \
    SHELL="/usr/bin/zsh" \
    DEBIAN_FRONTEND=noninteractive

# Single RUN layer for all system setup to minimize layers
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/home/vscode/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/home/vscode/.cache/go-build,sharing=locked \
    --mount=type=cache,target=/root/.npm,sharing=locked \
    set -eux && \
    # Update package lists
    apt-get update && \
    # Install core system dependencies in one go
    apt-get install -y --no-install-recommends \
        make curl git unzip jq ca-certificates gnupg zsh \
        software-properties-common lsb-release && \
    # Install Node.js LTS
    curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs && \
    # Install Terraform
    curl -fsSL https://apt.releases.hashicorp.com/gpg \
        | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
        > /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && \
    apt-get install -y terraform && \
    # Install Go tools with cache
    go install gotest.tools/gotestsum@v1.12.2 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 && \
    go install github.com/goreleaser/goreleaser/v2@v2.10.2 && \
    # Install npm packages
    npm install -g npm prettier && \
    # Setup Go directories  
    mkdir -p /home/vscode/go/pkg/mod && \
    # Skip Oh My Zsh installation if already exists, install otherwise
    if [ ! -d "/home/vscode/.oh-my-zsh" ]; then \
        curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -o /tmp/install-oh-my-zsh.sh && \
        chmod +x /tmp/install-oh-my-zsh.sh && \
        env ZSH="/home/vscode/.oh-my-zsh" RUNZSH=no CHSH=no KEEP_ZSHRC=yes bash /tmp/install-oh-my-zsh.sh && \
        rm -f /tmp/install-oh-my-zsh.sh; \
    fi && \
    if [ ! -d "/home/vscode/.oh-my-zsh/custom/themes/powerlevel10k" ]; then \
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git \
            /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k; \
    fi && \
    chsh -s /usr/bin/zsh vscode && \
    # Set ownership (ignore errors if already owned)
    chown -R vscode:vscode /home/vscode/.oh-my-zsh /home/vscode/go 2>/dev/null || true && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy config files in a single layer
COPY --chown=vscode:vscode ./zshrc /home/vscode/.zshrc
COPY --chown=vscode:vscode ./.p10k.zsh /home/vscode/.p10k.zsh

# Set shell and user
SHELL ["/usr/bin/zsh", "-c"]
USER vscode
