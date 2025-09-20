# Terraform configuration for SSH Keys example
terraform {
  required_providers {
    komodo-provider = {
      source = "registry.example.com/mattercoder/komodo-provider"
      version = ">= 1.0.0"
    }
  }
  required_version = ">= 0.13"
}

# Configure the Komodo provider
provider "komodo-provider" {
  endpoint       = var.komodo_endpoint
  api_key        = var.komodo_api_key
  api_secret     = var.komodo_api_secret
  github_token   = var.github_token
  github_orgname = var.github_orgname
}

# Create a user resource with SSH key generation
resource "komodo-provider_user" "ssh_example" {
  id               = var.client_id
  name             = var.client_name
  server_ip        = var.server_ip
  generate_ssh_keys = var.generate_ssh_keys
  file_contents = templatefile("${path.module}/config-template.toml", {
    client_name       = var.client_name
    client_name_lower = lower(replace(var.client_name, " ", "-"))
    domain           = var.domain
    email            = var.email
    admin_username   = var.admin_username
    admin_password   = var.admin_password
    admin_subdomain  = var.admin_subdomain
    crowdsec_key     = var.crowdsec_enrollment_key
    postgres_user    = var.postgres_user
    postgres_password = var.postgres_password
    postgres_host    = var.postgres_host
    static_page      = var.static_page
    client_id        = var.oauth_client_id
    client_secret    = var.oauth_client_secret
  })
}
