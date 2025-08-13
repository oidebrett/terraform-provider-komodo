# Developing the Komodo Terraform Provider

This document explains the development process and architecture of the Komodo Terraform Provider. It outlines the key components and design decisions made during development.

## Overview

The Komodo Terraform Provider was developed to enable infrastructure-as-code management of Komodo resources. Terraform is a powerful tool that allows users to manage multiple environments through declarative configuration files that can be stored in Git, providing benefits in terms of speed, safety, repeatability, auditing, and correctness.

## Development Environment

The development environment was set up using Docker to ensure consistency across different development machines. The Dockerfile includes:

- Alpine Linux as the base image
- Go programming language
- Terraform CLI
- Additional tools for development (curl, unzip, bash, etc.)

This containerized approach allows developers to work on the provider code while running Go in the container, with the local folder shared as a volume.

## Project Structure

The project is organized into the following key components:

- **Root directory**: Contains the main provider code and Go module definition
- **Examples directory**: Contains example Terraform configurations for different cloud providers
- **Internal/provider directory**: Contains the core provider implementation

## Terraform Configuration

The Terraform configuration files in the examples directory demonstrate how users can interact with the provider. These files:

1. Specify the required providers (including the Komodo provider)
2. Configure the provider with necessary credentials
3. Define resources to be managed by the provider

## Provider Implementation

The provider implementation consists of three main Go files:

### 1. Main Entry Point

The main.go file serves as the entry point for the provider. It creates a provider server that Terraform can connect to and use. When Terraform looks for the plugin to load it, this main function is called to get access to the provider.

### 2. Provider Definition

The provider.go file defines the provider structure and implements the required interfaces. It handles:

- Provider metadata and schema definition
- Configuration of the provider with user-supplied values
- Registration of resources that the provider can manage

### 3. Resource Implementation

The komodo_resource.go file implements the resource management functionality. This is where the core CRUD (Create, Read, Update, Delete) operations are defined.

## Resource Operations Summary

Instead of showing the full implementation details (which may change during development), here's a summary of what each resource operation does:

### Create Operation
- Processes the resource configuration from Terraform
- Creates a GitHub repository with the provided configuration file
- Creates a server resource in Komodo using the provided server IP
- Waits for the server to become available and enables it
- Creates resource syncs in Komodo to fetch configuration from GitHub
- Runs the initial sync to pull the configuration file
- Executes the ProcedureApply to deploy all stacks and services

### Read Operation
- Retrieves the current state of the resource
- Updates the Terraform state to match the remote state

### Update Operation
- Compares the current state with the desired state
- Updates the GitHub repository with new configuration if needed
- Updates Komodo resources as necessary
- Runs syncs and procedures to apply the changes

### Delete Operation
- Executes the ProcedureDestroy to clean up all deployed services
- Removes resource syncs from Komodo
- Removes the server from Komodo
- Deletes the GitHub repository

### Helper Functions
The implementation also includes several helper functions for:
- Interacting with the GitHub API
- Making API calls to Komodo
- Waiting for operations to complete
- Sanitizing repository names
- Error handling

## Development Considerations

During development, several key considerations were addressed:

1. **Authentication**: The provider needs GitHub credentials and Komodo endpoint information
2. **Error Handling**: Robust error handling to provide meaningful feedback to users
3. **State Management**: Proper management of Terraform state to track resources
4. **Idempotency**: Ensuring operations can be repeated without causing issues
5. **Cleanup**: Proper cleanup of resources during destroy operations

## Testing the Provider

For testing during development, a local Terraform configuration was used with the provider configured to use local development overrides. This approach allows testing without publishing the provider to a registry.

The testing process involves:
1. Building the provider binary
2. Configuring Terraform to use the local binary
3. Running Terraform operations (plan, apply, destroy)
4. Verifying that resources are created and managed correctly

## Conclusion

