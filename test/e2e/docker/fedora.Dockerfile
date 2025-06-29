# Fedora test container with SSH server
FROM fedora:39

# Install SSH server and basic tools
RUN dnf update -y && \
    dnf install -y \
        openssh-server \
        sudo \
        curl \
        wget \
        gnupg2 \
        redhat-lsb-core && \
    dnf clean all

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
