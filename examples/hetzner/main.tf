# Terraform configuration for Hetzner Cloud
terraform {
  required_providers {
    hcloud = {
      source = "hetznercloud/hcloud"
    }
  }
  required_version = ">= 0.13"
}

# Configure the Hetzner Cloud Provider
provider "hcloud" {
  token = var.hcloud_token
}

# SSH Key Resource
resource "hcloud_ssh_key" "default" {
  name       = "terraform-key"
  public_key = var.ssh_public_key
}

# Firewall Resource
resource "hcloud_firewall" "web_firewall" {
  name = "web-firewall"
  
  # SSH
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = var.allowed_source_ips
  }
  
  # HTTP
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = var.allowed_source_ips
  }
  
  # HTTPS
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = var.allowed_source_ips
  }
  
  # Custom port 8120
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "8120"
    source_ips = var.allowed_source_ips
  }
  
  # Custom port 9120
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "9120"
    source_ips = var.allowed_source_ips
  }
}

# Server Resource
resource "hcloud_server" "web" {
  name        = "my-server"
  image       = var.os_type
  server_type = var.server_type
  location    = var.location
  ssh_keys    = [hcloud_ssh_key.default.id]
  user_data   = file("${path.module}/user_data.yml")
  
  # Apply the firewall to this server
  firewall_ids = [hcloud_firewall.web_firewall.id]
}
