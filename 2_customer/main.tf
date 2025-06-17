# load the providers
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    myuserprovider = {
      source  = "example.com/me/myuserprovider"
      # version = "~> 1.0"
    }
  }
}



provider "google" {
  credentials = file("/workspace/2_customer/your-service-account-key.json")
  project     = "matter-test1-133b7"
  region      = "europe-west1"
  zone        = "europe-west1-b"
}

resource "google_compute_firewall" "allow-ssh-and-8120" {
  name    = "allow-ssh-and-8120"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22", "8120"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["gcp-client"]
}

resource "google_compute_instance" "gcp_vm" {
  name         = "gcp-client-instance"
  machine_type = "e2-micro"
  zone         = "europe-west1-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  network_interface {
    network       = "default"
    access_config {} # Needed for a public IP
  }

  metadata = {
    ssh-keys = "ubuntu:ssh-rsa YOURSSHKEYHERE ubuntu"
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3
    systemctl enable periphery
  EOF

  tags = ["gcp-client"]
}

provider "aws" {
  region = "eu-west-1"
  # Add your AWS credentials or use environment variables
}

#resource "aws_key_pair" "VPS-key" {
#  key_name   = "VPS-key"
#  public_key = "ssh-rsa YOURSSHKEYHERE"
#}

# Create AWS EC2 instance
#resource "aws_instance" "example" {
#  ami           = "ami-0df368112825f8d8f" 
#  instance_type = "t2.micro"
#  tags = {
#    Name = "client-instance"
#  }
#  associate_public_ip_address = true
#  key_name         = "VPS-key"
#  user_data = <<-EOF
#            #!/bin/bash -xe
#            curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3
#            systemctl enable periphery
#            EOF
#}

# configure the providers
provider "myuserprovider" {
  endpoint      = "http://localhost:6251/"
  #github_token   = "" # Changed to your github token
  #github_orgname = "contexttf" # Changed to your github orgname
}

# configure the resource
resource "myuserprovider_user" "client1_repo" {
  id               = "1"
  name             = "Client1"
  env_file_contents = "DOMAIN=mcpgateway.online\nEMAIL=ivobrett@iname.com\nADMIN_USERNAME=admin@mcpgateway.online"
  #server_ip        = aws_instance.example.public_ip
  server_ip        = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}