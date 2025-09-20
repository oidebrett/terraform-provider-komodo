package provider

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

var _ tfresource.Resource = &komodoResource{}
var _ tfresource.ResourceWithImportState = &komodoResource{}

type komodoResource struct {
	client        *http.Client
	endpoint      string
	apiKey        string
	apiSecret     string
	githubToken   string
	githubOrgname string // Changed from githubUsername
}

type KomodoModel struct {
	Id               tftypes.String `tfsdk:"id"`
	Name             tftypes.String `tfsdk:"name"`
	FileContents     tftypes.String `tfsdk:"file_contents"`
	ServerIP         tftypes.String `tfsdk:"server_ip"`
	GenerateSSHKeys  tftypes.Bool   `tfsdk:"generate_ssh_keys"`
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
			"generate_ssh_keys": tfschema.BoolAttribute{
				MarkdownDescription: "Whether to generate SSH keys and upload them as deploy keys to the GitHub repository",
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
	r.apiKey = provider.apiKey
	r.apiSecret = provider.apiSecret
	r.githubToken = provider.githubToken
	r.githubOrgname = provider.githubOrgname // Get the GitHub org name
}

func (r *komodoResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var state KomodoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A slice of functions to execute for cleanup if something goes wrong.
	var cleanupTasks []func()

	// Defer the execution of cleanup tasks in case of an error.
	defer func() {
		if resp.Diagnostics.HasError() {
			// Run cleanup tasks in reverse order.
			for i := len(cleanupTasks) - 1; i >= 0; i-- {
				cleanupTasks[i]()
			}
		}
	}()

	// First, create GitHub repository with file if contents provided
	fileContents := ""
	if !state.FileContents.IsNull() {
		fileContents = state.FileContents.ValueString()
	}

	generateSSHKeys := false
	if !state.GenerateSSHKeys.IsNull() {
		generateSSHKeys = state.GenerateSSHKeys.ValueBool()
	}

	err := r.createGitHubRepository(ctx, state.Name.ValueString(), fileContents, generateSSHKeys)
	if err != nil {
		resp.Diagnostics.AddError("GitHub Error", fmt.Sprintf("Error creating GitHub repository: %s", err))
		return
	}
	cleanupTasks = append(cleanupTasks, func() {
		if err := r.deleteGitHubRepository(ctx, state.Name.ValueString()); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete GitHub repository during cleanup: %s", err))
		}
	})

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
	cleanupTasks = append(cleanupTasks, func() {
		deleteServerPayload := fmt.Sprintf(`{
			"type": "DeleteServer",
			"params": {
				"id": "%s"
			}
		}`, serverName)
		if err := r.makeAPICall(deleteServerPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete server during cleanup: %s", err))
		}
	})

	// Wait for the server to become available, checking every 10 seconds for up to 5 minutes
	err = r.waitForServerAvailability(serverName, 30, 10*time.Second)
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

	// Wait for the server to reach OK state, checking every 10 seconds for up to 5 minutes
	err = r.waitForServerStateEnabled(serverName, 30, 10*time.Second)
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
	cleanupTasks = append(cleanupTasks, func() {
		deleteContextWareSyncPayload := fmt.Sprintf(`{
			"type": "DeleteResourceSync",
			"params": {
				"id": "%s_ContextWare"
			}
		}`, state.Name.ValueString())
		if err := r.makeAPICall(deleteContextWareSyncPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete ContextWare sync during cleanup: %s", err))
		}
	})

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
	cleanupTasks = append(cleanupTasks, func() {
		deleteResourceSetupSyncPayload := fmt.Sprintf(`{
			"type": "DeleteResourceSync",
			"params": {
				"id": "%s_ResourceSetup"
			}
		}`, state.Name.ValueString())
		if err := r.makeAPICall(deleteResourceSetupSyncPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete ResourceSetup sync during cleanup: %s", err))
		}
	})

	// Wait for the ContextWare sync to complete (up to 15 seconds)
	time.Sleep(15 * time.Second)

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
	cleanupTasks = append(cleanupTasks, func() {
		deleteProcedureApplyPayload := fmt.Sprintf(`{
			"type": "DeleteProcedure",
			"params": {
				"id": "%s_ProcedureApply"
			}
		}`, state.Name.ValueString())
		if err := r.makeAPICall(deleteProcedureApplyPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete apply procedure during cleanup: %s", err))
		}

		deleteProcedureDestroyPayload := fmt.Sprintf(`{
			"type": "DeleteProcedure",
			"params": {
				"id": "%s_ProcedureDestroy"
			}
		}`, state.Name.ValueString())
		if err := r.makeAPICall(deleteProcedureDestroyPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete destroy procedure during cleanup: %s", err))
		}

		deleteProcedureRestartPayload := fmt.Sprintf(`{
			"type": "DeleteProcedure",
			"params": {
				"id": "%s_ProcedureRestart"
			}
		}`, state.Name.ValueString())
		if err := r.makeAPICall(deleteProcedureRestartPayload, r.endpoint+"write"); err != nil {
			resp.Diagnostics.AddWarning("Cleanup Warning", fmt.Sprintf("Failed to delete restart procedure during cleanup: %s", err))
		}
	})

	// Wait for the ResourceSetup sync to complete (up to 15 seconds)
	time.Sleep(15 * time.Second)

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
	mathrand.Seed(time.Now().UnixNano())
	
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
				sleepTime := time.Duration(3*(i+1)+mathrand.Intn(3)) * time.Second
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

	// Wait a bit between procedure deletions
	time.Sleep(2 * time.Second)

	deleteProcedureRestartPayload := fmt.Sprintf(`{
		"type": "DeleteProcedure",
		"params": {
			"id": "%s_ProcedureRestart"
		}
	}`, data.Name.ValueString())

	err = retryAPICall(deleteProcedureRestartPayload, r.endpoint+"write", 5)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting restart procedure after retries: %s", err))
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
		
		generateSSHKeys := false
		if !state.GenerateSSHKeys.IsNull() {
			generateSSHKeys = state.GenerateSSHKeys.ValueBool()
		}
		fmt.Sprintf("GenerateSSHKeys is: ",generateSSHKeys)

		err := r.updateFileInRepository(ctx, sanitizeRepoName(state.Name.ValueString()), owner, state.FileContents.ValueString(), generateSSHKeys)
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
func (r *komodoResource) createGitHubRepository(ctx context.Context, repoName string, fileContents string, generateSSHKeys bool) error {
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

	var privateKey, publicKey string

	// Generate SSH key pair if requested
	if generateSSHKeys {
		privateKey, publicKey, err = r.generateSSHKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate SSH key pair: %v", err)
		}

		// Upload the public key as a deploy key
		deployKeyTitle := fmt.Sprintf("terraform-deploy-key-%d", time.Now().Unix())
		err = r.uploadDeployKey(ctx, repoOwner, sanitizedName, publicKey, deployKeyTitle, false)
		if err != nil {
			return fmt.Errorf("failed to upload deploy key: %v", err)
		}
	}

	// If file contents are provided, create the file
	if fileContents != "" {
		updatedFileContents := fileContents

		// Add SSH keys to the file contents if they were generated
		if generateSSHKeys && privateKey != "" && publicKey != "" {
			updatedFileContents = r.addSSHKeysToFileContents(fileContents, privateKey, publicKey)
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
		fileContent := []byte(updatedFileContents)
		commitMessage := "Add resources.toml via Terraform"
		if generateSSHKeys && privateKey != "" && publicKey != "" {
			commitMessage = "Add resources.toml with SSH keys via Terraform"
		}

		opts := &github.RepositoryContentFileOptions{
			Message:   github.String(commitMessage),
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
func (r *komodoResource) updateFileInRepository(ctx context.Context, repoName, owner, fileContents string, generateSSHKeys bool) error {
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

	// Get the current file to get its SHA and extract existing SSH keys
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repoName,
		"resources.toml", // Changed from contents.txt to resources.toml
		&github.RepositoryContentGetOptions{Ref: defaultBranch},
	)

	updatedFileContents := fileContents

	// Handle SSH keys based on the generateSSHKeys flag
	if generateSSHKeys {
		// If the file exists, try to preserve existing SSH keys
		if err == nil && fileContent != nil {
			existingContent, err := fileContent.GetContent()
			if err == nil {
				// Extract SSH keys from existing content
				privateKeyPattern := `SSH_PRIVATE_KEY=([^\n]+)`
				publicKeyPattern := `SSH_PUBLIC_KEY=([^\n]+)`

				privateKeyRe := regexp.MustCompile(privateKeyPattern)
				publicKeyRe := regexp.MustCompile(publicKeyPattern)

				privateKeyMatch := privateKeyRe.FindStringSubmatch(existingContent)
				publicKeyMatch := publicKeyRe.FindStringSubmatch(existingContent)

				if len(privateKeyMatch) > 1 && len(publicKeyMatch) > 1 {
					// Use existing SSH keys
					privateKey := strings.ReplaceAll(privateKeyMatch[1], "\\n", "\n")
					publicKey := publicKeyMatch[1]
					updatedFileContents = r.addSSHKeysToFileContents(fileContents, privateKey, publicKey)
				} else {
					// Generate new SSH keys if none exist
					privateKey, publicKey, err := r.generateSSHKeyPair()
					if err != nil {
						return fmt.Errorf("failed to generate SSH key pair: %v", err)
					}

					// Upload the new deploy key
					deployKeyTitle := fmt.Sprintf("terraform-deploy-key-%d", time.Now().Unix())
					err = r.uploadDeployKey(ctx, owner, repoName, publicKey, deployKeyTitle, false)
					if err != nil {
						return fmt.Errorf("failed to upload deploy key: %v", err)
					}

					updatedFileContents = r.addSSHKeysToFileContents(fileContents, privateKey, publicKey)
				}
			}
		} else {
			// File doesn't exist yet, generate new SSH keys
			privateKey, publicKey, err := r.generateSSHKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate SSH key pair: %v", err)
			}

			// Upload the new deploy key
			deployKeyTitle := fmt.Sprintf("terraform-deploy-key-%d", time.Now().Unix())
			err = r.uploadDeployKey(ctx, owner, repoName, publicKey, deployKeyTitle, false)
			if err != nil {
				return fmt.Errorf("failed to upload deploy key: %v", err)
			}

			updatedFileContents = r.addSSHKeysToFileContents(fileContents, privateKey, publicKey)
		}
	}

	// Create update options
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String("Update resources.toml via Terraform"),
		Content:   []byte(updatedFileContents),
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
	req.Header.Set("X-Api-Key", r.apiKey)
	req.Header.Set("X-Api-Secret", r.apiSecret)
	
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
		req.Header.Set("X-Api-Key", r.apiKey)
		req.Header.Set("X-Api-Secret", r.apiSecret)
		
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
		req.Header.Set("X-Api-Key", r.apiKey)
		req.Header.Set("X-Api-Secret", r.apiSecret)
		
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

// generateSSHKeyPair generates an ed25519 SSH key pair and returns (privateKey, publicKey, error)
func (r *komodoResource) generateSSHKeyPair() (string, string, error) {
	// Generate ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ed25519 key pair: %v", err)
	}

	// Convert to SSH format
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert public key to SSH format: %v", err)
	}

	// Format public key
	publicKeyString := string(ssh.MarshalAuthorizedKey(sshPublicKey))

	// Format private key in OpenSSH format
	privateKeyBytes, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %v", err)
	}

	// Convert to PEM format
	privateKeyPEM := pem.EncodeToMemory(privateKeyBytes)
	privateKeyString := string(privateKeyPEM)

	return privateKeyString, strings.TrimSpace(publicKeyString), nil
}

