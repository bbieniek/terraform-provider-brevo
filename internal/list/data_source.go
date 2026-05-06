package list

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
	_ datasource.DataSource              = &listDataSource{}
	_ datasource.DataSourceWithConfigure = &listDataSource{}
)

type listDataSource struct {
	client *lib.APIClient
}

type listDataSourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	FolderID          types.Int64  `tfsdk:"folder_id"`
	TotalBlacklisted  types.Int64  `tfsdk:"total_blacklisted"`
	TotalSubscribers  types.Int64  `tfsdk:"total_subscribers"`
	UniqueSubscribers types.Int64  `tfsdk:"unique_subscribers"`
	CreatedAt         types.String `tfsdk:"created_at"`
	DynamicList       types.Bool   `tfsdk:"dynamic_list"`
}

func NewDataSource() datasource.DataSource {
	return &listDataSource{}
}

func (d *listDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (d *listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a contact list in Brevo by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the list.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the list.",
				Computed:    true,
			},
			"folder_id": schema.Int64Attribute{
				Description: "ID of the folder containing the list.",
				Computed:    true,
			},
			"total_blacklisted": schema.Int64Attribute{
				Description: "Number of blacklisted contacts in the list.",
				Computed:    true,
			},
			"total_subscribers": schema.Int64Attribute{
				Description: "Number of contacts in the list.",
				Computed:    true,
			},
			"unique_subscribers": schema.Int64Attribute{
				Description: "Number of unique contacts in the list.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp of the list.",
				Computed:    true,
			},
			"dynamic_list": schema.BoolAttribute{
				Description: "Whether the list is dynamic.",
				Computed:    true,
			},
		},
	}
}

func (d *listDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *listDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config listDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, _, err := d.client.ContactsApi.GetList(ctx, config.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading list", err.Error())
		return
	}

	config.Name = types.StringValue(l.Name)
	config.FolderID = types.Int64Value(l.FolderId)
	config.TotalBlacklisted = types.Int64Value(l.TotalBlacklisted)
	config.TotalSubscribers = types.Int64Value(l.TotalSubscribers)
	config.UniqueSubscribers = types.Int64Value(l.UniqueSubscribers)
	config.CreatedAt = types.StringValue(l.CreatedAt)
	config.DynamicList = types.BoolValue(l.DynamicList)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
