# Komodo Docker Environment - Terraform Infrastructure

This repository contains Terraform configurations to deploy a Komodo-based Docker application environment on AWS, Google Cloud Platform (GCP), or Microsoft Azure. The infrastructure automatically sets up a virtual machine with Docker, configures the necessary networking, and deploys your application stack.

## 📁 Project Structure

```
terraform/
├── aws/                    # AWS-specific configuration
│   ├── main.tf            # Core AWS resources
│   ├── variables.tf       # AWS variable definitions
│   ├── outputs.tf         # AWS output values
│   ├── startup-script.sh  # AWS instance initialization
│   ├── config-template.toml # Application configuration template
│   └── terraform.tfvars.example # Example variables file
├── gcp/                    # GCP-specific configuration  
│   ├── main.tf            # Core GCP resources
│   ├── variables.tf       # GCP variable definitions
│   ├── outputs.tf         # GCP output values
│   ├── startup-script.sh  # GCP instance initialization
│   ├── config-template.toml # Application configuration template
│   └── terraform.tfvars.example # Example variables file
├── azure/                  # Azure-specific configuration
│   ├── main.tf            # Core Azure resources
│   ├── variables.tf       # Azure variable definitions
│   ├── outputs.tf         # Azure output values
│   ├── startup-script.sh  # Azure instance initialization
│   ├── config-template.toml # Application configuration template
│   └── terraform.tfvars.example # Example variables file
└── README.md              # This file
```

## 🚀 Quick Start

### Prerequisites

1. **Docker** - The only requirement! We provide a Docker container with Terraform pre-installed.
   ```bash
   # On macOS
   brew install docker
   
   # On Ubuntu/Debian
   sudo apt update && sudo apt install docker.io
   
   # On Windows
   # Download and install Docker Desktop from docker.com
   ```

2. **Cloud Provider Credentials** - You'll need credentials for your chosen provider
   - **AWS**: AWS Access Key and Secret Key
   - **GCP**: GCP Project ID and service account credentials
   - **Azure**: Azure Subscription ID, Tenant ID, Client ID, and Client Secret

### Docker Setup

Build the Docker image and start the container:

```bash
# Build the Docker image
docker build -t timage .

# Start the container with your workspace mounted
docker run -it --volume .:/workspace --name tbox timage

# If you stop the container and want to restart it later:
docker start -ai tbox
```

Your current folder will be shared with the Docker container as `/workspace`, so you can edit files on your computer while running Terraform inside the container.

### Choose Your Cloud Provider

**Inside the Docker container**, pick either AWS, GCP, or Azure and navigate to the corresponding directory:

```bash
# You're now inside the Docker container at /workspace
# Navigate to your chosen provider directory

# For AWS
cd terraform/aws

# For GCP  
cd terraform/gcp

# For Azure
cd terraform/azure
```

## ⚙️ Configuration

### Step 1: Create Your Variables File

Copy the example variables file and customize it:

```bash
cp terraform.tfvars.example terraform.tfvars
```

### Step 2: Configure Your Variables

Edit `terraform.tfvars` with your specific values:

#### AWS Configuration Example
```hcl
# AWS Configuration
aws_region     = "us-east-1"
aws_access_key = "YOUR_AWS_ACCESS_KEY"
aws_secret_key = "YOUR_AWS_SECRET_KEY"

# Instance Configuration
instance_type = "t3.medium"
key_name     = "my-key-pair"
ssh_public_key = "ssh-rsa AAAAB3NzaC1yc2E... your-public-key"

# Application Configuration
client_name    = "MyCompany"
domain        = "myapp.example.com"
admin_email   = "admin@example.com"
admin_password = "secure-password-123"

# GitHub Configuration
github_token = "ghp_your_github_token"
github_repo  = "https://github.com/yourusername/your-repo.git"

# Database Configuration
postgres_password = "secure-db-password"
```

#### GCP Configuration Example
```hcl
# GCP Configuration
gcp_project_id = "your-gcp-project-id"
gcp_region    = "us-central1"
gcp_zone      = "us-central1-a"

# Instance Configuration
machine_type = "e2-medium"
ssh_public_key = "ssh-rsa AAAAB3NzaC1yc2E... your-public-key"

# Application Configuration
client_name    = "MyCompany"
domain        = "myapp.example.com"
admin_email   = "admin@example.com"
admin_password = "secure-password-123"

# GitHub Configuration
github_token = "ghp_your_github_token"
github_repo  = "https://github.com/yourusername/your-repo.git"

# Database Configuration
postgres_password = "secure-db-password"
```

