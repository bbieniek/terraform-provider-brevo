package sender

import (
	"context"
	"fmt"

	"github.com/bbieniek/terraform-provider-brevo/internal/common"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &senderDataSource{}
	_ datasource.DataSourceWithConfigure = &senderDataSource{}
)

type senderDataSource struct {
	client *lib.APIClient
}

type senderDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func NewDataSource() datasource.DataSource {
	return &senderDataSource{}
}

func (d *senderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sender"
}

func (d *senderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a sender in Brevo by email.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the sender.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "From Name of the sender.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "From Email of the sender.",
				Required:    true,
			},
		},
	}
}

func (d *senderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*common.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T", req.ProviderData))
		return
	}
	d.client = data.Client
}

func (d *senderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config senderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := config.Email.ValueString()

	senders, _, err := d.client.SendersApi.GetSenders(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading senders", err.Error())
		return
	}

	for _, s := range senders.Senders {
		if s.Email == email {
			config.ID = types.Int64Value(s.Id)
			config.Name = types.StringValue(s.Name)
			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError("Sender not found",
		fmt.Sprintf("Sender with email %q not found in Brevo account", email))
}
