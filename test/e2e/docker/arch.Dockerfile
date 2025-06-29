# Arch Linux test container with SSH server
FROM archlinux:latest

# Install SSH server and basic tools
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
        openssh \
        sudo \
        curl \
        wget \
        gnupg \
        base-devel && \
    pacman -Scc --noconfirm

# Create SSH directory and configure SSH
RUN mkdir /var/run/sshd && \
    ssh-keygen -A && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create test user
RUN useradd -m -s /bin/bash testuser && \
    echo 'testuser:testpass' | chpasswd && \
    echo 'root:rootpass' | chpasswd && \
    usermod -aG wheel testuser

# Configure sudo for wheel group
RUN echo '%wheel ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Expose SSH port
EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D"]
