package webhook

import (
	"context"
	"fmt"

	"github.com/bbieniek/terraform-provider-brevo/internal/common"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &webhookDataSource{}
	_ datasource.DataSourceWithConfigure = &webhookDataSource{}
)

type webhookDataSource struct {
	client *lib.APIClient
}

type webhookDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Url         types.String `tfsdk:"url"`
	Description types.String `tfsdk:"description"`
	Events      types.List   `tfsdk:"events"`
	Type        types.String `tfsdk:"type"`
	Batched     types.Bool   `tfsdk:"batched"`
}

func NewDataSource() datasource.DataSource {
	return &webhookDataSource{}
}

func (d *webhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (d *webhookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a webhook in Brevo by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the webhook.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL of the webhook.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the webhook.",
				Computed:    true,
			},
			"events": schema.ListAttribute{
				Description: "Events triggering the webhook.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"type": schema.StringAttribute{
				Description: "Type of the webhook.",
				Computed:    true,
			},
			"batched": schema.BoolAttribute{
				Description: "Whether batched webhooks are enabled.",
				Computed:    true,
			},
		},
	}
}

func (d *webhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *webhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webhookDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wh, _, err := d.client.WebhooksApi.GetWebhook(ctx, config.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading webhook", err.Error())
		return
	}

	config.Url = types.StringValue(wh.Url)
	config.Type = types.StringValue(wh.Type_)
	config.Batched = types.BoolValue(wh.Batched)

	if wh.Description == "" {
		config.Description = types.StringNull()
	} else {
		config.Description = types.StringValue(wh.Description)
	}

	eventValues := make([]attr.Value, len(wh.Events))
	for i, e := range wh.Events {
		eventValues[i] = types.StringValue(e)
	}
	eventsList, diags := types.ListValue(types.StringType, eventValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Events = eventsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
