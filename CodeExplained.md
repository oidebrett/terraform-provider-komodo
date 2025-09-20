# How To Create a Terraform Provider — a Guide for Absolute Beginners

This tutorial shows you how to create a simple Terraform provider for your web service.

Terraform is a large, complicated piece of software, and the Terraform tutorials on creating a Terraform provider are lengthy and intimidating. But creating a provider doesn't have to be complicated.

In this guide, we strip away many of the unnecessary functions that Terraform demonstrates and create a provider that does nothing but create, read, update, and delete a resource via an API. You don't need any experience using Terraform to follow along — we'll explain everything as we go.

## Prerequisites

You need [Docker](https://docs.docker.com/get-docker) to run the code provided here. You can install Terraform and Go locally if you prefer, but you'll need to adjust the commands we provide to suit your operating system.

## Set Up Your System

Create a folder on your computer to work in. Open a terminal in the folder and run the commands below to create a basic project structure.

```bash
touch Dockerfile
mkdir -p examples
mkdir -p internal/provider
```

The `examples` folder represents how your users will call Terraform to talk to your service. This folder will hold a Terraform resource configuration file.

The `internal/provider` folder contains the custom Terraform provider that will let Terraform talk to your web service. This provider will have three files in Go (Terraform uses only Go for plugins).

Add the text below to the `Dockerfile`.

```dockerfile
FROM --platform=linux/amd64 alpine:3.19
WORKDIR /workspace
RUN apk add go curl unzip bash sudo nodejs npm vim
ENV GOPATH=/root/go
ENV PATH=$PATH:$GOPATH/bin
# install terraform:
RUN curl -O https://releases.hashicorp.com/terraform/1.7.0/terraform_1.7.0_linux_amd64.zip && \
    unzip terraform_1.7.0_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.7.0_linux_amd64.zip
```

Now build the Docker image and start working in it using the commands below. Your current folder will be shared with the Docker container as `/workspace`, so you can edit the files on your computer while running Go in the container.

```bash
docker build -t timage .
docker run -it --volume .:/workspace --name tbox timage
# if you stop the container and want to restart it later, run: docker start -ai tbox
```

## Create a Terraform Configuration File

So we have a backup komodo API, and in reality, you could use the API to call your service. Why do you need Terraform, too?

In summary, Terraform allows your customers to manage multiple environments with a single service (Terraform) through declarative configuration files that can be stored in Git. This means that if one of your customers wants to add a stack or VPS, they can copy a Terraform resource configuration file from an existing conf, update it, check it into GitHub, and get it approved. Then Terraform can run it automatically using continuous integration. This has benefits for your customers in terms of speed, safety, repeatability, auditing, and correctness.

Let's create a Terraform configuration file to demonstrate this now. Run the commands below:

```bash
cd /workspace/examples
touch main.tf
```

Paste the code below into `main.tf`:

```hcl
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
  endpoint = "http://localhost:6251/"
}

# configure the resource
resource "komodo-provider_user" "john_doe" {
  id   = "1"
  name = "John Doe"
}
```

In the first section, we tell Terraform that it will need to use a custom provider to interact with our service, `example.com/me/komodo-provider`. We name the service `komodo-provider`.

In the second section, we configure this provider with the URL of the web service.

The final section is what your customers will use most. Here we create a resource (a user) with an ID and a name. You could create hundreds of users here. Once the users are created, you can also change their names or delete them, and Terraform will automatically make the appropriate calls to your service to ensure that the API matches the state it recorded locally.

This `main.tf` file is all your customers need to work with once you've created a provider. Let's create the provider now.

## Create a Custom Terraform Provider

Run the commands below:

```bash
cd /workspace
touch go.mod
```

Here we create `go.mod` manually because a Terraform provider needs a lot of dependencies. (The dependencies come from the [Terraform provider scaffolding project](https://github.com/hashicorp/terraform-provider-scaffolding-framework).)

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
The first file you need is main.go. Create it in /workspace/provider/main.go and add the code below to it:

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
Create a internal/provider/provider.go file and add the code below to it:

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

Imports the Terraform Go framework.
Defines a KomodoProviderModel struct with an endpoint. This endpoint will come from the main.tf configuration file (the URL of your web service).
Defines a KomodoProvider struct that holds any data the provider needs throughout its life. In our case, we need only the web service URL and an HTTP client that we can pass to the resource manager (created in the next section).
Checks that KomodoProvider correctly implements all the functions Terraform needs in var _ tfprovider.Provider = &KomodoProvider{}. It creates a discarded _ variable and assigns it the type tfprovider.Provider so that the Go compiler can verify it.
Defines a New() function to return an instance of our provider. This function was called in the previous file in the provider server.

## Next, add the functions below to the provider.go file:

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

Metadata() contains the name of the provider.
Schema() must match the main.tf file so that Terraform can get the configuration settings for the provider.
Configure() gets the settings from the configuration file, creates an HTTP client, saves the settings to the KomodoProvider struct, and adds them to the method’s response type. We set ResourceData so that the resource manager has access to all the fields of the KomodoProvider struct.
Resources() creates a single NewKomodoResource instance. The NewKomodoResource function returns a KomodoResource type, which is what interacts with the users in the web service, and we create it in the next subsection. Since our provider doesn’t manage any DataSources, we don’t create any.

## The komodoResource.go File
Create a internal/provider/komodo_resource.go file and add the code below to it:

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

Next, add the code below to allow the resource to configure itself:

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
Schema() defines what Terraform will remember about the remote resource in local state.
Configure() gets the information from the provider we configured in the provider.go file in the Configure() method, resp.ResourceData = p. It receives an HTTP client and URL from the provider to use in the resource manager.
The if req.ProviderData == nil line is essential. Terraform can load the resource manager before the provider, so when the Configure() function is called, there may not yet be a provider to get configuration data from. In this case, the function will exit, and Terraform will call it again later when the provider has been loaded. It seems strange that Terraform would call the resource manager before the provider since it seems that the provider owns the resource manager, but that’s just how it is.

The last code you need to add to userProvider.go is the heart of the provider: Calling the web service with CRUD functions and returning the response to Terraform to update its state. This code is also the easiest to understand. We’ll explain the Create function after you’ve added the code below. The other functions are similar.

```go
func (r *komodoResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var state KomodoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// First, create GitHub repository with file if contents provided
	fileContents := ""
	if !state.FileContents.IsNull() {
		fileContents = state.FileContents.ValueString()
	}
	
	err := r.createGitHubRepository(ctx, state.Name.ValueString(), fileContents)
	if err != nil {
		resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Error creating GitHub repository: %s", err))
		return
	}
	
	// Skip the user creation API call that was here before
	// We're keeping the endpoint for other API calls
	
	// 1. Create a server using the server_ip
	serverName := fmt.Sprintf("server-%s", strings.ToLower(state.Name.ValueString()))
	createServerPayload := fmt.Sprintf(`{
		"type": "CreateServer",
		"params": {
			"name": "%s",
			"config": {
				"address": "https://%s:8120"
			},
			"tags": ["%s"]
		}
	}`, serverName, state.ServerIP.ValueString(), state.Name.ValueString())

	err = r.makeAPICall(createServerPayload, r.endpoint+"write")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error creating server: %s", err))
		return
	}
	
	// Wait for the server to become available, checking every 10 seconds for up to 3 minutes
	err = r.waitForServerAvailability(serverName, 18, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError("Server Error", fmt.Sprintf("Error waiting for server to become available: %s", err))
		return
	}
	
	// Enable the server
	updateServerPayload := fmt.Sprintf(`{
		"type": "UpdateServer",
		"params": {
			"id": "%s",
			"config": {
				"enabled": true
			}
		}
	}`, serverName)
	
	err = r.makeAPICall(updateServerPayload, r.endpoint+"write")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error enabling server: %s", err))
		return
	}
	
	// Wait for the server to reach OK state, checking every 10 seconds for up to 3 minutes
	err = r.waitForServerStateEnabled(serverName, 18, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError("Server Error", fmt.Sprintf("Error waiting for server to reach OK state: %s", err))
		return
	}

	// Now make the additional API calls
	// 2. Create Resource Sync for ContextWare
	createContextWarePayload := fmt.Sprintf(`{
		"type": "CreateResourceSync",
		"params": {
			"name": "%s_ContextWare",
			"config": {
				"file_contents": "[[resource_sync]]\nname = \"%s_ResourceSetup\"\n[resource_sync.config]\nrepo = \"manidaecloud/%s_syncresources\"\ngit_account = \"manidaecloud\"\nresource_path = [\"resources.toml\"]"
			}
		}
	}`, state.Name.ValueString(), state.Name.ValueString(), strings.ToLower(state.Name.ValueString()))
	
	err = r.makeAPICall(createContextWarePayload, r.endpoint+"write")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error creating ContextWare resource sync: %s", err))
		return
	}
	
	// 3. Run the ContextWare sync first
	runContextWarePayload := fmt.Sprintf(`{
		"type": "RunSync",
		"params": {
			"sync": "%s_ContextWare"
		}
	}`, state.Name.ValueString())
	
	err = r.makeAPICall(runContextWarePayload, r.endpoint+"execute")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running ContextWare sync: %s", err))
		return
	}
	
	// Wait for the ContextWare sync to complete (up to 5 seconds)
	time.Sleep(5 * time.Second)
	
	// 4. Now run the ResourceSetup sync
	runResourceSetupPayload := fmt.Sprintf(`{
		"type": "RunSync",
		"params": {
			"sync": "%s_ResourceSetup"
		}
	}`, state.Name.ValueString())
	
	err = r.makeAPICall(runResourceSetupPayload, r.endpoint+"execute")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running ResourceSetup sync: %s", err))
		return
	}

	// Wait for the ResourceSetup sync to complete (up to 5 seconds)
	time.Sleep(5 * time.Second)

	// 5. Run Procedure
	runProcedurePayload := fmt.Sprintf(`{
		"type": "RunProcedure",
		"params": {
			"procedure": "%s_ProcedureApply"
		}
	}`, state.Name.ValueString())
	
	err = r.makeAPICall(runProcedurePayload, r.endpoint+"execute")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running procedure: %s", err))
		return
	}
	
	resp.State.Set(ctx, &state)
}

func (r *komodoResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var state KomodoModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Skip the user read API call
	// Just set the state directly
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *komodoResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var data KomodoModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Initialize random number generator
	rand.Seed(time.Now().UnixNano())
	
	// First, run the destroy procedure
	runDestroyPayload := fmt.Sprintf(`{
		"type": "RunProcedure",
		"params": {
			"procedure": "%s_ProcedureDestroy"
		}
	}`, data.Name.ValueString())
	
	err := r.makeAPICall(runDestroyPayload, r.endpoint+"execute")
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running destroy procedure: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Wait for the ProcedureDestroy to complete (up to 5 seconds)
	time.Sleep(5 * time.Second)

	// Function to retry API calls with backoff
	retryAPICall := func(payload, url string, maxRetries int) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			err := r.makeAPICall(payload, url)
			if err == nil {
				return nil
			}
			
			lastErr = err
			// If we get a "Procedure busy" error, wait and retry
			if strings.Contains(err.Error(), "Procedure busy") {
				// Longer exponential backoff with more randomness
				sleepTime := time.Duration(3*(i+1)+rand.Intn(3)) * time.Second
				time.Sleep(sleepTime)
				continue
			}
			
			// For other errors, don't retry
			return err
		}
		return lastErr
	}
	
	// Delete the procedures with retry
	deleteProcedureApplyPayload := fmt.Sprintf(`{
		"type": "DeleteProcedure",
		"params": {
			"id": "%s_ProcedureApply"
		}
	}`, data.Name.ValueString())
	
	err = retryAPICall(deleteProcedureApplyPayload, r.endpoint+"write", 5)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting apply procedure after retries: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Wait a bit between procedure deletions
	time.Sleep(2 * time.Second)
	
	deleteProcedureDestroyPayload := fmt.Sprintf(`{
		"type": "DeleteProcedure",
		"params": {
			"id": "%s_ProcedureDestroy"
		}
	}`, data.Name.ValueString())
	
	err = retryAPICall(deleteProcedureDestroyPayload, r.endpoint+"write", 5)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting destroy procedure after retries: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Wait a bit before deleting resource syncs
	time.Sleep(2 * time.Second)
	
	// Delete the resource syncs with retry
	deleteResourceSetupSyncPayload := fmt.Sprintf(`{
		"type": "DeleteResourceSync",
		"params": {
			"id": "%s_ResourceSetup"
		}
	}`, data.Name.ValueString())
	
	err = retryAPICall(deleteResourceSetupSyncPayload, r.endpoint+"write", 5)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting ResourceSetup sync after retries: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Wait a bit between resource sync deletions
	time.Sleep(2 * time.Second)
	
	deleteContextWareSyncPayload := fmt.Sprintf(`{
		"type": "DeleteResourceSync",
		"params": {
			"id": "%s_ContextWare"
		}
	}`, data.Name.ValueString())
	
	err = retryAPICall(deleteContextWareSyncPayload, r.endpoint+"write", 5)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting ContextWare sync after retries: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Delete the server
	serverName := fmt.Sprintf("server-%s", strings.ToLower(data.Name.ValueString()))
	deleteServerPayload := fmt.Sprintf(`{
		"type": "DeleteServer",
		"params": {
			"id": "%s"
		}
	}`, serverName)
	
	err = retryAPICall(deleteServerPayload, r.endpoint+"write", 3)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting server after retries: %s", err))
		// Continue with deletion even if API call fails
	}
	
	// Delete the GitHub repository
	err = r.deleteGitHubRepository(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Error deleting GitHub repository: %s", err))
		// Continue with the API call even if GitHub deletion fails
	}
	
	// Skip the user deletion API call
	// Just clear the state
	data.Id = tftypes.StringValue("")
	data.Name = tftypes.StringValue("")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *komodoResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var state KomodoModel
	var oldState KomodoModel
	
	// Get the current state
	resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Get the planned new state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Skip the user update API call that was here before
	// We're keeping the endpoint for other API calls
	
	// Update GitHub repository file if needed
	if !state.FileContents.IsNull() && !state.FileContents.Equal(oldState.FileContents) {
		// Determine the owner (org or user)
		owner := ""
		if r.githubOrgname != "" {
			owner = r.githubOrgname
		} else {
			// Get the authenticated user
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: r.githubToken},
			)
			tc := oauth2.NewClient(ctx, ts)
			client := github.NewClient(tc)
			
			user, _, err := client.Users.Get(ctx, "")
			if err != nil {
				resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Failed to get authenticated user: %v", err))
				return
			}
			owner = *user.Login
		}
		
		err := r.updateFileInRepository(ctx, sanitizeRepoName(state.Name.ValueString()), owner, state.FileContents.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Error updating file in repository: %s", err))
			return
		}
		
		// Run the API calls again to update the resources
		// 1. Create/Update Resource Sync
		createSyncPayload := fmt.Sprintf(`{
			"type": "CreateResourceSync",
			"params": {
				"name": "%s_ContextWare",
				"config": {
					"file_contents": "[[resource_sync]]\nname = \"%s_ResourceSetup\"\n[resource_sync.config]\nrepo = \"manidaecloud/%s_syncresources\"\ngit_account = \"manidaecloud\"\nresource_path = [\"resources.toml\"]"
				}
			}
		}`, state.Name.ValueString(), state.Name.ValueString(), strings.ToLower(state.Name.ValueString()))
		
		err = r.makeAPICall(createSyncPayload, r.endpoint+"write")
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error updating resource sync: %s", err))
			return
		}
		
		// 2. Run Sync
		runSyncPayload := fmt.Sprintf(`{
			"type": "RunSync",
			"params": {
				"sync": "%s_ResourceSetup"
			}
		}`, state.Name.ValueString())
		
		err = r.makeAPICall(runSyncPayload, r.endpoint+"execute")
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running sync: %s", err))
			return
		}
		
		// 3. Run Procedure
		runProcedurePayload := fmt.Sprintf(`{
			"type": "RunProcedure",
			"params": {
				"procedure": "%s_ProcedureApply"
			}
		}`, state.Name.ValueString())
		
		err = r.makeAPICall(runProcedurePayload, r.endpoint+"execute")
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error running procedure: %s", err))
			return
		}
	}
	
	resp.State.Set(ctx, &state)
}

func (r *komodoResource) ImportState(ctx context.Context, req tfresource.ImportStateRequest, resp *tfresource.ImportStateResponse) {
	tfresource.ImportStatePassthroughID(ctx, tfpath.Root("id"), req, resp)
}

// Add this new method to create GitHub repository
func (r *komodoResource) createGitHubRepository(ctx context.Context, repoName string, fileContents string) error {
	// Sanitize the repository name and append "_syncresources"
	sanitizedName := sanitizeRepoName(repoName) + "_syncresources"
	
	// Check if token is available
	if r.githubToken == "" {
		return fmt.Errorf("GitHub token is not set. Please provide it via configuration or GITHUB_TOKEN environment variable")
	}
	
	// Use the token from the provider configuration
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	
	// Create repository - set Private to true
	repo := &github.Repository{
		Name:        github.String(sanitizedName),
		Description: github.String("Created by custom Terraform provider"),
		Private:     github.Bool(true), // Changed to true to make the repository private
		AutoInit:    github.Bool(true), // Initialize with README so we have a main branch
	}
	
	// If org name is provided, create in that organization
	// Otherwise, create in the authenticated user's account
	owner := ""
	if r.githubOrgname != "" {
		owner = r.githubOrgname
	}
	
	_, _, err := client.Repositories.Create(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub repository: %v", err)
	}
	
	// If file contents are provided, create the file
	if fileContents != "" {
		// Wait a moment for the repository to be fully initialized
		time.Sleep(2 * time.Second)
		
		// Get the default branch name
		repoOwner := owner
		if repoOwner == "" {
			user, _, err := client.Users.Get(ctx, "")
			if err != nil {
				return fmt.Errorf("failed to get authenticated user: %v", err)
			}
			repoOwner = *user.Login
		}
		
		// Get the repository to determine the default branch
		repoInfo, _, err := client.Repositories.Get(ctx, repoOwner, sanitizedName)
		if err != nil {
			return fmt.Errorf("failed to get repository info: %v", err)
		}
		
		defaultBranch := "main"
		if repoInfo.DefaultBranch != nil {
			defaultBranch = *repoInfo.DefaultBranch
		}
		
		// Create the file - changed from contents.txt to resources.toml
		fileContent := []byte(fileContents)
		opts := &github.RepositoryContentFileOptions{
			Message:   github.String("Add resources.toml via Terraform"),
			Content:   fileContent,
			Branch:    github.String(defaultBranch),
			Committer: &github.CommitAuthor{
				Name:  github.String("Terraform Provider"),
				Email: github.String("terraform@example.com"),
			},
		}
		
		_, _, err = client.Repositories.CreateFile(ctx, repoOwner, sanitizedName, "resources.toml", opts)
		if err != nil {
			return fmt.Errorf("failed to create file in repository: %v", err)
		}
	}
	
	return nil
}

// Delete GitHub repository
func (r *komodoResource) deleteGitHubRepository(ctx context.Context, repoName string) error {
	// Sanitize the repository name and append _syncresources
	sanitizedName := sanitizeRepoName(repoName) + "_syncresources"
	
	// Use the token from the provider configuration
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	
	// Determine the owner (org or user)
	owner := ""
	if r.githubOrgname != "" {
		owner = r.githubOrgname
	} else {
		// Get the authenticated user
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to get authenticated user: %v", err)
		}
		owner = *user.Login
	}
	
	// Delete the repository
	_, err := client.Repositories.Delete(ctx, owner, sanitizedName)
	if err != nil {
		return fmt.Errorf("failed to delete GitHub repository: %v", err)
	}
	
	return nil
}

// Helper function to sanitize repository names
func sanitizeRepoName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	
	// Remove special characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	name = reg.ReplaceAllString(name, "")
	
	// Convert to lowercase
	name = strings.ToLower(name)
	
	return name
}

// Add this new method to update the file in the repository
func (r *komodoResource) updateFileInRepository(ctx context.Context, repoName, owner, fileContents string) error {
	// Append _syncresources to the repo name
	repoName = repoName + "_syncresources"
	
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	
	// Get the repository to determine the default branch
	repoInfo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %v", err)
	}
	
	defaultBranch := "main"
	if repoInfo.DefaultBranch != nil {
		defaultBranch = *repoInfo.DefaultBranch
	}
	
	// Get the current file to get its SHA
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repoName,
		"resources.toml", // Changed from contents.txt to resources.toml
		&github.RepositoryContentGetOptions{Ref: defaultBranch},
	)
	
	// Create update options
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String("Update resources.toml via Terraform"),
		Content:   []byte(fileContents),
		Branch:    github.String(defaultBranch),
		Committer: &github.CommitAuthor{
			Name:  github.String("Terraform Provider"),
			Email: github.String("terraform@example.com"),
		},
	}
	
	// If the file exists, include its SHA
	if err == nil && fileContent != nil {
		opts.SHA = fileContent.SHA
	}
	
	_, _, err = client.Repositories.CreateFile(ctx, owner, repoName, "resources.toml", opts)
	if err != nil {
		return fmt.Errorf("failed to update file in repository: %v", err)
	}
	
	return nil
}

// Helper method to make API calls
func (r *komodoResource) makeAPICall(payload string, url string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", "REMOVED")
	req.Header.Set("X-Api-Secret", "REMOVED")
	
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK HTTP status: %s, body: %s", resp.Status, string(bodyBytes))
	}
	
	return nil
}

// Add this helper function to check if a server is available
func (r *komodoResource) waitForServerAvailability(serverName string, maxAttempts int, sleepDuration time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create the GetServer payload
		getServerPayload := fmt.Sprintf(`{
			"type": "GetServer",
			"params": {
				"server": "%s"
			}
		}`, serverName)
		
		// Make the API call
		client := &http.Client{}
		req, err := http.NewRequest("POST", r.endpoint+"read", bytes.NewBuffer([]byte(getServerPayload)))
		if err != nil {
			return fmt.Errorf("error creating request: %s", err)
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "REMOVED")
		req.Header.Set("X-Api-Secret", "REMOVED")
		
		resp, err := client.Do(req)
		if err != nil {
			// If there's an error, wait and try again
			time.Sleep(sleepDuration)
			continue
		}
		
		defer resp.Body.Close()
		
		// If we get a 200 OK, the server exists
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %s", err)
			}
			
			// Parse the response to check if the server is properly configured
			var result map[string]interface{}
			err = json.Unmarshal(bodyBytes, &result)
			if err != nil {
				return fmt.Errorf("error parsing response: %s", err)
			}
			
			// Check if the server has the expected configuration
			if config, ok := result["config"].(map[string]interface{}); ok {
				if _, ok := config["address"]; ok {
					// Server exists and has an address configured
					return nil
				}
			}
		}
		
		// Wait before the next attempt
		time.Sleep(sleepDuration)
	}
	
	return fmt.Errorf("server %s did not become available after %d attempts", serverName, maxAttempts)
}

// Add this helper function to check if a server is in OK state
func (r *komodoResource) waitForServerStateEnabled(serverName string, maxAttempts int, sleepDuration time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create the GetServerState payload
		getServerStatePayload := fmt.Sprintf(`{
			"type": "GetServerState",
			"params": {
				"server": "%s"
			}
		}`, serverName)
		
		// Make the API call
		client := &http.Client{}
		req, err := http.NewRequest("POST", r.endpoint+"read", bytes.NewBuffer([]byte(getServerStatePayload)))
		if err != nil {
			return fmt.Errorf("error creating request: %s", err)
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "REMOVED")
		req.Header.Set("X-Api-Secret", "REMOVED")
		
		resp, err := client.Do(req)
		if err != nil {
			// If there's an error, wait and try again
			time.Sleep(sleepDuration)
			continue
		}
		
		defer resp.Body.Close()
		
		// If we get a 200 OK, check the server state
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %s", err)
			}
			
			// Parse the response to check the server state
			var result map[string]interface{}
			err = json.Unmarshal(bodyBytes, &result)
			if err != nil {
				return fmt.Errorf("error parsing response: %s", err)
			}
			
			// Check if the server state is "Ok"
			if status, ok := result["status"].(string); ok && status == "Ok" {
				return nil
			}
		}
		
		// Wait before the next attempt
		time.Sleep(sleepDuration)
	}
	
	return fmt.Errorf("server %s did not reach OK state after %d attempts", serverName, maxAttempts)
}

```

The Create function looks like a web handler, with a context, request, and response. As mentioned earlier, Terraform uses the web metaphor to structure its plugins. Like the other three functions, Create() does three things:

Loads the Terraform state for the resource with req.Plan.Get(ctx, &state). This represents what Terraform thinks the remote resource is, or what it wants it to be.
Calls the web service and gets the response with r.client.Post(r.endpoint+state.Id.ValueString().
Saves the response to the local Terraform state with resp.State.Set(ctx, &state).
Note that you don’t have to write any logic to reason about changing the remote state, for example, adding or updating the user if the response from the web service is not what you anticipated. That’s what Terraform Core is for. Terraform will call the correct sequence of CRUD functions to work out how to change the remote users based on your desired users in the configuration file.

Be careful to use only ValueString() when working with Terraform string types. There are similar functions, like String() and Value(), that can add extra " marks to your fields. You’ll encounter confusing errors with infinite update loops calling Terraform if you don’t notice that you’re adding extra string quotes to every web service call when you use the wrong method.

Run the Provider

Let’s recapitulate. You’ve:

Created a one-file web service to manage users that represents your company’s product that you sell to customers.
Created a main.tf Terraform configuration file to say that you want to use the komodo-provider provider to create a user called “John Doe” using the web service.
Created a Terraform provider with three files: a provider server, a provider, and a user resource manager.
Now it’s time to run Terraform pretending that you’re one of your customers calling your web service and check that your provider works with the configuration file.

Because your provider isn’t hosted on the online Terraform registry, you need to tell Terraform to use the local project.

Create a file called .terraformrc in the workspace folder:

```sh
cd /workspace

touch .terraformrc
```

Insert the text below:

```go
provider_installation {
    dev_overrides {
        "example.com/me/komodo-provider" = "/workspace/provider/bin"
    }
    direct {} # For all other providers, install directly from their origin provider.
}

```

In the Docker terminal, run the command below to copy this Terraform settings file to the container home folder (where you’re user root), so that Terraform knows where to look for your provider.

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

Terraform should return:

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:

create
Terraform will perform the following actions:

# komodo-provider_user.john_doe will be created

```
resource "komodo-provider_user" "john_doe" {

id = "1"

name = "John Doe"

}

Plan: 1 to add, 0 to change, 0 to destroy.

komodo-provider_user.john_doe: Creating...

POST: 1 John Doe

komodo-provider_user.john_doe: Creation complete after 0s [id=1]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

(If you’ve used Terraform before and are used to running terraform init, that won’t work with the dev_overrides setting. The Init command isn’t necessary because there’s no need to download any plugins.)

If you need to do any debugging while working on the provider, set the environment variable for logging in the terminal with export TF_LOG=WARN, and ask Terraform to write information to the terminal in your komodoResource.go with:

```go
import "github.com/hashicorp/terraform-plugin-log/tflog" // at the top

tflog.Info(ctx, "We are inside CREATE\n") // in a function
```

Notice that Terraform created /workspace/examples/terraform.tfstate. This state file holds what Terraform thinks the remote state is. Never alter this file manually. If you need to update Terraform state because you added users directly through the web service, you’ll need to implement the Terraform import command.

