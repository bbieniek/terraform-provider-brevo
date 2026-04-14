package provider

import (
	"context"
	"os"

	"github.com/bbieniek/terraform-provider-brevo/internal/domain"
	"github.com/bbieniek/terraform-provider-brevo/internal/sender"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &brevoProvider{}

type brevoProvider struct {
	version string
}

type brevoProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &brevoProvider{
			version: version,
		}
	}
}

func (p *brevoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "brevo"
	resp.Version = p.version
}

func (p *brevoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Brevo API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Brevo API key. Can also be set with the BREVO_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *brevoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config brevoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("BREVO_API_KEY")
	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider cannot create the Brevo API client because the API key is missing. "+
				"Set the api_key attribute in the provider configuration or the BREVO_API_KEY environment variable.",
		)
		return
	}

	cfg := lib.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	client := lib.NewAPIClient(cfg)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *brevoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		domain.NewResource,
		sender.NewResource,
	}
}

func (p *brevoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		domain.NewDataSource,
		sender.NewDataSource,
	}
}
