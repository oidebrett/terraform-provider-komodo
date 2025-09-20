# Komodo Configuration
variable "komodo_endpoint" {
  description = "Komodo API endpoint URL"
  type        = string
  default     = "http://localhost:9120"
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

# GitHub Configuration
variable "github_token" {
  description = "GitHub Personal Access Token"
  type        = string
  sensitive   = true
}

variable "github_orgname" {
  description = "GitHub organization name (optional)"
  type        = string
  default     = ""
}

# Client Configuration
variable "client_name" {
  description = "Name of the client"
  type        = string
  default     = "SSHExample"
}

variable "client_id" {
  description = "Client ID for the deployment"
  type        = string
  default     = "1"
}

variable "server_ip" {
  description = "Public IP address of the server"
  type        = string
}

variable "generate_ssh_keys" {
  description = "Whether to generate SSH keys and upload them as deploy keys to the GitHub repository"
  type        = bool
  default     = true
}

# Application Configuration
variable "domain" {
  description = "Domain name for the application"
  type        = string
  default     = "example.com"
}

variable "email" {
  description = "Email address for the application"
  type        = string
  default     = "admin@example.com"
}

variable "admin_username" {
  description = "Admin username for the application"
  type        = string
  default     = "admin@example.com"
}

variable "admin_password" {
  description = "Admin password for the application"
  type        = string
  default     = "SecurePassword123!"
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
  default     = "SecureDBPassword123!"
  sensitive   = true
}

variable "postgres_host" {
  description = "PostgreSQL host"
  type        = string
  default     = "komodo-postgres-1"
}

variable "static_page" {
  description = "Enable static page"
  type        = bool
  default     = true
}

variable "oauth_client_id" {
  description = "OAuth client ID"
  type        = string
  default     = ""
}

variable "oauth_client_secret" {
  description = "OAuth client secret"
  type        = string
  default     = ""
  sensitive   = true
}
