# Ubuntu test container with SSH server
FROM ubuntu:22.04

# Install SSH server and basic tools
RUN apt-get update && \
    apt-get install -y \
        openssh-server \
        sudo \
        curl \
        wget \
        gnupg \
        lsb-release \
        apt-transport-https \
        ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create SSH directory and configure SSH
RUN mkdir /var/run/sshd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    echo 'UsePAM yes' >> /etc/ssh/sshd_config

# Create test user
RUN useradd -m -s /bin/bash testuser && \
    echo 'testuser:testpass' | chpasswd && \
    echo 'root:rootpass' | chpasswd && \
    usermod -aG sudo testuser

# Add testuser to sudoers with NOPASSWD for automation
RUN echo 'testuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Expose SSH port
EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