// uploadDeployKey uploads the public key as a deploy key to the GitHub repository
func (r *komodoResource) uploadDeployKey(ctx context.Context, owner, repoName, publicKey, title string, readOnly bool) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create deploy key
	deployKey := &github.Key{
		Title:    github.String(title),
		Key:      github.String(publicKey),
		ReadOnly: github.Bool(readOnly),
	}

	_, _, err := client.Repositories.CreateKey(ctx, owner, repoName, deployKey)
	if err != nil {
		return fmt.Errorf("failed to upload deploy key: %v", err)
	}

	return nil
}

// addSSHKeysToFileContents adds SSH keys to the environment section of the file contents
func (r *komodoResource) addSSHKeysToFileContents(fileContents, privateKey, publicKey string) string {
	// Find the environment section and add SSH keys before the closing """
	envPattern := `environment = """([^"]*?)"""`
	re := regexp.MustCompile(envPattern)

	return re.ReplaceAllStringFunc(fileContents, func(match string) string {
		// Extract the current environment content
		envMatch := re.FindStringSubmatch(match)
		if len(envMatch) < 2 {
			return match
		}

		currentEnv := envMatch[1]

		// Add SSH keys to the environment
		sshKeysSection := fmt.Sprintf("\nSSH_PRIVATE_KEY=%s\nSSH_PUBLIC_KEY=%s",
			strings.ReplaceAll(privateKey, "\n", "\\n"),
			publicKey)

		// Return the updated environment section
		return fmt.Sprintf(`environment = """%s%s"""`, currentEnv, sshKeysSection)
	})
}
