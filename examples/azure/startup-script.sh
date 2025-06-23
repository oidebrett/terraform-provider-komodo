#!/bin/bash

# Log startup
echo "Starting startup script at $(date)" > /var/log/startup.log

# Update system
apt-get update
apt-get upgrade -y

# Install Docker and Docker Compose
apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io
systemctl enable docker
systemctl start docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.3/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create komodo directory structure
mkdir -p /etc/komodo/stacks

# Install Python and required packages
apt-get install -y python3 python3-pip
pip3 install requests toml pyyaml

echo "Starting komodo periphery at $(date)" >> /var/log/startup.log
# Run your setup script
curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3
systemctl enable periphery

echo "User data script finished at $(date)" >> /var/log/startup.log