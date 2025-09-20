variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "DigitalOcean region"
  type        = string
  default     = "nyc3" # Example default for DigitalOcean
}

variable "instance_type" {
  description = "DigitalOcean Droplet size (e.g., s-1vcpu-1gb)"
  type        = string
  default     = "s-1vcpu-1gb" # Example default for DigitalOcean
}

variable "do_image_slug" {
  description = "DigitalOcean Droplet image slug (e.g., ubuntu-22-04-x64)"
  type        = string
  default     = "ubuntu-22-04-x64" # Example default
}

variable "ssh_public_key" {
  description = "SSH public key for server access"
  type        = string
}

variable "instance_name" {
  description = "Name of the DigitalOcean Droplet"
  type        = string
}

variable "client_name" {
  description = "Client's name"
  type        = string
}

variable "client_name_lower" {
  description = "Lowercase version of the client's name"
  type        = string
}

variable "client_id" {
  description = "Unique identifier for the client"
  type        = string
}

variable "create_firewall" {
  description = "Whether to create firewall rules"
  type        = bool
  default     = true
}

variable "allowed_source_ips" {
  description = "List of allowed source IP ranges for SSH"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "domain" {
  description = "Application domain"
  type        = string
}

variable "admin_email" {
  description = "Admin user email"
  type        = string
}

variable "admin_username" {
  description = "Admin username"
  type        = string
}

variable "admin_password" {
  description = "Admin password"
  type        = string
  sensitive   = true
}

variable "admin_subdomain" {
  description = "Admin subdomain"
  type        = string
  default     = "admin"
}



variable "github_repo" {
  description = "GitHub repository for client resources"
  type        = string
}

variable "komodo_provider_endpoint" {
  description = "Komodo provider API endpoint"
  type        = string
}

variable "komodo_api_key" {
  description = "Komodo API key"
  type        = string
  sensitive   = true
}

variable "komodo_api_secret" {
  description = "Komodo API secret"
  type        = string
  sensitive   = true
}

variable "github_token" {
  description = "GitHub token for authentication"
  type        = string
  sensitive   = true
}



