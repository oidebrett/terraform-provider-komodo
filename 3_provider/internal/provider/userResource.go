package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

var _ tfresource.Resource = &UserResource{}
var _ tfresource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	client        *http.Client
	endpoint      string
	githubToken   string
	githubOrgname string // Changed from githubUsername
}

type UserModel struct {
	Id           tftypes.String `tfsdk:"id"`
	Name         tftypes.String `tfsdk:"name"`
	FileContents tftypes.String `tfsdk:"env_file_contents"`
	ServerIP     tftypes.String `tfsdk:"server_ip"`
}

func NewUserResource() tfresource.Resource {
	return &UserResource{}
}


func (r *UserResource) Metadata(ctx context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user" // matches in main.tf: resource "myuserprovider_user" "john_doe" {
}

func (r *UserResource) Schema(ctx context.Context, req tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
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
			"env_file_contents": tfschema.StringAttribute{
				MarkdownDescription: "Contents to write to contents.txt in the GitHub repository",
				Optional:            true,
			},
			"server_ip": tfschema.StringAttribute{
				MarkdownDescription: "The public IP address of the EC2 instance",
				Optional:            true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
	if req.ProviderData == nil { // this means the provider.go Configure method hasn't been called yet, so wait longer
		return
	}
	provider, ok := req.ProviderData.(*UserProvider)
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

func (r *UserResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var state UserModel
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
	
	// Create the user
	response, err := r.client.Post(
		r.endpoint+state.Id.ValueString(), 
		"application/text", 
		bytes.NewBuffer([]byte(state.ServerIP.ValueString())),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error sending request: %s", err))
		return
	}
	
	defer response.Body.Close()
	
	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))
		return
	}
	
	resp.State.Set(ctx, &state)
}

func (r *UserResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var state UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Get the user
	response, err := r.client.Get(r.endpoint + state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}
	defer response.Body.Close()
	
	if response.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	
	if response.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			resp.Diagnostics.AddError("Error reading response body", err.Error())
			return
		}
		state.ServerIP = tftypes.StringValue(string(bodyBytes))
		
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}
	
	resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received bad HTTP status: %s", response.Status))
}

func (r *UserResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// First, delete the GitHub repository
	err := r.deleteGitHubRepository(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Error deleting GitHub repository: %s", err))
		// Continue with the API call even if GitHub deletion fails
	}
	
	// Then proceed with your existing API call
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
	data.Id = tftypes.StringValue("")
	data.Name = tftypes.StringValue("")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var state UserModel
	var oldState UserModel
	
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
	
	// Update the user name
	webserviceCall, err := http.NewRequest("PUT", r.endpoint+state.Id.ValueString(), bytes.NewBuffer([]byte(state.Name.ValueString())))
	if err != nil {
		resp.Diagnostics.AddError("Go Error", fmt.Sprintf("Error creating request: %s", err))
		return
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
	}
	
	resp.State.Set(ctx, &state)
}

func (r *UserResource) ImportState(ctx context.Context, req tfresource.ImportStateRequest, resp *tfresource.ImportStateResponse) {
	tfresource.ImportStatePassthroughID(ctx, tfpath.Root("id"), req, resp)
}

// Add this new method to create GitHub repository
func (r *UserResource) createGitHubRepository(ctx context.Context, repoName string, fileContents string) error {
	// Sanitize the repository name
	sanitizedName := sanitizeRepoName(repoName)
	
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
	
	// Create repository
	repo := &github.Repository{
		Name:        github.String(sanitizedName),
		Description: github.String("Created by custom Terraform provider"),
		Private:     github.Bool(false),
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
		
		// Create the file
		fileContent := []byte(fileContents)
		opts := &github.RepositoryContentFileOptions{
			Message:   github.String("Add contents.txt via Terraform"),
			Content:   fileContent,
			Branch:    github.String(defaultBranch),
			Committer: &github.CommitAuthor{
				Name:  github.String("Terraform Provider"),
				Email: github.String("terraform@example.com"),
			},
		}
		
		_, _, err = client.Repositories.CreateFile(ctx, repoOwner, sanitizedName, ".env", opts)
		if err != nil {
			return fmt.Errorf("failed to create file in repository: %v", err)
		}
	}
	
	return nil
}

// Delete GitHub repository
func (r *UserResource) deleteGitHubRepository(ctx context.Context, repoName string) error {
	// Sanitize the repository name
	sanitizedName := sanitizeRepoName(repoName)
	
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
func (r *UserResource) updateFileInRepository(ctx context.Context, repoName, owner, fileContents string) error {
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
		"contents.txt",
		&github.RepositoryContentGetOptions{Ref: defaultBranch},
	)
	
	// Create update options
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String("Update contents.txt via Terraform"),
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
	
	_, _, err = client.Repositories.CreateFile(ctx, owner, repoName, "contents.txt", opts)
	if err != nil {
		return fmt.Errorf("failed to update file in repository: %v", err)
	}
	
	return nil
}



