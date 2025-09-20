#!/bin/bash
set -e

# Log all output to a file for debugging
exec > >(tee /var/log/user-data.log) 2>&1

echo "Starting user data script execution at $(date)"

# Update system packages
apt-get update

# Install required packages
apt-get install -y \
  ca-certificates \
  curl \
  gnupg \
  lsb-release \
  git \
  postgresql-client

# Add Docker's official GPG key
mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up the Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Create docker group and add default user
usermod -aG docker root

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.3/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create komodo directory structure
mkdir -p /etc/komodo/stacks

# Install Python and required packages
apt-get install -y python3 python3-pip
pip3 install requests toml pyyaml

echo "Starting komodo periphery at $(date)"
# Run setup script for komodo periphery
curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3

# Configure periphery with allowed IPs and passkeys
sed -i 's/allowed_ips = \[\]/allowed_ips = ["198.12.108.91"]/' /etc/komodo/periphery.config.toml
sed -i 's/passkeys = \[\]/passkeys = ["HNtSkk+UPrhsan8MaNbSGFkQ09HiftmSOWdjn5n0p4k="]/' /etc/komodo/periphery.config.toml

# Restart periphery service to apply configuration changes
systemctl restart periphery

# Set up firewall (ufw)
ufw --force enable
ufw allow OpenSSH
ufw allow 80/tcp # HTTP
ufw allow 443/tcp # HTTPS
ufw allow 8120/tcp # Komodo core
ufw allow 9120/tcp # Komodo periphery
ufw allow 51820/udp # WireGuard if applicable
ufw default deny incoming
ufw default allow outgoing
ufw status verbose

# Enable periphery service
systemctl enable periphery



echo "User data script finished at $(date)"