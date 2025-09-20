terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0" # Use an appropriate version
    }
    komodo-provider = {
      source = "registry.example.com/mattercoder/komodo-provider"
      version = ">= 1.0.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_ssh_key" "default" {
  name       = "terraform-key-${var.client_name_lower}"
  public_key = var.ssh_public_key
}

resource "digitalocean_droplet" "main" {
  image      = var.do_image_slug
  name       = var.instance_name
  region     = var.region
  size       = var.instance_type
  ssh_keys   = [digitalocean_ssh_key.default.id]
  user_data  = file("${path.module}/startup-script.sh")
  tags       = [var.client_name_lower]
}

resource "digitalocean_firewall" "web" {
  count = var.create_firewall ? 1 : 0
  name  = "firewall-${var.client_name_lower}"
  
  droplet_ids = [digitalocean_droplet.main.id]

  inbound_rule {
    protocol   = "tcp"
    port_range = "22"
    source_addresses = var.allowed_source_ips
  }
  inbound_rule {
    protocol   = "tcp"
    port_range = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol   = "tcp"
    port_range = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol   = "tcp"
    port_range = "8120"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol   = "tcp"
    port_range = "9120"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule { # Add WireGuard port if used
    protocol   = "udp"
    port_range = "51820"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

provider "komodo-provider" {
  endpoint     = var.komodo_provider_endpoint
  api_key      = var.komodo_api_key
  api_secret   = var.komodo_api_secret
  github_token = var.github_token
}

resource "komodo-provider_user" "client_syncresources" {
  depends_on = [digitalocean_droplet.main]

  id            = var.client_id
  name          = var.client_name
  generate_ssh_keys = true
  file_contents = templatefile("${path.module}/config-template.toml", {
    client_name_lower       = lower(var.client_name)
    client_name             = var.client_name
    domain                  = var.domain
    admin_email             = var.admin_email
    admin_username          = var.admin_username
    admin_password          = var.admin_password
    admin_subdomain         = var.admin_subdomain
    github_repo             = var.github_repo
    
    
  })
  server_ip     = digitalocean_droplet.main.ipv4_address
}