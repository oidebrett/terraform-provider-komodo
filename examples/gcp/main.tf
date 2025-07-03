# Terraform configuration
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    komodo-provider = {
      source  = "example.com/me/komodo-provider"
    }
  }
}

# Google Cloud Provider
provider "google" {
  credentials = var.gcp_credentials_file != "" ? file(var.gcp_credentials_file) : null
  project     = var.gcp_project_id
  region      = var.gcp_region
  zone        = var.gcp_zone
}

# Custom User Provider
provider "komodo-provider" {
  endpoint     = var.komodo_provider_endpoint
  api_key      = var.komodo_api_key
  api_secret   = var.komodo_api_secret
  github_token = var.github_token
}

# Firewall rule to allow SSH and application port
resource "google_compute_firewall" "allow_ports" {
  name    = var.firewall_name
  network = "default"

  allow {
    protocol = "tcp"
    ports    = var.allowed_ports
  }

  source_ranges = var.firewall_source_ranges
  target_tags   = var.instance_tags
}

# GCP Compute Instance
resource "google_compute_instance" "gcp_vm" {
  name         = var.instance_name
  machine_type = var.machine_type
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      image = var.instance_image
    }
  }

  network_interface {
    network       = "default"
    access_config {} # Needed for a public IP
  }

  metadata = {
    ssh-keys = var.ssh_public_key != "" ? "${var.ssh_username}:${var.ssh_public_key}" : null
  }

  metadata_startup_script = templatefile("${path.module}/startup-script.sh", {})

  tags = var.instance_tags

  lifecycle {
    ignore_changes = [
      metadata["ssh-keys"]
    ]
  }
}

# Custom provider resource with templated configuration
resource "komodo-provider_user" "client_syncresources" {
  id                = var.client_id
  name              = var.client_name
  file_contents = templatefile("${path.module}/config-template.toml", {
    client_name_lower        = lower(var.client_name)
    client_name             = var.client_name
    domain                  = var.domain
    admin_email             = var.admin_email
    admin_username          = var.admin_username
    admin_password          = var.admin_password
    admin_subdomain         = var.admin_subdomain
    crowdsec_enrollment_key = var.crowdsec_enrollment_key
    postgres_user           = var.postgres_user
    postgres_password       = var.postgres_password
    postgres_host           = var.postgres_host
    static_page_domain      = lower(tostring(var.static_page_domain))
    oauth_client_id         = var.oauth_client_id
    oauth_client_secret     = var.oauth_client_secret
    github_repo             = var.github_repo
    komodo_host_ip          = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
  })
  server_ip = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}