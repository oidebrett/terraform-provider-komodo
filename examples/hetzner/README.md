# Hetzner Cloud Example - Terraform

## Overview

This example demonstrates how to use Terraform to provision a simple server on Hetzner Cloud with Docker and PostgreSQL client installed.

## Prerequisites

1. **Hetzner Cloud Account**
   - Sign up at [https://console.hetzner.cloud](https://console.hetzner.cloud)
   - Create a project
   - Generate an API token with read/write permissions

2. **Terraform Installed**
   - Version 0.13 or newer

3. **SSH Key**
   - Generate an SSH key if you don't have one:
     ```bash
     ssh-keygen -f ~/.ssh/hetzner_key -t rsa -b 4096 -N ''
     ```

## Getting Started

1. **Copy the example variables file**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit the variables file with your values**
   - Add your Hetzner Cloud API token
   - Configure the server type and location
   - Add your SSH public key
   - Optionally restrict firewall source IPs

3. **Initialize Terraform**
   ```bash
   terraform init
   ```

4. **Plan the deployment**
   ```bash
   terraform plan
   ```

5. **Apply the configuration**
   ```bash
   terraform apply
   ```

6. **Access your server**
   After deployment completes, you can access the server via SSH:
   ```bash
   ssh -i ~/.ssh/hetzner_key admin@<server_ip>
   ```

7. **Clean up when finished**
   ```bash
   terraform destroy
   ```

## Firewall Configuration

The server is configured with a firewall that allows inbound traffic on the following ports:
- Port 22: SSH access
- Port 80: HTTP traffic
- Port 443: HTTPS traffic
- Port 8120: Custom application port
- Port 9120: Custom API port

For better security, consider restricting the source IP ranges in your `terraform.tfvars` file.

## Customization

Modify the `terraform.tfvars` file to adjust:
- Server type and location
- Operating system image
- Firewall source IP restrictions