#### Azure Configuration Example
```hcl
# Azure Configuration
azure_subscription_id = "your-subscription-id"
azure_tenant_id       = "your-tenant-id"
azure_client_id       = "your-client-id"
azure_client_secret   = "your-client-secret"
azure_location        = "East US"

# Instance Configuration
instance_name = "azure-client-instance"
vm_size       = "Standard_B2s"
client_name   = "MyCompany"
client_id     = "1"

# SSH Configuration
ssh_public_key = "ssh-rsa AAAAB3NzaC1yc2E... your-public-key"
ssh_username   = "adminuser"

# Application Configuration
domain                  = "myapp.example.com"
admin_email             = "admin@example.com"
admin_subdomain         = "admin"
crowdsec_enrollment_key = "your-crowdsec-key"

# Database Configuration
postgres_user     = "admin"
postgres_password = "secure-db-password"
postgres_host     = "pangolin-postgres"

# GitHub Configuration
github_token = "ghp_your_github_token"
github_repo  = "https://github.com/yourusername/your-repo.git"
```

### Step 3: Set Up Authentication

**Note**: All these commands should be run inside the Docker container.

#### For AWS
```bash
# Option 1: Set environment variables (recommended for Docker)
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"

# Option 2: If you have AWS CLI configured on your host machine,
# you can mount your AWS credentials directory:
# docker run -it --volume .:/workspace --volume ~/.aws:/root/.aws --name tbox timage
```

#### For GCP
```bash
# Option 1: Use service account key file (recommended for Docker)
# Place your service account key file in the project directory and reference it:
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"

# Option 2: If you have gcloud configured on your host machine,
# you can mount your gcloud config:
# docker run -it --volume .:/workspace --volume ~/.config/gcloud:/root/.config/gcloud --name tbox timage
```

#### For Azure
```bash
# Option 1: Set environment variables (recommended for Docker)
export ARM_SUBSCRIPTION_ID="your-subscription-id"
export ARM_TENANT_ID="your-tenant-id"
export ARM_CLIENT_ID="your-client-id"
export ARM_CLIENT_SECRET="your-client-secret"

# Option 2: If you have Azure CLI configured on your host machine,
# you can mount your Azure config:
# docker run -it --volume .:/workspace --volume ~/.azure:/root/.azure --name tbox timage
```

### Azure Credentials Setup

To set up the required Azure credentials for Terraform, follow these steps:

