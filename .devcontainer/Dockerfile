FROM mcr.microsoft.com/devcontainers/go:latest

# 1. Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      make curl git unzip jq ca-certificates gnupg zsh \
      software-properties-common lsb-release && \
    curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs && \
    curl -fsSL https://apt.releases.hashicorp.com/gpg \
      | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
      > /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && \
    apt-get install -y terraform && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# 2. Install Go and JS tools
RUN go install gotest.tools/gotestsum@latest && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.56.2 && \
    curl -sSfL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz \
      | tar -xz -C /usr/local/bin goreleaser && \
    npm install -g npm@latest prettier

# 3. Setup Go module cache in user space to avoid permission issues
ENV GOPATH=/home/vscode/go
ENV GOMODCACHE=/home/vscode/go/pkg/mod
RUN mkdir -p /home/vscode/go/pkg/mod && \
    chown -R vscode:vscode /home/vscode/go

# 4. Install Oh My Zsh (no prompt)
RUN rm -rf /home/vscode/.oh-my-zsh && \
    curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -o /tmp/install-oh-my-zsh.sh && \
    chmod +x /tmp/install-oh-my-zsh.sh && \
    env ZSH="/home/vscode/.oh-my-zsh" RUNZSH=no CHSH=no KEEP_ZSHRC=yes bash /tmp/install-oh-my-zsh.sh && \
    chsh -s /usr/bin/zsh vscode && \
    chown -R vscode:vscode /home/vscode/.oh-my-zsh && \
    rm -f /tmp/install-oh-my-zsh.sh

# 5. Install Powerlevel10k theme
RUN rm -rf /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k && \
    git clone --depth=1 https://github.com/romkatv/powerlevel10k.git \
      /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k && \
    chown -R vscode:vscode /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k

# 6. Inject user config
COPY ./zshrc /home/vscode/.zshrc
COPY ./.p10k.zsh /home/vscode/.p10k.zsh
RUN chown vscode:vscode /home/vscode/.zshrc /home/vscode/.p10k.zsh

# 7. Set zsh as default shell
SHELL ["/usr/bin/zsh", "-c"]
ENV ZSH="/home/vscode/.oh-my-zsh"
ENV SHELL="/usr/bin/zsh"

USER vscode
