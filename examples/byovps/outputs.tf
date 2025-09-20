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