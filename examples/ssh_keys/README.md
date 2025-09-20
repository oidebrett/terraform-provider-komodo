# SSH Keys Example - Terraform Provider Komodo

## Overview

This example demonstrates how the Terraform Provider Komodo automatically generates SSH key pairs and embeds them into the GitHub repository file contents. This enables secure access to the repository using the generated private key.

## Features

- **Optional SSH Key Generation**: The provider can optionally generate an ed25519 SSH key pair
- **Deploy Key Upload**: The public key is automatically uploaded as a deploy key to the GitHub repository
- **Private Key Embedding**: The private key is embedded in the file contents for use by the application
- **Public Key Embedding**: The public key is also embedded for reference
- **Backward Compatibility**: SSH key generation is optional and can be disabled

## How It Works

1. When creating a GitHub repository with file contents and `generate_ssh_keys = true`, the provider:
   - Generates a new ed25519 SSH key pair
   - Uploads the public key as a deploy key to the repository
   - Embeds both keys in the environment section of the file contents

2. When `generate_ssh_keys = false` or is not specified, the provider:
   - Creates the repository and file contents without SSH keys
   - No deploy keys are uploaded
   - File contents remain unchanged

3. When SSH keys are generated, they are added to the `environment` section in this format:
   ```
   environment = """
   DOMAIN=example.com
   EMAIL=user@example.com
   ...existing environment variables...
   SSH_PRIVATE_KEY=-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----
   SSH_PUBLIC_KEY=ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...
   """
   ```

## Prerequisites

1. **GitHub Token**: A GitHub Personal Access Token with repository permissions
2. **Komodo API**: Access to a Komodo API endpoint
3. **Terraform**: Version 0.13 or newer

## Configuration

1. **Copy the example variables file**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit terraform.tfvars with your values**
   ```hcl
   # Komodo Configuration
   komodo_endpoint = "http://your-komodo-api:9120"
   komodo_api_key = "your-api-key"
   komodo_api_secret = "your-api-secret"
   
   # GitHub Configuration
   github_token = "ghp_your-github-token"
   github_orgname = "your-github-org"  # Optional
   
   # Instance Configuration
   client_name = "MyClient"
   client_id = "1"
   server_ip = "1.2.3.4"

   # SSH Key Configuration
   generate_ssh_keys = true  # Set to false to disable SSH key generation
   ```

## Enabling/Disabling SSH Key Generation

The SSH key generation feature is controlled by the `generate_ssh_keys` parameter:

- **`generate_ssh_keys = true`** (default): Generates SSH keys and uploads them as deploy keys
- **`generate_ssh_keys = false`**: Skips SSH key generation entirely

### Example with SSH keys enabled:
```hcl
resource "komodo-provider_user" "example" {
  id               = "1"
  name             = "MyClient"
  server_ip        = "1.2.3.4"
  generate_ssh_keys = true
  file_contents    = "..."
}
```

### Example with SSH keys disabled:
```hcl
resource "komodo-provider_user" "example" {
  id               = "1"
  name             = "MyClient"
  server_ip        = "1.2.3.4"
  generate_ssh_keys = false
  file_contents    = "..."
}
```

## Usage

1. **Initialize Terraform**
   ```bash
   terraform init
   ```

2. **Plan the deployment**
   ```bash
   terraform plan
   ```

3. **Apply the configuration**
   ```bash
   terraform apply
   ```

## What Gets Created

- A private GitHub repository named `{client_name_lower}_syncresources`
- A `resources.toml` file with your configuration and embedded SSH keys
- A deploy key in the repository for secure access
- Komodo server and resource sync configurations

## Accessing the SSH Keys

After deployment, you can:

1. **View the repository**: Check your GitHub account for the new repository
2. **Access the private key**: The private key is embedded in the `resources.toml` file
3. **Use the deploy key**: The public key is configured as a deploy key for repository access

## Security Notes

- The private key is embedded in the file contents and should be treated as sensitive
- The repository is created as private by default
- The deploy key provides read/write access to the repository
- Consider rotating keys periodically for enhanced security

## Cleanup

To destroy all created resources:
```bash
terraform destroy
```

This will remove:
- The GitHub repository
- All Komodo resources (server, syncs, procedures)
- The deploy key (automatically removed with repository deletion)
