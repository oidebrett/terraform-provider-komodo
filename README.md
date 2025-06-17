How To Create a Terraform Provider — a Guide for Absolute Beginners
===================================================================

![Tristan Cartledge](/_next/image?url=%2Fassets%2Fauthor-headshots%2Ftristan.jpeg&w=1920&q=75)

Tristan Cartledge

March 13, 2024

![How To Create a Terraform Provider — a Guide for Absolute Beginners](/_next/image?url=%2Fassets%2Fapi-advice-how-to-create-a-TF-provider.png&w=1920&q=85)

This tutorial shows you how to create a simple Terraform provider for your web service.

Terraform is a large, complicated piece of software, and the Terraform tutorials on creating a Terraform provider are lengthy and intimidating. But creating a provider doesn’t have to be complicated.

In this guide, we strip away many of the unnecessary functions that Terraform demonstrates and create a provider that does nothing but create, read, update, and delete a resource via an API. You don’t need any experience using Terraform to follow along — we’ll explain everything as we go.

Prerequisites[](#prerequisites)
-------------------------------

You need [Docker](https://docs.docker.com/get-docker)  to run the code provided here. You can install Terraform and Go locally if you prefer, but you’ll need to adjust the commands we provide to suit your operating system.

Set Up Your System[](#set-up-your-system)
-----------------------------------------

Create a folder on your computer to work in. Open a terminal in the folder and run the commands below to create a basic project structure.

touch Dockerfile

mkdir 1\_webservice

mkdir 2\_customer

mkdir \-p 3\_provider/internal/provider

The `1_webservice` folder represents the service that your company sells online. In this example, we’ll have an API that can add and remove users. This service can be in any programming language.

The `2_customer` folder represents how your users will call Terraform to talk to your service. This folder will hold a Terraform resource configuration file.

The `3_provider` folder is the custom Terraform provider that will let Terraform talk to your web service. This provider will have three files in Go (Terraform uses only Go for plugins).

Add the text below to the `Dockerfile`.

FROM \--platform=linux/amd64 alpine:3.19

WORKDIR /workspace

RUN apk add go curl unzip bash sudo nodejs npm vim

ENV GOPATH=/root/go

ENV PATH=$PATH:$GOPATH/bin

\# install terraform:

RUN curl \-O https://releases.hashicorp.com/terraform/1.7.0/terraform\_1.7.0\_linux\_amd64.zip && \\

    unzip terraform\_1.7.0\_linux\_amd64.zip && \\

    mv terraform /usr/local/bin/ && \\

    rm terraform\_1.7.0\_linux\_amd64.zip

Now build the Docker image and start working in it using the commands below. Your current folder will be shared with the Docker container as `/workspace`, so you can edit the files on your computer while running Go in the container.

docker build \-t timage .

docker run \-it --volume .:/workspace \--name tbox timage

\# if you stop the container and want to restart it later, run: docker start -ai tbox

Create the Web Service[](#create-the-web-service)
-------------------------------------------------

Since we’ll write the Terraform provider in Go, let’s create a basic web service in Go, too.

Terraform uses CRUD (create, read, update, and delete) operations to manage any resource in any system, from AWS and Azure to your company’s finance, software, or healthcare product.

In the `1_webservice` folder, create a single-file service that allows a customer to create users with an ID and name, and update and delete users. Run the commands below in the Docker container terminal:

cd /workspace/1\_webservice

go mod init main

go get github.com/go-chi/chi/v5

touch main.go

These commands create a `go.mod` file in the folder and add the `chi` web framework dependency.

In a text editor, copy the code below into `main.go`.

package main

import (

	"fmt"

	"io"

	"net/http"

	"sync"

	"github.com/go-chi/chi/v5"

)

var users \= make(map\[string\]string) // Map to store users with id as key and name as value

var mutex \= &sync.RWMutex{}         // Mutex to protect access to the map as server is multithreaded

func main() {

	router := chi.NewRouter()

	router.Post("/{id}", func(response http.ResponseWriter, request \*http.Request) {

		id := chi.URLParam(request, "id")

		name, err := io.ReadAll(request.Body)

		if err != nil {

			http.Error(response, "Failed to read request body", http.StatusBadRequest)

			return

		}

		mutex.Lock()

		defer mutex.Unlock()

		users\[id\] \= string(name)

		fmt.Fprintf(response, "%s", string(name))

		fmt.Println("POST: ", id, " ", string(name))

	})

This code imports parts of the Go standard library related to HTTP and the `chi` web framework.

The code then makes a variable called `users` to hold the IDs and names of users. As we aren’t using a database here, this will work only as long as the service is running, and all users will be lost when the service stops.

We then create a mutex to handle safe writing to the users lists. Go is multithreaded, so we need a mutex to allow each HTTP handler to check that no other thread is trying to update the users list at the same time.

Finally, we have an HTTP `POST` handler to set a user in the list. To keep this guide short, we do no fancy checks for existing users or errors. The service overwrites items in the list with `users[id] = string(name)`. The function returns plain text (not JSON) to the caller with `fmt.Fprintf(response, "%s", string(name))`. The last line prints to the console to show that it’s working.

So far, we have only a `Create` handler with the `Post` method. Let’s add read, update, and delete. Append the code below to `main.go`.

  router.Get("/{id}", func(response http.ResponseWriter, request \*http.Request) {

		id := chi.URLParam(request, "id")

		mutex.RLock()

		defer mutex.RUnlock()

		name, ok := users\[id\]

		if !ok {

			http.NotFound(response, request)

			return

		}

		fmt.Fprintf(response, "%s", name)

		fmt.Println("GET: ", id, " ", name)

	})

	router.Put("/{id}", func(response http.ResponseWriter, request \*http.Request) {

		id := chi.URLParam(request, "id")

		name, err := io.ReadAll(request.Body)

		if err != nil {

			http.Error(response, "Failed to read request body", http.StatusBadRequest)

			return

		}

		mutex.Lock()

		defer mutex.Unlock()

		if \_, ok := users\[id\]; !ok {

			http.NotFound(response, request)

			return

		}

		users\[id\] \= string(name)

		fmt.Fprintf(response, "%s", string(name))

		fmt.Println("PUT: ", id, " ", users\[id\])

	})

	router.Delete("/{id}", func(response http.ResponseWriter, request \*http.Request) {

		id := chi.URLParam(request, "id")

		mutex.Lock()

		defer mutex.Unlock()

		name, ok := users\[id\]

		if !ok {

			http.NotFound(response, request)

			return

		}

		delete(users, id)

		fmt.Fprintf(response, "%s", name)

		fmt.Println("PUT: ", id, " ", users\[id\])

	})

	http.ListenAndServe(":6251", router)

}

Test the service by running the commands below.

go run main.go &

curl \-X POST \-d "Jane" http://localhost:6251/2

curl http://localhost:6251/2

This series of commands starts the service in the background, creates a user `[2,Jane]`, and retrieves the user from the service.

To stop and restart the background service, run the command below:

killall main; go run main.go &

Create a Terraform Configuration File[](#create-a-terraform-configuration-file)
-------------------------------------------------------------------------------

So you have a web service, and in reality, you might even have an SDK in Python, Go, Java, and other languages that your customers could use to call your service. Why do you need Terraform, too?

We answer this question in detail in [our blog post about using Terraform as a SaaS API interface](/post/build-terraform-providers). In summary, Terraform allows your customers to manage multiple environments with a single service (Terraform) through declarative configuration files that can be stored in Git. This means that if one of your customers wants to add a new user or a whole new franchise, they can copy a Terraform resource configuration file from an existing franchise, update it, check it into GitHub, and get it approved. Then Terraform can run it automatically using continuous integration. This has benefits for your customers in terms of speed, safety, repeatability, auditing, and correctness.

Let’s create a Terraform configuration file to demonstrate this now. Run the commands below:

cd /workspace/2\_customer

touch main.tf

Paste the code below into `main.tf`:

\# load the provider

terraform {

  required\_providers {

    myuserprovider \= {

      source  \= "example.com/me/myuserprovider"

      \# version = "~> 1.0"

    }

  }

}

\# configure the provider

provider "myuserprovider" {

  endpoint \= "http://localhost:6251/"

}

\# configure the resource

resource "myuserprovider\_user" "john\_doe" {

  id   \= "1"

  name \= "John Doe"

}

In the first section, we tell Terraform that it will need to use a custom provider to interact with our service, `example.com/me/myuserprovider`. We name the service `myuserprovider`.

In the second section, we configure this provider with the URL of the web service.

The final section is what your customers will use most. Here we create a resource (a user) with an ID and a name. You could create hundreds of users here. Once the users are created, you can also change their names or delete them, and Terraform will automatically make the appropriate calls to your service to ensure that the API matches the state it recorded locally.

This `main.tf` file is all your customers need to work with once you’ve created a provider. Let’s create the provider now.

Create a Custom Terraform Provider[](#create-a-custom-terraform-provider)
-------------------------------------------------------------------------

Run the commands below:

cd /workspace/3\_provider

touch go.mod

Here we create `go.mod` manually because a Terraform provider needs a lot of dependencies. (The dependencies come from the [Terraform provider scaffolding project](https://github.com/hashicorp/terraform-provider-scaffolding-framework) .)

Add the text below to `go.mod`.

module example.com/me/myuserprovider

go 1.21

require (

	github.com/hashicorp/go-version v1.6.0

	github.com/hashicorp/terraform\-plugin\-docs v0.18.0

	github.com/hashicorp/terraform\-plugin\-framework v1.6.1

	github.com/hashicorp/terraform\-plugin\-go v0.22.0

	github.com/hashicorp/terraform\-plugin\-log v0.9.0

	github.com/hashicorp/terraform\-plugin\-testing v1.7.0

)

require (

	github.com/Kunde21/markdownfmt/v3 v3.1.0 // indirect

	github.com/Masterminds/goutils v1.1.1 // indirect

	github.com/Masterminds/semver/v3 v3.2.0 // indirect

	github.com/Masterminds/sprig/v3 v3.2.3 // indirect

	github.com/ProtonMail/go-crypto v1.1.0\-alpha.0 // indirect

	github.com/agext/levenshtein v1.2.2 // indirect

	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect

	github.com/armon/go-radix v1.0.0 // indirect

	github.com/bgentry/speakeasy v0.1.0 // indirect

	github.com/cloudflare/circl v1.3.7 // indirect

	github.com/fatih/color v1.16.0 // indirect

	github.com/golang/protobuf v1.5.3 // indirect

	github.com/google/go-cmp v0.6.0 // indirect

	github.com/google/uuid v1.4.0 // indirect

	github.com/hashicorp/cli v1.1.6 // indirect

	github.com/hashicorp/errwrap v1.1.0 // indirect

	github.com/hashicorp/go-checkpoint v0.5.0 // indirect

	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect

	github.com/hashicorp/go-cty v1.4.1\-0.20200414143053\-d3edf31b6320 // indirect

	github.com/hashicorp/go-hclog v1.6.2 // indirect

	github.com/hashicorp/go-multierror v1.1.1 // indirect

	github.com/hashicorp/go-plugin v1.6.0 // indirect

	github.com/hashicorp/go-uuid v1.0.3 // indirect

	github.com/hashicorp/hc\-install v0.6.3 // indirect

	github.com/hashicorp/hcl/v2 v2.20.0 // indirect

	github.com/hashicorp/logutils v1.0.0 // indirect

	github.com/hashicorp/terraform\-exec v0.20.0 // indirect

	github.com/hashicorp/terraform\-json v0.21.0 // indirect

	github.com/hashicorp/terraform\-plugin\-sdk/v2 v2.33.0 // indirect

	github.com/hashicorp/terraform\-registry\-address v0.2.3 // indirect

	github.com/hashicorp/terraform\-svchost v0.1.1 // indirect

	github.com/hashicorp/yamux v0.1.1 // indirect

	github.com/huandu/xstrings v1.3.3 // indirect

	github.com/imdario/mergo v0.3.15 // indirect

	github.com/mattn/go-colorable v0.1.13 // indirect

	github.com/mattn/go-isatty v0.0.20 // indirect

	github.com/mattn/go-runewidth v0.0.9 // indirect

	github.com/mitchellh/copystructure v1.2.0 // indirect

	github.com/mitchellh/go-testing\-interface v1.14.1 // indirect

	github.com/mitchellh/go-wordwrap v1.0.0 // indirect

	github.com/mitchellh/mapstructure v1.5.0 // indirect

	github.com/mitchellh/reflectwalk v1.0.2 // indirect

	github.com/oklog/run v1.0.0 // indirect

	github.com/posener/complete v1.2.3 // indirect

	github.com/russross/blackfriday v1.6.0 // indirect

	github.com/shopspring/decimal v1.3.1 // indirect

	github.com/spf13/cast v1.5.0 // indirect

	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect

	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect

	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect

	github.com/yuin/goldmark v1.6.0 // indirect

	github.com/yuin/goldmark\-meta v1.1.0 // indirect

	github.com/zclconf/go-cty v1.14.3 // indirect

	golang.org/x/crypto v0.21.0 // indirect

	golang.org/x/exp v0.0.0\-20230809150735\-7b3493d9a819 // indirect

	golang.org/x/mod v0.15.0 // indirect

	golang.org/x/net v0.21.0 // indirect

	golang.org/x/sys v0.18.0 // indirect

	golang.org/x/text v0.14.0 // indirect

	golang.org/x/tools v0.13.0 // indirect

	google.golang.org/appengine v1.6.8 // indirect

	google.golang.org/genproto/googleapis/rpc v0.0.0\-20231106174013\-bbf56f31fb17 // indirect

	google.golang.org/grpc v1.61.1 // indirect

	google.golang.org/protobuf v1.32.0 // indirect

	gopkg.in/yaml.v2 v2.3.0 // indirect

)

Note the module name at the top of the file, `module example.com/me/myuserprovider`. This name consists of an example URL to make the module globally unique, and the name used for the provider in the `main.tf` file — `myuserprovider`.

There are only three code files that are essential to create a provider. They are each presented in a subsection below.

### The `main.go` File[](#the-maingo-file)

The first file you need is `main.go`. Create it in `/workspace/3_provider/main.go` and add the code below to it:

package main

import (

	"context"

	"log"

	"example.com/me/myuserprovider/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

)

func main() {

	opts := providerserver.ServeOpts{

		Address: "example.com/me/myuserprovider",

	}

	err := providerserver.Serve(context.Background(), provider.New(), opts)

	if err != nil {

		log.Fatal(err.Error())

	}

}

This file creates a `providerserver`, a server that hosts the provider plugin that Terraform can connect to and use. When Terraform looks for your plugin to load it, this `main` function is what Terraform calls to get access to the provider, created with `provider.New()`.

Providers are structured like a Go web service. Functions receive a `context`, which holds state, a request, and a response. Functions can add data to the `context` that Terraform will use when the function exits. We’ll see an example of this when we create the resource file.

### The `provider.go` File[](#the-providergo-file)

Create a `3_provider/internal/provider/provider.go` file and add the code below to it:

package provider

import (

	"context"

	"net/http"

	tfdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	tffunction "github.com/hashicorp/terraform-plugin-framework/function"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"

	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	tftypes "github.com/hashicorp/terraform-plugin-framework/types"

)

type UserProviderModel struct {

	Endpoint tftypes.String \`tfsdk:"endpoint"\`

}

type UserProvider struct {

	endpoint string

	client   \*http.Client

}

var \_ tfprovider.Provider \= &UserProvider{}

var \_ tfprovider.ProviderWithFunctions \= &UserProvider{}

func New() func() tfprovider.Provider {

	return func() tfprovider.Provider {

		return &UserProvider{}

	}

}

This code does the following:

*   Imports the Terraform Go framework.
*   Defines a `UserProviderModel` struct with an `endpoint`. This endpoint will come from the `main.tf` configuration file (the URL of your web service).
*   Defines a `UserProvider` struct that holds any data the provider needs throughout its life. In our case, we need only the web service URL and an HTTP client that we can pass to the resource manager (created in the next section).
*   Checks that `UserProvider` correctly implements all the functions Terraform needs in `var _ tfprovider.Provider = &UserProvider{}`. It creates a discarded `_` variable and assigns it the type `tfprovider.Provider` so that the Go compiler can verify it.
*   Defines a `New()` function to return an instance of our provider. This function was called in the previous file in the provider server.

Next, add the functions below to the `provider.go` file:

func (p \*UserProvider) Metadata(ctx context.Context, req tfprovider.MetadataRequest, resp \*tfprovider.MetadataResponse) {

	resp.TypeName \= "myuserprovider" // matches in your .tf file \`resource "myuserprovider\_user" "john\_doe" {\`

}

func (p \*UserProvider) Schema(ctx context.Context, req tfprovider.SchemaRequest, resp \*tfprovider.SchemaResponse) {

	resp.Schema \= tfschema.Schema{

		Attributes: map\[string\]tfschema.Attribute{

			"endpoint": tfschema.StringAttribute{

				MarkdownDescription: "Endpoint of the API, e.g. - http://localhost:6251/",

				Required:            true,

			},

		},

	}

}

func (p \*UserProvider) Configure(ctx context.Context, req tfprovider.ConfigureRequest, resp \*tfprovider.ConfigureResponse) {

	var data UserProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		return

	}

	p.endpoint \= data.Endpoint.ValueString()

	p.client \= http.DefaultClient

	resp.DataSourceData \= p // will be usable by DataSources

	resp.ResourceData \= p   // will be usable by Resources

}

func (p \*UserProvider) Resources(ctx context.Context) \[\]func() tfresource.Resource {

	return \[\]func() tfresource.Resource{

		NewUserResource,

	}

}

func (p \*UserProvider) DataSources(ctx context.Context) \[\]func() tfdatasource.DataSource {

	return \[\]func() tfdatasource.DataSource{}

}

func (p \*UserProvider) Functions(ctx context.Context) \[\]func() tffunction.Function {

	return \[\]func() tffunction.Function{}

}

*   `Metadata()` contains the name of the provider.
*   `Schema()` must match the `main.tf` file so that Terraform can get the configuration settings for the provider.
*   `Configure()` gets the settings from the configuration file, creates an HTTP client, saves the settings to the `UserProvider` struct, and adds them to the method’s response type. We set `ResourceData` so that the resource manager has access to all the fields of the `UserProvider` struct.
*   `Resources()` creates a single `NewUserResource` instance. The `NewUserResource` function returns a `UserResource` type, which is what interacts with the users in the web service, and we create it in the next subsection. Since our provider doesn’t manage any `DataSources`, we don’t create any.

### The `userResource.go` File[](#the-userresourcego-file)

Create a `3_provider/internal/provider/userResource.go` file and add the code below to it:

package provider

import (

	"bytes"

	"context"

	"fmt"

	"io"

	"net/http"

	tfpath "github.com/hashicorp/terraform-plugin-framework/path"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"

	tftypes "github.com/hashicorp/terraform-plugin-framework/types"

)

var \_ tfresource.Resource \= &UserResource{}

var \_ tfresource.ResourceWithImportState \= &UserResource{}

type UserResource struct {

	client   \*http.Client

	endpoint string

}

type UserModel struct {

	Id   tftypes.String \`tfsdk:"id"\`

	Name tftypes.String \`tfsdk:"name"\`

}

func NewUserResource() tfresource.Resource {

	return &UserResource{}

}

This code is similar to the code in the previous file we created. It loads dependencies, checks the interfaces compile, and defines the struct the resource will use.

Note the `UserModel`. This struct is what will communicate between the web service and Terraform core. Terraform will save the values here for `Id` and `Name` into a local state file that mimics what Terraform thinks the web service state is. Terraform uses its own types to do this, `terraform-plugin-framework/types`, not plain Go types.

Next, add the code below to allow the resource to configure itself:

func (r \*UserResource) Metadata(ctx context.Context, req tfresource.MetadataRequest, resp \*tfresource.MetadataResponse) {

	resp.TypeName \= req.ProviderTypeName + "\_user" // matches in main.tf: resource "myuserprovider\_user" "john\_doe" {

}

func (r \*UserResource) Schema(ctx context.Context, req tfresource.SchemaRequest, resp \*tfresource.SchemaResponse) {

	resp.Schema \= tfschema.Schema{

		MarkdownDescription: "User resource interacts with user web service",

		Attributes: map\[string\]tfschema.Attribute{

			"id": tfschema.StringAttribute{

				MarkdownDescription: "The user ID",

				Required:            true,

			},

			"name": tfschema.StringAttribute{

				MarkdownDescription: "The name of the user",

				Required:            true,

			},

		},

	}

}

func (r \*UserResource) Configure(ctx context.Context, req tfresource.ConfigureRequest, resp \*tfresource.ConfigureResponse) {

	if req.ProviderData \== nil { // this means the provider.go Configure method hasn't been called yet, so wait longer

		return

	}

	provider, ok := req.ProviderData.(\*UserProvider)

	if !ok {

		resp.Diagnostics.AddError(

			"Could not create HTTP client",

			fmt.Sprintf("Expected \*http.Client, got: %T", req.ProviderData),

		)

		return

	}

	r.client \= provider.client

	r.endpoint \= provider.endpoint

}

Again, this code looks similar to the code in the previous file.

*   Note how the `Metadata()` function combines the provider and resource names with `_` in `myuserprovider_user`. This matches the name in `main.tf` and is a Terraform naming standard.
*   `Schema()` defines what Terraform will remember about the remote resource in local state.
*   `Configure()` gets the information from the provider we configured in the `provider.go` file in the `Configure()` method, `resp.ResourceData = p`. It receives an HTTP client and URL from the provider to use in the resource manager.

The `if req.ProviderData == nil` line is essential. Terraform can load the resource manager before the provider, so when the `Configure()` function is called, there may not yet be a provider to get configuration data from. In this case, the function will exit, and Terraform will call it again later when the provider has been loaded. It seems strange that Terraform would call the resource manager before the provider since it seems that the provider **owns** the resource manager, but that’s just how it is.

The last code you need to add to `userProvider.go` is the heart of the provider: Calling the web service with CRUD functions and returning the response to Terraform to update its state. This code is also the easiest to understand. We’ll explain the `Create` function after you’ve added the code below. The other functions are similar.

func (r \*UserResource) Create(ctx context.Context, req tfresource.CreateRequest, resp \*tfresource.CreateResponse) {

	var state UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {

		return

	}

	response, err := r.client.Post(r.endpoint+state.Id.ValueString(), "application/text", bytes.NewBuffer(\[\]byte(state.Name.ValueString())))

	if err != nil {

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error sending request: %s", err))

		return

	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))

		return

	}

	body, err := io.ReadAll(response.Body)

	if err != nil {

		resp.Diagnostics.AddError("Failed to Read Response Body", fmt.Sprintf("Could not read response body: %s", err))

		return

	}

	state.Name \= tftypes.StringValue(string(body))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r \*UserResource) Read(ctx context.Context, req tfresource.ReadRequest, resp \*tfresource.ReadResponse) {

	var state UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {

		return

	}

	response, err := r.client.Get(r.endpoint + state.Id.ValueString())

	if err != nil {

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))

		return

	}

	defer response.Body.Close()

	if response.StatusCode \== http.StatusNotFound {

		resp.State.RemoveResource(ctx)

		return

	}

	if response.StatusCode \== http.StatusOK {

		bodyBytes, err := io.ReadAll(response.Body)

		if err != nil {

			resp.Diagnostics.AddError("Error reading response body", err.Error())

			return

		}

		state.Name \= tftypes.StringValue(string(bodyBytes))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		return

	}

	resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received bad HTTP status: %s", response.Status))

}

func (r \*UserResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp \*tfresource.DeleteResponse) {

	var data UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		return

	}

	request, err := http.NewRequest(http.MethodDelete, r.endpoint+data.Id.ValueString(), nil)

	if err != nil {

		resp.Diagnostics.AddError("Request Creation Failed", fmt.Sprintf("Could not create HTTP request: %s", err))

		return

	}

	response, err := r.client.Do(request)

	if err != nil {

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))

		return

	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))

		return

	}

	data.Id \= tftypes.StringValue("")

	data.Name \= tftypes.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*UserResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp \*tfresource.UpdateResponse) {

	var state UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {

		return

	}

	webserviceCall, err := http.NewRequest("PUT", r.endpoint+state.Id.ValueString(), bytes.NewBuffer(\[\]byte(state.Name.ValueString())))

	if err != nil {

		resp.Diagnostics.AddError("Go Error", fmt.Sprintf("Error sending request: %s", err))

	}

	response, err := r.client.Do(webserviceCall)

	if err != nil {

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error sending request: %s", err))

		return

	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))

		return

	}

	body, err := io.ReadAll(response.Body)

	if err != nil {

		resp.Diagnostics.AddError("Failed to Read Response Body", fmt.Sprintf("Could not read response body: %s", err))

		return

	}

	state.Name \= tftypes.StringValue(string(body))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r \*UserResource) ImportState(ctx context.Context, req tfresource.ImportStateRequest, resp \*tfresource.ImportStateResponse) {

	tfresource.ImportStatePassthroughID(ctx, tfpath.Root("id"), req, resp)

}

The `Create` function looks like a web handler, with a context, request, and response. As mentioned earlier, Terraform uses the web metaphor to structure its plugins. Like the other three functions, `Create()` does three things:

*   Loads the Terraform state for the resource with `req.Plan.Get(ctx, &state)`. This represents what Terraform thinks the remote resource is, or what it wants it to be.
*   Calls the web service and gets the response with `r.client.Post(r.endpoint+state.Id.ValueString()`.
*   Saves the response to the local Terraform state with `resp.State.Set(ctx, &state)`.

Note that you don’t have to write any logic to reason about changing the remote state, for example, adding or updating the user if the response from the web service is not what you anticipated. That’s what Terraform Core is for. Terraform will call the correct sequence of CRUD functions to work out how to change the remote users based on your desired users in the configuration file.

Be careful to use only `ValueString()` when working with Terraform string types. There are similar functions, like `String()` and `Value()`, that can add extra `"` marks to your fields. You’ll encounter confusing errors with infinite update loops calling Terraform if you don’t notice that you’re adding extra string quotes to every web service call when you use the wrong method.

Run the Provider[](#run-the-provider)
-------------------------------------

Let’s recapitulate. You’ve:

*   Created a one-file web service to manage users that represents your company’s product that you sell to customers.
*   Created a `main.tf` Terraform configuration file to say that you want to use the `myuserprovider` provider to create a user called “John Doe” using the web service.
*   Created a Terraform provider with three files: a provider server, a provider, and a user resource manager.

Now it’s time to run Terraform pretending that you’re one of your customers calling your web service and check that your provider works with the configuration file.

Because your provider isn’t hosted on the online Terraform registry, you need to tell Terraform to use the local project.

Create a file called `.terraformrc` in the `workspace` folder:

cd /workspace

touch .terraformrc

Insert the text below:

provider\_installation {

    dev\_overrides {

        "example.com/me/myuserprovider" \= "/workspace/3\_provider/bin"

    }

    direct {} \# For all other providers, install directly from their origin provider.

}

In the Docker terminal, run the command below to copy this Terraform settings file to the container home folder (where you’re user `root`), so that Terraform knows where to look for your provider.

cp /workspace/.terraformrc /root/

Now let’s run the provider and test it. Run the commands below.

cd /workspace/3\_provider

go mod tidy \# download dependencies

go build \-o ./bin/terraform-provider-myuserprovider

cd /workspace/2\_customer

terraform plan

terraform apply \-auto-approve

Terraform should return:

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:

  + create

Terraform will perform the following actions:

  \# myuserprovider\_user.john\_doe will be created

  + resource "myuserprovider\_user" "john\_doe" {

      + id   = "1"

      + name = "John Doe"

    }

Plan: 1 to add, 0 to change, 0 to destroy.

myuserprovider\_user.john\_doe: Creating...

POST:  1   John Doe

myuserprovider\_user.john\_doe: Creation complete after 0s \[id=1\]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

(If you’ve used Terraform before and are used to running `terraform init`, that won’t work with the `dev_overrides` setting. The `Init` command isn’t necessary because there’s no need to download any plugins.)

If you need to do any debugging while working on the provider, set the environment variable for logging in the terminal with `export TF_LOG=WARN`, and ask Terraform to write information to the terminal in your `userResource.go` with:

import "github.com/hashicorp/terraform-plugin-log/tflog" // at the top

tflog.Info(ctx, "We are inside CREATE\\n") // in a function

Notice that Terraform created `/workspace/2_customer/terraform.tfstate`. This state file holds what Terraform thinks the remote state is. Never alter this file manually. If you need to update Terraform state because you added users directly through the web service, you’ll need to implement the Terraform `import` command.

Experiment to see how Terraform calls the CRUD functions depending on how you change your state. Add more users to the `main.tf` file, change their names, call `curl -X POST -d "Jane" http://localhost:6251/1` to try to confuse Terraform, and see how it handles the changes.

Limitations and Further Reading[](#limitations-and-further-reading)
-------------------------------------------------------------------

You’re done with writing code for this guide and now have a working minimal example of a Terraform provider that you can enhance. But this provider isn’t ready for production use yet. There are features you’ll probably want to add, for example:

*   Markup responses (JSON or XML). This simplistic web service currently returns either a 404 or a string containing a user name directly in the response body. In reality, you’ll use a markup language. You may even want to have your `userResource` call a Go SDK for your service instead of making web calls directly.
*   Versioning and continuous integration. Your web service will change over time. The provider will need to change to match it. Your customers will need to use the correct versions of each. You will also want to automatically build and release your provider from GitHub, using GitHub actions.
*   Testing. A real web service is complex, and you will need to write a lot of integration tests to ensure that every provider version you release does exactly what it’s supposed to when calling the service.
*   Documentation. Your customers want to know exactly how to set up and configure your provider to manage whatever resources your service offers.
*   Publishing the provider to the Terraform registry. Until you add metadata to your provider and release it in the Terraform ecosystem, no one can use it.
*   You also might want to add additional functionality, like handling data sources (which are different from resources) and external imports of resources.

If you want to learn how to enhance your provider, the best place to start is the official [Terraform provider creation tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider) . You can also clone the [provider scaffolding repository](https://github.com/hashicorp/terraform-provider-scaffolding-framework)  and read through it to see how Terraform structures a provider and uses `.github` to offer continuous integration.

Once you have worked through the tutorial, we recommend reading the theory on [Terraform plugins in the documentation](https://developer.hashicorp.com/terraform/plugin) . Especially promising is the 2024 HashiCorp release of an [automated provider generator](https://developer.hashicorp.com/terraform/plugin/code-generation)  from an OpenAPI schema or their custom specification language. Unfortunately, the HashiCorp provider generator is not ready for production use yet — you still need to write a lot of code yourself — but it’s something to watch. We have an article discussing its features [here](/post/how-to-build-terraform-providers).

A Simpler Way[](#a-simpler-way)
-------------------------------

You might feel that creating and maintaining your own Terraform provider is far too much work when you’re busy trying to run a business and provide your core service. Luckily, there is a much easier way. We at Speakeasy are passionate about and dedicated to making web APIs easy for customers to use. Our service can automatically generate a complete Terraform provider with documentation that’s ready to offer to your customers on the Terraform registry. All you need is an OpenAPI schema for your service and a few custom attributes.

Read about how you can create a Terraform provider with us in a few clicks in this [article](/post/how-to-build-terraform-providers) and see how we can massively reduce your workload.