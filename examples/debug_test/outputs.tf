output "client_id" {
  description = "Client ID"
  value       = komodo-provider_user.debug_test.id
}

output "client_name" {
  description = "Client name"
  value       = komodo-provider_user.debug_test.name
}

output "server_ip" {
  description = "Server IP address"
  value       = komodo-provider_user.debug_test.server_ip
}

# SSH Key Outputs (only available when generate_ssh_keys = true)
output "ssh_private_key" {
  description = "Generated SSH private key for repository access"
  value       = komodo-provider_user.debug_test.ssh_private_key
  sensitive   = true
}

output "ssh_public_key" {
  description = "Generated SSH public key (also uploaded as deploy key)"
  value       = komodo-provider_user.debug_test.ssh_public_key
}

output "generate_ssh_keys" {
  description = "Whether SSH keys were generated"
  value       = komodo-provider_user.debug_test.generate_ssh_keys
}
