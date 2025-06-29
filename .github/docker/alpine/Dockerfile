# Alpine Linux with SSH server
FROM alpine:3.18

# Install OpenSSH server and required packages
RUN apk add --no-cache \
    openssh-server \
    openssh-client \
    sudo \
    shadow \
    ca-certificates \
    tzdata

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
    adduser -D -s /bin/sh testuser && \
    echo 'testuser:superviz123' | chpasswd && \
    adduser testuser wheel

# Configure sudo for wheel group
RUN echo '%wheel ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

# Generate SSH key for root (for testing)
RUN ssh-keygen -t rsa -b 2048 -f /root/.ssh/id_rsa -N "" && \
    cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys && \
    chmod 700 /root/.ssh && \
    chmod 600 /root/.ssh/* && \
    chmod 755 /var/run/sshd

# Expose SSH port
EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D", "-e"]
