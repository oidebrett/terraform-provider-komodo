variable "client_name" {
  description = "Client's name"
  type        = string
  default     = "DebugTest"
}

variable "client_id" {
  description = "Unique identifier for the client"
  type        = string
  default     = "debug-test-001"
}

variable "server_ip" {
  description = "Server IP address"
  type        = string
  default     = "1.2.3.4"
}

variable "generate_ssh_keys" {
  description = "Whether to generate SSH keys"
  type        = bool
  default     = true
}

variable "domain" {
  description = "Application domain"
  type        = string
  default     = "debug.example.com"
}

variable "admin_email" {
  description = "Admin user email"
  type        = string
  default     = "admin@debug.example.com"
}

variable "admin_username" {
  description = "Admin username"
  type        = string
  default     = "admin@debug.example.com"
}

variable "admin_password" {
  description = "Admin password"
  type        = string
  default     = "DebugPassword123!"
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
  default     = "debug/test-repo"
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
