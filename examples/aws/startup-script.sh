#!/bin/bash -xe

# Log all output to a file for debugging
exec > >(tee /var/log/user-data.log) 2>&1

echo "Starting user data script execution at $(date)"

# Try to wait for cloud-init with a timeout
timeout 60 cloud-init status --wait || echo "Cloud-init wait timed out after 60 seconds, continuing anyway"

# Sleep a bit to give system time to settle
sleep 30

echo "Proceeding with Docker installation at $(date)"

# Update package lists
apt-get update

# Install prerequisites
apt-get install -y \
  ca-certificates \
  curl \
  gnupg \
  lsb-release \
  git \
  postgresql-client

# Set up Docker repository
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Update package lists again with Docker repository
apt-get update

# Install Docker
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add the default user to the docker group
usermod -aG docker ubuntu

# Enable Docker service
systemctl enable docker.service
systemctl enable containerd.service
systemctl start docker.service

echo "Docker installation completed successfully at $(date)!"

echo "Starting komodo periphery at $(date)"

# Run setup script for komodo periphery
curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3

# Enable periphery service
systemctl enable periphery

echo "User data script finished at $(date)"