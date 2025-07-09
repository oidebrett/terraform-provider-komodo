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
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

# Add Docker repository
echo \
"deb [arch=\"$(dpkg --print-architecture)\" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
\"$(. /etc/os-release && echo \"$VERSION_CODENAME\")\" stable" | \
tee /etc/apt/sources.list.d/docker.list > /dev/null

# Update package index with Docker repository
apt-get update -y

# Install Docker
apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

# Add ubuntu user to docker group
# usermod -aG docker ubuntu # not needed as hetzer runs script as root

# Activate the new group membership
newgrp docker

echo "Starting komodo periphery at $(date)"

# Run setup script for komodo periphery
curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3

# Enable periphery service
systemctl enable periphery

echo "User data script finished at $(date)"