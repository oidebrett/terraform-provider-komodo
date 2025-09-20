terraform {
  required_providers {
    komodo-provider = {
      source = "registry.example.com/mattercoder/komodo-provider"
      version = ">= 1.0.0"
    }
  }
}

provider "komodo-provider" {
  endpoint     = var.komodo_provider_endpoint
  api_key      = var.komodo_api_key
  api_secret   = var.komodo_api_secret
  github_token = var.github_token
}

resource "komodo-provider_user" "debug_test" {
  id               = var.client_id
  name             = var.client_name
  generate_ssh_keys = var.generate_ssh_keys
  server_ip        = var.server_ip
  file_contents = templatefile("${path.module}/config-template.toml", {
    client_name_lower = lower(var.client_name)
    client_name       = var.client_name
    domain           = var.domain
    admin_email      = var.admin_email
    admin_username   = var.admin_username
    admin_password   = var.admin_password
    admin_subdomain  = var.admin_subdomain
    github_repo      = var.github_repo
  })
}
