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