1. **Find your Azure Tenant ID**:
   - Sign in to the [Azure portal](https://portal.azure.com)
   - Navigate to **Azure Active Directory** (or **Microsoft Entra ID**)
   - Go to **Properties**
   - The **Tenant ID** will be listed as the **Directory ID**

2. **Create an Application Registration**:
   - In **Microsoft Entra ID**, go to **App registrations**
   - Click **+ New registration**
   - Enter a name for your application
   - Select the appropriate supported account types
   - Click **Register**
   - Take note of the **Application (client) ID** - this will be your `azure_client_id`

3. **Create a Client Secret**:
   - In your newly created application, go to **Certificates & secrets**
   - Click **+ New client secret**
   - Add a description and select an expiration period
   - Click **Add**
   - **IMPORTANT**: Copy the **Value** of the secret immediately - this will be your `azure_client_secret` and will only be shown once

4. **Assign the Required Role**:
   - Open a terminal and use the Azure CLI to assign the Contributor role:
   ```bash
   az role assignment create \
     --assignee YOUR_CLIENT_ID \
     --role "Contributor" \
     --scope /subscriptions/YOUR_SUBSCRIPTION_ID
   ```
   - Replace `YOUR_CLIENT_ID` with your application's client ID
   - Replace `YOUR_SUBSCRIPTION_ID` with your Azure subscription ID
   - This grants your application permission to create resources in your subscription

## 🔧 Deployment

**All Terraform commands should be run inside the Docker container.**

### Initialize Terraform
```bash
terraform init
```

### Plan the Deployment
```bash
terraform plan
```

Review the planned changes to ensure everything looks correct.

### Apply the Configuration
```bash
terraform apply
```

Type `yes` when prompted to confirm the deployment.

## 📊 What Gets Created

### AWS Resources
- **EC2 Instance**: Virtual machine running Ubuntu with Docker
- **Security Group**: Firewall rules allowing HTTP/HTTPS and custom ports
- **Key Pair**: SSH key for instance access (if not existing)
- **Elastic IP**: Static public IP address (optional)

### GCP Resources
- **Compute Instance**: Virtual machine running Ubuntu with Docker
- **Firewall Rules**: Network security rules for application access
- **External IP**: Static public IP address

### Azure Resources
- **Resource Group**: Container for all Azure resources
- **Virtual Network & Subnet**: Network infrastructure for the VM
- **Network Interface**: Network connection for the VM
- **Public IP**: Static public IP address
- **Linux Virtual Machine**: Ubuntu VM with Docker installed
- **Network Security Group**: Firewall rules for VM access

### Common Features
- **Docker Installation**: Automatic Docker and Docker Compose setup
- **Application Deployment**: Your Komodo-based application stack
- **SSL/TLS Support**: Automatic certificate management
- **Database Setup**: PostgreSQL configuration
- **Monitoring**: Basic system monitoring setup

## 🔍 Accessing Your Application

After deployment, Terraform will output important information:

```bash
# View outputs (run inside the Docker container)
terraform output
```

You'll see:
- **instance_public_ip**: The public IP address of your server
- **instance_id**: The cloud provider instance identifier
- **ssh_command**: Command to SSH into your instance

### SSH Access
```bash
# SSH from your host machine (not from inside the Docker container)
ssh -i ~/.ssh/your-private-key ubuntu@<instance_public_ip>
```

### Application Access
- **Main Application**: `https://<your-domain>` or `http://<instance_public_ip>`
- **Admin Panel**: `https://<your-domain>/admin` or `http://<instance_public_ip>:8120`

## 🛠️ Customization

### Modifying Instance Specifications

Edit the variables in your `terraform.tfvars` file:

```hcl
# AWS
instance_type = "t3.large"  # Upgrade to larger instance

# GCP  
machine_type = "e2-standard-2"  # Upgrade to 2 vCPUs

# Azure
vm_size = "Standard_B4ms"  # Upgrade to larger instance
```

### Adding Custom Ports

Modify the `allowed_ports` variable:

```hcl
allowed_ports = [80, 443, 8120, 9000]  # Add port 9000
```

### Custom Startup Scripts

The startup scripts can be modified to install additional software or configure services. Edit `startup-script.sh` in your chosen provider directory.

## 🔒 Security Best Practices

### 1. Protect Sensitive Variables
Never commit `terraform.tfvars` to version control:

```bash
echo "terraform.tfvars" >> .gitignore
echo "*.tfstate*" >> .gitignore
echo ".terraform/" >> .gitignore
```

### 2. Use Environment Variables
For CI/CD or shared environments, use environment variables:

```bash
# For AWS
export TF_VAR_aws_access_key="your-access-key"
export TF_VAR_aws_secret_key="your-secret-key"

# For GCP
export TF_VAR_gcp_credentials_file="/path/to/credentials.json"

# For Azure
export TF_VAR_azure_client_id="your-client-id"
export TF_VAR_azure_client_secret="your-client-secret"
```

### 3. Restrict Network Access
Limit SSH and application access to specific IP ranges:

```hcl
# For AWS/GCP
ssh_cidr_blocks = ["203.0.113.0/24"]  # Your office IP range
app_cidr_blocks = ["0.0.0.0/0"]       # Public access or restrict as needed

# For Azure
firewall_source_ranges = ["203.0.113.0/24"]  # Your office IP range
```

### 4. Regular Updates
Keep your Terraform provider versions updated and review security groups regularly.

## 🐛 Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify your cloud provider credentials are correctly configured
   - Check that your project/account has necessary permissions
   - For Azure: Ensure the service principal has the Contributor role

2. **SSH Key Issues**
   - Ensure your SSH public key is correctly formatted
   - Verify the corresponding private key is available locally

3. **Port Access Issues**
   - Check that security groups/firewall rules are properly configured
   - Verify the application is running on the expected ports

4. **Instance Boot Issues**
   - SSH into the instance and check `/var/log/user-data.log` or `/var/log/startup.log`
   - Verify Docker installation completed successfully

### Viewing Logs

SSH into your instance and check various logs:

```bash
# User data execution log
sudo tail -f /var/log/user-data.log  # AWS/GCP
sudo tail -f /var/log/startup.log    # Azure

# Docker service status
sudo systemctl status docker

# Application logs
sudo docker logs <container-name>
```

## 🔄 Updates and Maintenance

### Updating the Infrastructure
```bash
# Pull latest changes (on your host machine)
git pull origin main

# Enter the Docker container
docker start -ai tbox

# Inside the container, plan and apply updates
terraform plan
terraform apply
```

### Scaling Resources
Modify instance types or add additional resources by updating variables and running:

```bash
# Inside the Docker container
terraform plan
terraform apply
```

## 🗑️ Cleanup

To destroy all created resources:

```bash
# Inside the Docker container
terraform destroy
```

**⚠️ Warning**: This will permanently delete all resources created by Terraform. Make sure to backup any important data first.

## 📞 Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review Terraform and cloud provider documentation
3. Check application-specific logs on the deployed instance
4. Ensure all prerequisites are properly installed and configured

## 📋 Variable Reference

### Required Variables
- `client_name`: Your organization/client name
- `domain`: Your application domain
- `admin_email`: Administrator email address
- `github_repo`: Your application repository URL

### Optional Variables with Defaults
- Instance sizing and configuration
- Network ports and access rules
- Database and application settings
- Cloud provider specific settings

See the `variables.tf` file in your chosen provider directory for complete variable documentation.

---

**Happy Deploying! 🚀**
