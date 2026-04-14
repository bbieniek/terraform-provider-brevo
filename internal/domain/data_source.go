package domain

import (
	"context"
	"fmt"

	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &domainDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDataSource{}
)

type domainDataSource struct {
	client *lib.APIClient
}

type domainDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DkimRecord1 types.String `tfsdk:"dkim_record_1"`
	DkimRecord2 types.String `tfsdk:"dkim_record_2"`
	BrevoCode   types.String `tfsdk:"brevo_code"`
	Verified    types.Bool   `tfsdk:"verified"`
}

func NewDataSource() datasource.DataSource {
	return &domainDataSource{}
}

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a sender domain in Brevo by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the domain.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Domain name to look up.",
				Required:    true,
			},
			"dkim_record_1": schema.StringAttribute{
				Description: "DKIM CNAME record value (first key).",
				Computed:    true,
			},
			"dkim_record_2": schema.StringAttribute{
				Description: "DKIM CNAME record value (second key).",
				Computed:    true,
			},
			"brevo_code": schema.StringAttribute{
				Description: "Brevo verification code TXT record value.",
				Computed:    true,
			},
			"verified": schema.BoolAttribute{
				Description: "Whether the domain has been verified.",
				Computed:    true,
			},
		},
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*lib.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *lib.APIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config domainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := config.Name.ValueString()

	// Get all domains to find the ID by name.
	domains, _, err := d.client.DomainsApi.GetDomains(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading domains", err.Error())
		return
	}

	var found bool
	for _, dom := range domains.Domains {
		if dom.DomainName == domainName {
			config.ID = types.Int64Value(dom.Id)
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError("Domain not found",
			fmt.Sprintf("Domain %q not found in Brevo account", domainName))
		return
	}

	// Get DNS configuration.
	domainConfig, _, err := d.client.DomainsApi.GetDomainConfiguration(ctx, domainName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading domain configuration", err.Error())
		return
	}

	config.Verified = types.BoolValue(domainConfig.Verified)

	if domainConfig.DnsRecords != nil {
		if domainConfig.DnsRecords.DkimRecord != nil {
			config.DkimRecord1 = types.StringValue(domainConfig.DnsRecords.DkimRecord.Value)
		}
		if domainConfig.DnsRecords.BrevoCode != nil {
			config.BrevoCode = types.StringValue(domainConfig.DnsRecords.BrevoCode.Value)
		}
	}

	if config.DkimRecord2.IsNull() || config.DkimRecord2.IsUnknown() {
		config.DkimRecord2 = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
