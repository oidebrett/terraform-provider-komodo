# Komodo Terraform Provider

The Komodo Terraform Provider allows you to manage your Komodo resources through Terraform. This provider enables infrastructure-as-code management of your Komodo deployments.

## Documentation

- [Komodo](https://komo.do/) - Detailed explanation of Komodo
- [Usage Guide](docs/CodeExplained.md) - Detailed explanation of the provider code
- [Examples](examples/) - Example configurations for different cloud providers

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)
- Docker (for local development)

## Example Usage

```hcl
terraform {
  required_providers {
    komodo-provider = {
      source = "example.com/me/komodo-provider"
      # version = "~> 1.0"
    }
  }
}

# Configure the provider
provider "komodo-provider" {
  endpoint = "http://your-komodo-api-endpoint:9120"
  api_key = "your-api-key"
  api_secret = "your-api-secret"
  github_token = "your-github-token"
}

# Create a user resource
resource "komodo-provider_user" "example" {
  id = "1"
  name = "Example Client"
  file_contents = templatefile("${path.module}/config-template.toml", {
    client_name = "Example Client"
    domain = "example.com"
    # Additional variables as needed
  })
  server_ip = "127.0.0.1"
}
```

The server_ip is the public IP address of the server that will be managed by Komodo. This is typically the IP address of the VM that is running the periphery agent. The [examples](examples/) show you how you can use terraform modules to provision a Cloud VM and then pass the IP address to the Komodo provider.

The `config-template.toml` file provides the template file for your resources syncs. The resources in here will be provisioned in Komodo. Heres a very simple example from the [Hello World Example](examples/hello_world/)

```toml
[[stack]]
name = "${client_name_lower}_stack"
[stack.config]
server = "server-${client_name_lower}"
file_contents = """
services:
  python-http:
    image: python:3.11-slim
    container_name: python-http
    working_dir: /app
    command: python -m http.server 80
    ports:
      - "80:80"
    restart: unless-stopped
"""

[[procedure]]
name = "${client_name}_ProcedureApply"
description = "This procedure runs the stack deployment"

[[procedure.config.stage]]
name = "${client_name}_Stack"
enabled = true
executions = [
  { execution.type = "DeployStack", execution.params.stack = "${client_name_lower}_stack", execution.params.services = [], enabled = true }
]

[[procedure]]
name = "${client_name}_ProcedureDestroy"
description = "This procedure destroys the stack deployment"

[[procedure.config.stage]]
name = "${client_name}_Stack"
enabled = true
executions = [
  { execution.type = "DestroyStack", execution.params.stack = "${client_name_lower}_stack", execution.params.services = [], execution.params.remove_orphans = false, enabled = true }
]

```
## Example Flow

                 ┌────────────────────┐
                 │   terraform apply  │
                 └────────┬───────────┘
                          │
          ┌───────────────▼─────────────────┐
          │     Cloud VM instance created   │
          │ (with Docker + periphery agent) │
          └───────────────┬─────────────────┘
                          │
            ┌─────────────▼─────────────┐
            │ Komodo Server resource    │
            │ created with VM's IP      │
            └─────────────┬─────────────┘
                          │
        ┌─────────────────▼──────────────────┐
        │  GitHub repo <client>_syncresources│
        │ created + resources.toml uploaded  │
        └─────────────────┬──────────────────┘
                          │
             ┌────────────▼────────────┐
             │  ResourceSync created   │
             │   (Sync'ed + Setup)     │
             └────────────┬────────────┘
                          │
              ┌───────────▼───────────────┐
              │   <client>_ProcedureApply │◄────┐
              │ (Runs DeployStack step)   │     │
              └───────────────────────────┘     │
                                                │
           ┌────────────────────────────┐       │
           │   terraform destroy        │       │
           └────────────┬───────────────┘       │
                        │                       │
            ┌───────────▼───────────────┐       │
            │ <client>_ProcedureDestroy │───────┘
            │ (Runs DestroyStack step)  │
            └───────────────────────────┘


## Authentication

The Komodo provider requires an endpoint URL, API keys and GitHub token for authentication:

### Static Credentials

```hcl
provider "komodo-provider" {
  endpoint = "http://your-komodo-api-endpoint:9120"
  api_key = "your-api-key"
  api_secret = "your-api-secret"
  github_token = "your-github-token"
}
```

### Environment Variables

You can provide your own environment variables for the Komodo provider. Have a look at the [GCP Examples](examples/gcp/) where additional variables are defined and set in the main.tf file.

## Getting Started

A very simple Komodo provider that spins up a Python http server can be found in :

- [Hello World Example](examples/hello_world/)

## Cloud Provider Integration

The Komodo provider can be used alongside major cloud providers to create complete infrastructure deployments:

- [AWS Examples](examples/aws/)
- [GCP Examples](examples/gcp/)
- [Azure Examples](examples/azure/)

## Building The Provider

1. Clone the repository
2. Build the provider using Go:
   ```bash
   go build -o bin/terraform-provider-komodo-provider
   ```

## Using Docker for Development

We provide a Docker environment for easy development:

```bash
# Build the Docker image
docker build -t timage .

# Start the container with your workspace mounted
docker run -it --volume .:/workspace --name tbox timage

# If you stop the container and want to restart it later:
docker start -ai tbox
```

## Using Terraform 

From within the Docker container, you can run Terraform commands. For example, to apply the Hello World example:

```bash
cd /workspace/examples/hello_world
terraform plan
terraform apply
```
