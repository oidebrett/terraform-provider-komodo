# AWS Provider Configuration
variable "aws_region" {
  description = "The AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_access_key" {
  description = "AWS access key ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret access key"
  type        = string
  default     = ""
  sensitive   = true
}

# Instance Configuration
variable "instance_name" {
  description = "Name tag for the AWS EC2 instance"
  type        = string
  default     = "aws-client-instance"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "ami_id" {
  description = "AMI ID for the instance (Ubuntu 20.04 LTS)"
  type        = string
  default     = "ami-0c02fb55956c7d316"  # Ubuntu 20.04 LTS in us-east-1
}

# SSH Configuration
variable "key_pair_name" {
  description = "Name of the AWS key pair"
  type        = string
  default     = "aws-client-key"
}

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
  default     = ""
  sensitive   = true
}

variable "ssh_username" {
  description = "Username for SSH access"
  type        = string
  default     = "ubuntu"
}

# Security Group Configuration
variable "security_group_name" {
  description = "Name of the security group"
  type        = string
  default     = "client-instance-sg"
}

variable "allowed_ports" {
  description = "List of ports to allow through security group"
  type        = list(string)
  default     = ["22", "8120"]
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed for inbound traffic"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# Custom Provider Configuration
variable "komodo_provider_endpoint" {
  description = "Endpoint for custom user provider"
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