# Alpine test container with SSH server
FROM alpine:3.18

# Install SSH server and basic tools
RUN apk add --no-cache \
        openssh \
        sudo \
        curl \
        wget \
        bash \
        shadow && \
    # Generate SSH host keys
    ssh-keygen -A && \
    # Configure SSH
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create test user
RUN adduser -D -s /bin/bash testuser && \
    echo 'testuser:testpass' | chpasswd && \
    echo 'root:rootpass' | chpasswd && \
    adduser testuser wheel

# Configure sudo for wheel group
RUN echo '%wheel ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Expose SSH port
EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D", "-e"]
