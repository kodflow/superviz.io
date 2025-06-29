# Debian Linux with SSH server
FROM debian:bookworm-slim

# Install OpenSSH server and required packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        openssh-server \
        openssh-client \
        sudo \
        ca-certificates \
        tzdata && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/*

# Create SSH directories and generate host keys
RUN mkdir -p /var/run/sshd && \
    mkdir -p /root/.ssh && \
    ssh-keygen -A

# Configure SSH daemon
RUN sed -i 's/#PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#Port 22/Port 22/' /etc/ssh/sshd_config

# Set root password and create test user
RUN echo 'root:superviz123' | chpasswd && \
    useradd -m -s /bin/bash testuser && \
    echo 'testuser:superviz123' | chpasswd && \
    usermod -aG sudo testuser

# Configure sudo for sudo group
RUN echo '%sudo ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

# Generate SSH key for root (for testing)
RUN ssh-keygen -t rsa -b 2048 -f /root/.ssh/id_rsa -N "" && \
    cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys && \
    chmod 700 /root/.ssh && \
    chmod 600 /root/.ssh/* && \
    chmod 755 /var/run/sshd

# Create non-root user with proper home directory
RUN mkdir -p /home/testuser/.ssh && \
    chown testuser:testuser /home/testuser/.ssh && \
    chmod 700 /home/testuser/.ssh

# Expose SSH port
EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D", "-e"]