The Komodo Terraform Provider demonstrates how to create a custom Terraform provider that integrates with both a custom API (Komodo) and third-party services (GitHub). The provider enables users to manage complex deployments through simple, declarative configuration files.

By focusing on the core CRUD operations and providing clear integration with cloud providers, the Komodo Terraform Provider offers a powerful tool for managing Komodo resources in an infrastructure-as-code approach.


# Deep dive into the code

The remaining sections shows you how we created the komodo Terraform provider.

Terraform is a large, complicated piece of software, and the Terraform tutorials on creating a Terraform provider are lengthy and intimidating. But creating a provider doesn't have to be complicated.

In this article, we strip away many of the unnecessary functions that Terraform demonstrates and create a provider that does nothing but create, read, update, and delete a resource via an API to Komodo. You don't need any experience using Terraform to follow along — we'll explain everything as we go.

## Prerequisites

We needed [Docker](https://docs.docker.com/get-docker) to run the code provided here. You can install Terraform and Go locally if you prefer, but you'll need to adjust the commands we provide to suit your operating system.

## Set Up System

We created a folder on our computer to work in. We opened a terminal in the folder and ran the commands below to create a basic project structure.

```bash
touch Dockerfile
mkdir -p examples
mkdir -p internal/provider
```

The `examples` folder represents how our users will call Terraform to talk to your service. This folder will hold a Terraform resource configuration file.

The `internal/provider` folder contains the custom Terraform provider that will let Terraform talk to the Komodo API (and github). This provider will have three files in Go (Terraform uses only Go for plugins).

We added the text below to the `Dockerfile`. Please note to check for the latest version of Terraform and Go.

```dockerfile
FROM --platform=linux/amd64 alpine:3.19
WORKDIR /workspace
RUN apk add go curl unzip bash sudo nodejs npm vim
ENV GOPATH=/root/go
ENV PATH=$PATH:$GOPATH/bin
# install terraform:
RUN curl -O https://releases.hashicorp.com/terraform/1.12.2/terraform_1.12.2_linux_amd64.zip && \
    unzip terraform_1.12.2_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.12.2_linux_amd64.zip
```

Next we built the Docker image and started working in it using the commands below. The current folder will be shared with the Docker container as `/workspace`, so we can edit the files on our computer while running Go in the container.

```bash
docker build -t timage .
docker run -it --volume .:/workspace --name tbox timage
# if you stop the container and want to restart it later, run: docker start -ai tbox
```

## Create a Terraform Configuration File

So we have a komodo API, and in reality, you could use the API to call komodo directly. Why do you need Terraform, too?

In summary, Terraform allows your customers to manage multiple environments with a single service (Terraform) through declarative configuration files that can be stored in Git. This means that if one of your customers wants to add a stack or VPS, they can copy a Terraform resource configuration file from an existing conf, update it, check it into GitHub, and get it approved. Then Terraform can run it automatically using continuous integration. This has benefits for your customers in terms of speed, safety, repeatability, auditing, and correctness.

Next we created a Terraform configuration file to demonstrate this now. We ran the commands below:

```bash
cd /workspace/examples
touch main.tf
```

To get started we paste the code below into `main.tf`:

```go
# load the provider
terraform {
  required_providers {
    komodo-provider = {
      source = "registry.example.com/mattercoder/komodo-provider"
      version = ">= 1.0.0"
    }
  }
}

# configure the provider
provider "komodo-provider" {
  endpoint = "http://localhost:9120/" #this is the komodo API 
  github_token = your_github_token
}

# Custom provider resource with templated configuration
resource "komodo-provider_user" "client_syncresources" {
  id                = var.client_id
  name              = var.client_name
  file_contents = templatefile("./config-template.toml")  
  server_ip = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}
```

In the first section, we tell Terraform that it will need to use a custom provider to interact with our service, `example.com/me/komodo-provider`. We name the service `komodo-provider`.

In the second section, we configure this provider with the URL of the web service.

The final section is what our customers will use most. Here we create a resource (a user) with an ID and a name. You could create hundreds of resources here. Once the resources are created, you can also change their names or delete them, and Terraform will automatically make the appropriate calls to your service to ensure that the API matches the state it recorded locally.

This `main.tf` file is all our customers need to work with once we've created a provider. Let's create the provider now.

## Create a Custom Terraform Provider

Run the commands below:

```bash
cd /workspace
touch go.mod
```

Here we created `go.mod` manually because a Terraform provider needs a lot of dependencies. (The dependencies come from the [Terraform provider scaffolding project](https://github.com/hashicorp/terraform-provider-scaffolding-framework).)

Add the text below to `go.mod`.

```go
module example.com/me/komodo-provider

go 1.23.0

toolchain go1.24.4

require (
	github.com/google/go-github/v53 v53.2.0
	github.com/hashicorp/terraform-plugin-framework v1.15.0
	golang.org/x/oauth2 v0.26.0
)

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/terraform-plugin-go v0.27.0 // indirect
	github.com/hashicorp/terraform-plugin-log v0.9.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.5 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
```

Note the module name at the top of the file, module example.com/me/komodo-provider. This name consists of an example URL to make the module globally unique, and the name used for the provider in the main.tf file — komodo-provider.

There are only three code files that are essential to create a provider. They are each presented in a subsection below.

## The main.go File
The first file we need is main.go. We created it in /workspace/provider/main.go and added the code below to it:

```go
package main

import (
	"context"
	"log"

	"example.com/me/komodo-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "example.com/me/komodo-provider",
	}
	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
```

This file creates a providerserver, a server that hosts the provider plugin that Terraform can connect to and use. When Terraform looks for your plugin to load it, this main function is what Terraform calls to get access to the provider, created with provider.New().

Providers are structured like a Go web service. Functions receive a context, which holds state, a request, and a response. Functions can add data to the context that Terraform will use when the function exits. We’ll see an example of this when we create the resource file.

## The provider.go File
We created a internal/provider/provider.go file and add the code below to it:

```go
package provider

import (
	"context"
	"net/http"
	"strings"

	tfdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	tffunction "github.com/hashicorp/terraform-plugin-framework/function"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

type KomodoProviderModel struct {
	Endpoint     tftypes.String `tfsdk:"endpoint"`
	GithubToken  tftypes.String `tfsdk:"github_token"`
	GithubOrgname tftypes.String `tfsdk:"github_orgname"` // Changed from github_username
}

type KomodoProvider struct {
	endpoint      string
	githubToken   string
	githubOrgname string // Changed from githubUsername
	client        *http.Client
}

var _ tfprovider.Provider = &KomodoProvider{}
var _ tfprovider.ProviderWithFunctions = &KomodoProvider{}

func New() func() tfprovider.Provider {
	return func() tfprovider.Provider {
		return &KomodoProvider{}
	}
}

```

This code does the following:

- Imports the Terraform Go framework.
- Defines a KomodoProviderModel struct with an endpoint. This endpoint will come from the main.tf configuration file (the URL of your web service).
- Defines a KomodoProvider struct that holds any data the provider needs throughout its life. In our case, we need only the web service URL and an HTTP client that we can pass to the resource manager (created in the next section).
- Checks that KomodoProvider correctly implements all the functions Terraform needs in var _ tfprovider.Provider = &KomodoProvider{}. It creates a discarded _ variable and assigns it the type tfprovider.Provider so that the Go compiler can verify it.
- Defines a New() function to return an instance of our provider. This function was called in the previous file in the provider server.

## Next, we added the functions below to the provider.go file:

```go
func (p *KomodoProvider) Metadata(ctx context.Context, req tfprovider.MetadataRequest, resp *tfprovider.MetadataResponse) {
	resp.TypeName = "komodo-provider" // matches in your .tf file `resource "komodo-provider_user" "john_doe" {`
}

func (p *KomodoProvider) Schema(ctx context.Context, req tfprovider.SchemaRequest, resp *tfprovider.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"endpoint": tfschema.StringAttribute{
				Required:    true,
				Description: "The endpoint URL of the user service",
			},
			"github_token": tfschema.StringAttribute{
				Required:    true,
				Sensitive:   true, // Mark as sensitive to hide in logs
				Description: "GitHub personal access token",
			},
			"github_orgname": tfschema.StringAttribute{
				Optional:    true, // Make it optional
				Description: "GitHub organization name for repository creation",
			},
		},
	}
}

func (p *KomodoProvider) Configure(ctx context.Context, req tfprovider.ConfigureRequest, resp *tfprovider.ConfigureResponse) {
	var data KomodoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Ensure the endpoint ends with a trailing slash
	endpoint := data.Endpoint.ValueString()
	if !strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint + "/"
	}
	
	p.endpoint = endpoint
	p.githubToken = data.GithubToken.ValueString() // Store the GitHub token
	p.githubOrgname = data.GithubOrgname.ValueString() // Store the GitHub org name
	p.client = http.DefaultClient
	
	resp.DataSourceData = p
	resp.ResourceData = p
}

func (p *KomodoProvider) Resources(ctx context.Context) []func() tfresource.Resource {
	return []func() tfresource.Resource{
		NewKomodoResource,
	}
}

func (p *KomodoProvider) DataSources(ctx context.Context) []func() tfdatasource.DataSource {
	return []func() tfdatasource.DataSource{}
}

func (p *KomodoProvider) Functions(ctx context.Context) []func() tffunction.Function {
	return []func() tffunction.Function{}
}


```

- Metadata() contains the name of the provider.
- Schema() must match the main.tf file so that Terraform can get the configuration settings for the provider.
- Configure() gets the settings from the configuration file, creates an HTTP client, saves the settings to the KomodoProvider struct, and adds them to the method’s response type. We set ResourceData so that the resource manager has access to all the fields of the KomodoProvider struct.
- Resources() creates a single NewKomodoResource instance. The NewKomodoResource function returns a KomodoResource type, which is what interacts with the users in the web service, and we create it in the next subsection. Since our provider doesn’t manage any DataSources, we don’t create any.

## The komodo_resource.go File
We created a internal/provider/komodo_resource.go file and add the code below to it:

```go
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
)

var _ tfresource.Resource = &komodoResource{}
var _ tfresource.ResourceWithImportState = &komodoResource{}

type komodoResource struct {
	client        *http.Client
	endpoint      string
	githubToken   string
	githubOrgname string // Changed from githubUsername
}

type KomodoModel struct {
	Id           tftypes.String `tfsdk:"id"`
	Name         tftypes.String `tfsdk:"name"`
	FileContents tftypes.String `tfsdk:"file_contents"`
	ServerIP     tftypes.String `tfsdk:"server_ip"`
}

func NewKomodoResource() tfresource.Resource {
	return &komodoResource{}
}

```
This code is similar to the code in the previous file we created. It loads dependencies, checks the interfaces compile, and defines the struct the resource will use.

Note the KomodoModel. This struct is what will communicate between the web service and Terraform core. Terraform will save the values here for Id and Name into a local state file that mimics what Terraform thinks the web service state is. Terraform uses its own types to do this, terraform-plugin-framework/types, not plain Go types.

Next, we added the code below to allow the resource to configure itself:

```go
func (r *komodoResource) Metadata(ctx context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user" // matches in main.tf: resource "komodo-provider_user" "john_doe" {
}

func (r *komodoResource) Schema(ctx context.Context, req tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "User resource interacts with user web service",
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				MarkdownDescription: "The user ID",
				Required:            true,
			},
			"name": tfschema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Required:            true,
			},
			"file_contents": tfschema.StringAttribute{
				MarkdownDescription: "Contents to write to resources.toml in the GitHub repository",
				Optional:            true,
			},
			"server_ip": tfschema.StringAttribute{
				MarkdownDescription: "The public IP address of the EC2 instance",
				Optional:            true,
			},
		},
	}
}

func (r *komodoResource) Configure(ctx context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
	if req.ProviderData == nil { // this means the provider.go Configure method hasn't been called yet, so wait longer
		return
	}
	provider, ok := req.ProviderData.(*KomodoProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Could not create HTTP client",
			fmt.Sprintf("Expected *http.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = provider.client
	r.endpoint = provider.endpoint
	r.githubToken = provider.githubToken
	r.githubOrgname = provider.githubOrgname // Get the GitHub org name
}


```

Again, this code looks similar to the code in the previous file.

Note how the Metadata() function combines the provider and resource names with _ in komodo-provider_user. This matches the name in main.tf and is a Terraform naming standard.
- Schema() defines what Terraform will remember about the remote resource in local state.
- Configure() gets the information from the provider we configured in the provider.go file in the Configure() method, resp.ResourceData = p. It receives an HTTP client and URL from the provider to use in the resource manager.
T- he if req.ProviderData == nil line is essential. Terraform can load the resource manager before the provider, so when the Configure() function is called, there may not yet be a provider to get configuration data from. In this case, the function will exit, and Terraform will call it again later when the provider has been loaded. It seems strange that Terraform would call the resource manager before the provider since it seems that the provider owns the resource manager, but that’s just how it is.

The last code we needed to add to userProvider.go was the heart of the provider: Calling the web service with CRUD functions and returning the response to Terraform to update its state. This code is also the easiest to understand. We’ll explain the Create function after you’ve added the code below. The other functions are similar.

You can see the final code in [komodo_resource.go](https://github.com/oidebrett/terraform-provider-komodo/blob/main/internal/provider/komodo_resource.go) but we will summarize the key functionality below. This Go code defines a custom Terraform resource implementation (`komodoResource`) that orchestrates infrastructure and GitHub operations via API calls. Here's a summarized breakdown of what the code does:

#### **Create**

When a resource is created:

1. **(Optional)** Creates a **GitHub repository** named `<name>_syncresources`, with a `resources.toml` file if contents are provided.
2. **Creates a server** at a specified `ServerIP` using a custom API (`CreateServer`).
3. **Waits** for the server to be available (polls every 10s up to 3 minutes).
4. **Enables the server** via `UpdateServer`.
5. **Waits** until the server is in an "OK" state.
6. **Creates and runs "ContextWare" and "ResourceSetup" resource syncs** using the GitHub repo contents.
7. **Runs a procedure** named `<name>_ProcedureApply`.

#### **Read**

* Simply rehydrates the Terraform state from the last known state.
* Skips external API calls (read-only optimization).

#### **Update**

If the `FileContents` changes:

1. **Updates the GitHub repo** (`resources.toml`).
2. **Re-runs sync and procedure** steps to apply changes (as in `Create`).

#### **Delete**

Tears everything down:

1. **Runs a destroy procedure** (`<name>_ProcedureDestroy`).
2. **Deletes procedures**, **syncs**, and **server** with retry logic (for handling API race conditions).
3. **Deletes the GitHub repository**.
4. **Cleans up Terraform state**.


The Create function looks like a web handler, with a context, request, and response. As mentioned earlier, Terraform uses the web metaphor to structure its plugins. Like the other three functions, Create() does three things:

- Loads the Terraform state for the resource with req.Plan.Get(ctx, &state). This represents what Terraform thinks the remote resource is, or what it wants it to be.
- Calls the web service and gets the response with r.client.Post(r.endpoint+state.Id.ValueString().
- Saves the response to the local Terraform state with resp.State.Set(ctx, &state).
- Note that you don’t have to write any logic to reason about changing the remote state, for example, adding or updating the user if the response from the web service is not what you anticipated. That’s what Terraform Core is for. Terraform will call the correct sequence of CRUD functions to work out how to change the remote users based on your desired users in the configuration file.

Be careful to use only ValueString() when working with Terraform string types. There are similar functions, like String() and Value(), that can add extra " marks to your fields. You’ll encounter confusing errors with infinite update loops calling Terraform if you don’t notice that you’re adding extra string quotes to every web service call when you use the wrong method.

## Run the Provider

Let’s recap. We've:

- Created a main.tf Terraform configuration file to say that you want to use the komodo-provider provider to create a resource using the komodo API service.
- Created a Terraform provider with three files: a provider server, a provider, and a user resource manager.

Now it’s time to run Terraform pretending that we’re one of our customers calling Komodo and check that our provider works with the configuration file.

Because our provider isn’t hosted on the online Terraform registry, we need to tell Terraform to use the local project.

We created a file called .terraformrc in the workspace folder:

```sh
cd /workspace

touch .terraformrc
```

We inserted the text below:

```go
provider_installation {
    dev_overrides {
        "example.com/me/komodo-provider" = "/workspace/bin"
    }
    direct {} # For all other providers, install directly from their origin provider.
}

```

In the Docker terminal, we ran the command below to copy this Terraform settings file to the container home folder (where we’re user root), so that Terraform knows where to look for your provider.

```sh
cp /workspace/.terraformrc /root/
```

Now let’s run the provider and test it. Run the commands below.

```sh
cd /workspace/

go mod tidy # download dependencies

go build -o ./bin/terraform-provider-komodo-provider

cd /workspace/examples/hello_world

terraform plan

terraform apply -auto-approve
```

Terraform used the selected providers to generate an execution plan.

(If you’ve used Terraform before and are used to running terraform init, that won’t work with the dev_overrides setting. The Init command isn’t necessary because there’s no need to download any plugins.). When we encountered an error doing terraform init or plan. Something that looks like this

```shell
╷
│ Error: Invalid provider registry host
│ 
│ The host "example.com" given in provider source address
│ "example.com/me/komodo-provider" does not offer a Terraform provider
│ registry.
```

we then commentted out the komodo-provider section in the main.tf file and ran terraform init again. Then uncomment it and ran terraform plan again.


If you need to do any debugging while working on the provider, set the environment variable for logging in the terminal with export TF_LOG=WARN, and ask Terraform to write information to the terminal in your komodoResource.go with:

```go
import "github.com/hashicorp/terraform-plugin-log/tflog" // at the top

tflog.Info(ctx, "We are inside CREATE\n") // in a function
```

Notice that Terraform created /workspace/examples/hello_world/terraform.tfstate. This state file holds what Terraform thinks the remote state is. Never alter this file manually. If you need to update Terraform state because you added users directly through the web service, you’ll need to implement the Terraform import command.

*** Important ***
Notice that Terraform also created /workspace/examples/hello_world/terraform.tfstate.backup. This file holds that actual values of your variables. **DO NOT CHECK THIS INTO GITHUB**. It will contain sensitive information like your GitHub token.

## Gitignore

Here is the gitignore we used
```
# Go build artifacts
/bin/
/dist/
/build/
*.exe
*.test
*.out

# Go modules and package cache
/go.sum
/go.work
/go.work.sum

# Terraform plugin registry format (if using `make install`)
*.zip
*.tar.gz

# Terraform state files
*.tfstate
*.tfstate.backup
crash.log

# .terraform directory (plugin binaries, state, etc.)
**/.terraform/
.terraform.lock.hcl

# Editor/IDE specific
.vscode/
.idea/
*.swp
*.swo
*.DS_Store

# Test coverage and result files
coverage.*
*.coverprofile

# GitHub workflows if they are generated
/.github/

# Local environment config (if any)
.env
.env.*

# Logs
*.log

# Ignore your own local provider binary (name it below)
**/terraform-provider-komodo-provider*

**/your-service-account-key.json

**/terraform.tfstate
**/terraform.tfstate.backup

# .terraform variable values
**/terraform.tfvars
```
