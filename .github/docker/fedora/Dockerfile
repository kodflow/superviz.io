# Fedora SSH Test Container
FROM fedora:39

# Update packages and install required packages
RUN dnf update -y && \
    dnf install -y \
        openssh-server \
        sudo \
        bash \
        curl \
        wget \
        git \
        vim \
        passwd && \
    dnf clean all

# Create SSH host keys
RUN ssh-keygen -A

# Configure SSH daemon
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config

# Set root password
RUN echo 'root:superviz123' | chpasswd

# Create test user with sudo privileges
RUN useradd -m -s /bin/bash testuser && \
    echo 'testuser:superviz123' | chpasswd && \
    echo 'testuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Create directory for SSH daemon
RUN mkdir -p /run/sshd

# Expose SSH port
EXPOSE 22

# Default command
CMD ["/usr/sbin/sshd", "-D"]
