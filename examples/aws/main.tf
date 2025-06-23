# Terraform configuration
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    komodo-provider = {
      source = "example.com/me/komodo-provider"
      # version = "~> 1.0"
    }
  }
}

# AWS Provider
provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key != "" ? var.aws_access_key : null
  secret_key = var.aws_secret_key != "" ? var.aws_secret_key : null
}

# Custom Komodo Provider
provider "komodo-provider" {
  endpoint     = var.komodo_provider_endpoint
  github_token = var.github_token
}

# Data source to get the default VPC
data "aws_vpc" "default" {
  default = true
}

# Create AWS Key Pair
resource "aws_key_pair" "client_key" {
  count      = var.ssh_public_key != "" ? 1 : 0
  key_name   = var.key_pair_name
  public_key = var.ssh_public_key

  tags = {
    Name = var.key_pair_name
  }
}

# Security Group for EC2 instance
resource "aws_security_group" "client_sg" {
  name        = var.security_group_name
  description = "Security group for client instance"
  vpc_id      = data.aws_vpc.default.id

  # Inbound rules for allowed ports
  dynamic "ingress" {
    for_each = var.allowed_ports
    content {
      description = "Allow port ${ingress.value}"
      from_port   = tonumber(ingress.value)
      to_port     = tonumber(ingress.value)
      protocol    = "tcp"
      cidr_blocks = var.allowed_cidr_blocks
    }
  }

  # Outbound rule - allow all traffic
  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = var.security_group_name
  }
}

# EC2 Instance
resource "aws_instance" "client_instance" {
  ami                         = var.ami_id
  instance_type              = var.instance_type
  key_name                   = var.ssh_public_key != "" ? aws_key_pair.client_key[0].key_name : null
  vpc_security_group_ids     = [aws_security_group.client_sg.id]
  associate_public_ip_address = true

  user_data = templatefile("${path.module}/startup-script.sh", {})

  tags = {
    Name = var.instance_name
  }

  # Ensure the instance is properly initialized before proceeding
  lifecycle {
    create_before_destroy = true
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
  })
  server_ip = aws_instance.client_instance.public_ip
}