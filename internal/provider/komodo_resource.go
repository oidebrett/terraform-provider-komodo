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
				"file_contents": "[[resource_sync]]\nname = \"%s_ResourceSetup\"\n[resource_sync.config]\nrepo = \"oidebrett/%s_syncresources\"\ngit_account = \"oidebrett\"\nresource_path = [\"resources.toml\"]"
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
					"file_contents": "[[resource_sync]]\nname = \"%s_ResourceSetup\"\n[resource_sync.config]\nrepo = \"oidebrett/%s_syncresources\"\ngit_account = \"oidebrett\"\nresource_path = [\"resources.toml\"]"
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
