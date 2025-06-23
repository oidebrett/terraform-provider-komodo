# Azure Provider Configuration
variable "azure_subscription_id" {
  description = "The Azure subscription ID"
  type        = string
  default     = ""
}

variable "azure_client_id" {
  description = "The Azure client ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "azure_client_secret" {
  description = "The Azure client secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "azure_tenant_id" {
  description = "The Azure tenant ID"
  type        = string
  default     = ""
}

variable "azure_location" {
  description = "The Azure region"
  type        = string
  default     = "East US"
}

# Instance Configuration
variable "instance_name" {
  description = "Name of the Azure virtual machine"
  type        = string
  default     = "azure-client-instance"
}

variable "vm_size" {
  description = "Size of the Azure virtual machine"
  type        = string
  default     = "Standard_B2s"
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
  default     = "adminuser"
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

# Database Configuration
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

# Application Configuration
variable "static_page_domain" {
  description = "Static page domain"
  type        = string
  default     = "www"
}

# OAuth Configuration
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

# Custom Provider Configuration
variable "myuserprovider_endpoint" {
  description = "Custom provider endpoint"
  type        = string
  default     = "http://localhost:9120"
}

variable "github_token" {
  description = "GitHub token for repository access"
  type        = string
  default     = ""
  sensitive   = true
}

# Repository Configuration
variable "github_repo" {
  description = "GitHub repository for application code"
  type        = string
  default     = "username/repo"
}