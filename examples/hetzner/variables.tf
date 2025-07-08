variable "hcloud_token" {
  description = "Hetzner Cloud API Token"
  type        = string
  sensitive   = true
}

variable "location" {
  description = "Hetzner datacenter location"
  type        = string
  default     = "nbg1"  # Nuremberg, Germany
}

variable "server_type" {
  description = "Hetzner server type/size"
  type        = string
  default     = "cx11"  # 1 vCPU, 2GB RAM
}

variable "os_type" {
  description = "Operating system image"
  type        = string
  default     = "ubuntu-20.04"
}

variable "ssh_public_key" {
  description = "SSH public key for server access"
  type        = string
}

variable "allowed_source_ips" {
  description = "List of allowed source IPs for firewall rules (CIDR notation)"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Allow from anywhere by default
}
