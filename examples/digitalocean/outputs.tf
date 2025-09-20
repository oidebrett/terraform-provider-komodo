output "instance_ip" {
  description = "Public IPv4 address of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.ipv4_address
}

output "instance_name" {
  description = "Name of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.name
}

output "instance_id" {
  description = "ID of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.id
}

output "instance_status" {
  description = "Status of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.status
}

output "instance_region" {
  description = "Region of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.region
}

output "instance_plan" {
  description = "Size (plan) of the DigitalOcean Droplet"
  value       = digitalocean_droplet.main.size
}

output "client_id" {
  description = "Client ID"
  value       = var.client_id
}

output "client_name" {
  description = "Client name"
  value       = var.client_name
}

output "domain" {
  description = "Application domain"
  value       = var.domain
}

output "firewall_id" {
  description = "ID of the DigitalOcean Firewall"
  value       = var.create_firewall ? digitalocean_firewall.web[0].id : null
}

output "ssh_key_id" {
  description = "ID of the SSH key"
  value       = digitalocean_ssh_key.default.id
}

# SSH Key Outputs (only available when generate_ssh_keys = true)
output "ssh_private_key" {
  description = "Generated SSH private key for repository access"
  value       = komodo-provider_user.client_syncresources.ssh_private_key
  sensitive   = true
}

output "ssh_public_key" {
  description = "Generated SSH public key (also uploaded as deploy key)"
  value       = komodo-provider_user.client_syncresources.ssh_public_key
}