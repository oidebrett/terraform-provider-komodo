# Output values for the SSH Keys example

output "client_name" {
  description = "Name of the created client"
  value       = komodo-provider_user.ssh_example.name
}

output "client_id" {
  description = "ID of the created client"
  value       = komodo-provider_user.ssh_example.id
}

output "server_ip" {
  description = "IP address of the server"
  value       = komodo-provider_user.ssh_example.server_ip
}

output "github_repository" {
  description = "GitHub repository name (with _syncresources suffix)"
  value       = "${lower(replace(var.client_name, " ", "-"))}_syncresources"
}

output "github_repository_url" {
  description = "GitHub repository URL"
  value       = var.github_orgname != "" ? "https://github.com/${var.github_orgname}/${lower(replace(var.client_name, " ", "-"))}_syncresources" : "https://github.com/[your-username]/${lower(replace(var.client_name, " ", "-"))}_syncresources"
}

output "ssh_keys_info" {
  description = "Information about SSH keys"
  value = {
    generated = "SSH keys have been automatically generated and embedded in the repository file contents"
    location  = "Check the environment section in resources.toml for SSH_PRIVATE_KEY and SSH_PUBLIC_KEY"
    deploy_key = "Public key has been uploaded as a deploy key to the repository"
  }
}
