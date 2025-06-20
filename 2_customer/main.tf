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
  region      = "us-central1"
  zone        = "us-central1-a"
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
  machine_type = "e2-medium"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-2004-focal-v20240307b"
    }
  }

  network_interface {
    network       = "default"
    access_config {} # Needed for a public IP
  }

  metadata = {
#    ssh-keys = "ubuntu:ssh-rsa YOURSSHKEYHERE ubuntu"
    ssh-keys = "ubuntu:ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8FMj1FPHBLvj5ov22HN3Cy+/nzrBVq7AlxxcceanRCB+J3qdAg3K6MBvFj3K0+YF9eUw7LeGt16TDmz3OefrRLpCip65iKr1pugzdf3y2pW/+qAZNvtjNGTJpaaFA+jMgkv3wDuZp/6cZJa0BEr39VM3Pj1APfiOP88Q2XWchorZWcgrUTXL9IQh4AmTK0WRsaaAeC+SyWys94BNrEOfOCd/TQZh58Ogljkr//LlzbkF+O9IHP3AHgDZS/SGCOD9C3ENtyVS8MnlR6LuvzEksIbi5Ner3Mm3qkjctb5xqLPe7Dc1O5MBcCQDQHv70e5IR8SLW/Kei6BaV3jdOREjZegbSKoc5O7z8JaVHfU4pOTkp41O6lf0eCCmdPGi/cE5Y5+EjloJu8sJfctAYHY0ZmDJgEWeLd3ot3jbu4Bq+i5sDhdUXEGU9/zdu+v/598iqCcYGwJ+qH9qU1lsWnGETOl6IXjLrjeIoCj4n+c/303cg8fxAXf7GT5sjQJcZHH0= ubuntu"
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e

    # Log all output to a file for debugging
    exec > >(tee /var/log/user-data.log) 2>&1
    
    echo "Starting user data script execution at $(date)"

    apt-get update
    apt-get install -y \
      ca-certificates \
      curl \
      gnupg \
      lsb-release \
      git \
      postgresql-client

    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    echo \
    "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -y

    apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

    usermod -aG docker ubuntu

    newgrp docker

    echo "Starting komodo periphery at $(date)"
    # Run your setup script
    curl -sSL https://raw.githubusercontent.com/moghtech/komodo/main/scripts/setup-periphery.py | sudo python3
    systemctl enable periphery

    echo "User data script finished at $(date)"

  EOF

  tags = ["gcp-client"]
}

output "instance_public_ip" {
  value = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}

provider "aws" {
  region = "eu-west-1"
  # Add your AWS credentials or use environment variables
  access_key = "AKIAUOBTRCWEXB5AYW7E"
  secret_key = "URT1ODAc1rtiMqjQXCafmvHQOIokMlHgRvy2Mmxz"
}

#resource "aws_key_pair" "VPS-key" {
#  key_name   = "VPS-key"
#  public_key = "ssh-rsa YOURSSHKEYHERE"
#  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8FMj1FPHBLvj5ov22HN3Cy+/nzrBVq7AlxxcceanRCB+J3qdAg3K6MBvFj3K0+YF9eUw7LeGt16TDmz3OefrRLpCip65iKr1pugzdf3y2pW/+qAZNvtjNGTJpaaFA+jMgkv3wDuZp/6cZJa0BEr39VM3Pj1APfiOP88Q2XWchorZWcgrUTXL9IQh4AmTK0WRsaaAeC+SyWys94BNrEOfOCd/TQZh58Ogljkr//LlzbkF+O9IHP3AHgDZS/SGCOD9C3ENtyVS8MnlR6LuvzEksIbi5Ner3Mm3qkjctb5xqLPe7Dc1O5MBcCQDQHv70e5IR8SLW/Kei6BaV3jdOREjZegbSKoc5O7z8JaVHfU4pOTkp41O6lf0eCCmdPGi/cE5Y5+EjloJu8sJfctAYHY0ZmDJgEWeLd3ot3jbu4Bq+i5sDhdUXEGU9/zdu+v/598iqCcYGwJ+qH9qU1lsWnGETOl6IXjLrjeIoCj4n+c/303cg8fxAXf7GT5sjQJcZHH0= ivob@DESKTOP-J0S44C6"
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
  endpoint      = "http://198.12.108.91:9120"  # Works without trailing slash
  github_token  = "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" # Replace with your actual GitHub token
}

# configure the resource
resource "myuserprovider_user" "client1_syncresources" {
  id               = "1"
  name             = "Client1"
  env_file_contents = "[[stack]]\nname = \"client1_pangolin-setup\" \n[stack.config]\nserver = \"server-client1\"\nrepo = \"oidebrett/getcontextware\"\nreclone = true\nfile_paths = [\"docker-compose-setup.yml\"]\nenvironment = \"\"\"\nDOMAIN=mcpgateway.online\nEMAIL=ivobrett@iname.com\nADMIN_USERNAME=admin@mcpgateway.online\nADMIN_PASSWORD=Mcpgateway123q!\nADMIN_SUBDOMAIN=pangolin\nCROWDSEC_ENROLLMENT_KEY=cm9vtmyk3000pjx08brfsa6wd\nPOSTGRES_USER=admin\nPOSTGRES_PASSWORD=dDuScoEE53vA2Q==\nPOSTGRES_HOST=pangolin-postgres\nSTATIC_PAGE=true\nCLIENT_ID=252740974698-qjlglbdoh616kahvtht5d8flfsog16hv.apps.googleusercontent.com\nCLIENT_SECRET=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\n\"\"\"\n\n[[stack]]\nname = \"client1_pangolin-stack\" \n[stack.config]\nserver = \"server-client1\"\nfiles_on_host = true\nreclone = true\nrun_directory = \"/etc/komodo/stacks/client1_pangolin-setup\" \n\n[[procedure]]\nname = \"Client1_ProcedureApply\" \ndescription = \"This procedure runs the initial setup that write out a compose file for the main stack deployment\"\n\n[[procedure.config.stage]]\nname = \"Client1_Setup\" \nenabled = true\nexecutions = [\n  { execution.type = \"DeployStack\", execution.params.stack = \"client1_pangolin-setup\", execution.params.services = [], enabled = true } \n]\n\n[[procedure.config.stage]]\nname = \"Wait For Compose Write\"\nenabled = true\nexecutions = [\n  { execution.type = \"Sleep\", execution.params.duration_ms = 10000, enabled = true }\n]\n\n[[procedure.config.stage]]\nname = \"Client1_Stack\" \nenabled = true\nexecutions = [\n  { execution.type = \"DeployStack\", execution.params.stack = \"client1_pangolin-stack\", execution.params.services = [], enabled = true } \n]\n\n[[procedure]]\nname = \"Client1_ProcedureDestroy\" \n\n[[procedure.config.stage]]\nname = \"Client1_Stack\" \nenabled = true\nexecutions = [\n  { execution.type = \"DestroyStack\", execution.params.stack = \"client1_pangolin-stack\", execution.params.services = [], execution.params.remove_orphans = false, enabled = true } \n]\n\n[[procedure.config.stage]]\nname = \"Client1_Setup\" \nenabled = true\nexecutions = [\n  { execution.type = \"DestroyStack\", execution.params.stack = \"client1_pangolin-setup\", execution.params.services = [], execution.params.remove_orphans = false, enabled = true }\n]"
  #server_ip        = "127.0.0.1"
  server_ip        = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}
