# GCP Provider Configuration
variable "gcp_project_id" {
  description = "The GCP project ID"
  type        = string
  default     = "your-project-id"
}

variable "gcp_region" {
  description = "The GCP region"
  type        = string
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "The GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "gcp_credentials_file" {
  description = "Path to the GCP service account key file"
  type        = string
  default     = ""
  sensitive   = true
}

# Instance Configuration
variable "instance_name" {
  description = "Name of the GCP compute instance"
  type        = string
  default     = "gcp-client-instance"
}

variable "machine_type" {
  description = "Machine type for the GCP instance"
  type        = string
  default     = "e2-medium"
}

variable "instance_image" {
  description = "Boot disk image for the instance"
  type        = string
  default     = "ubuntu-2004-focal-v20240307b"
}

variable "instance_tags" {
  description = "Network tags for the instance"
  type        = list(string)
  default     = ["gcp-client"]
}

# SSH Configuration
variable "ssh_public_key" {
  description = "SSH public key for instance access"
  type        = string
  default     = ""
  sensitive   = true
}

variable "ssh_username" {
  description = "Username for SSH access"
  type        = string
  default     = "ubuntu"
}

# Firewall Configuration
variable "firewall_name" {
  description = "Name of the firewall rule"
  type        = string
  default     = "allow-ssh-and-8120"
}

variable "allowed_ports" {
  description = "List of ports to allow through firewall"
  type        = list(string)
  default     = ["22", "8120"]
}

variable "firewall_source_ranges" {
  description = "Source IP ranges for firewall rule"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# Custom Provider Configuration
variable "komodo_provider_endpoint" {
  description = "Endpoint for custom komodo provider"
  type        = string
  default     = "http://127.0.0.1:9120"
}

variable "komodo_api_key" {
  description = "API key for komodo provider"
  type        = string
  default     = ""
}

variable "komodo_api_secret" {
  description = "API secret for komodo provider"
  type        = string
  sensitive   = true
  default     = ""
}

variable "github_token" {
  description = "GitHub token for custom provider"
  type        = string
  default     = ""
  sensitive   = true
}

# Client Configuration
variable "client_id" {
  description = "Client ID for the deployment"
  type        = string
  default     = "1"
}

variable "client_name" {
  description = "Client name for the deployment"
  type        = string
  default     = "Client1"
}

variable "domain" {
  description = "Domain for the application"
  type        = string
  default     = "example.com"
}

variable "admin_email" {
  description = "Admin email address"
  type        = string
  default     = "admin@example.com"
}

variable "admin_username" {
  description = "Admin username"
  type        = string
  default     = "admin@example.com"
}

variable "admin_password" {
  description = "Admin password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "admin_subdomain" {
  description = "Admin subdomain"
  type        = string
  default     = "admin"
}

variable "crowdsec_enrollment_key" {
  description = "CrowdSec enrollment key"
  type        = string
  default     = ""
  sensitive   = true
}

variable "postgres_user" {
  description = "PostgreSQL username"
  type        = string
  default     = "admin"
}

variable "postgres_password" {
  description = "PostgreSQL password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "postgres_host" {
  description = "PostgreSQL host"
  type        = string
  default     = "postgres"
}

variable "oauth_client_id" {
  description = "OAuth client ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "oauth_client_secret" {
  description = "OAuth client secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "github_repo" {
  description = "GitHub repository for deployment"
  type        = string
  default     = "oidebrett/getcontextware"
}

variable "static_page_domain" {
  description = "Static page domain"
  type        = string
  default     = "www"
}