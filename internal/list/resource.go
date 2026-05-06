package list

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bbieniek/terraform-provider-brevo/internal/common"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &listResource{}
	_ resource.ResourceWithConfigure   = &listResource{}
	_ resource.ResourceWithImportState = &listResource{}
)

type listResource struct {
	client *lib.APIClient
}

type listResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	FolderID          types.Int64  `tfsdk:"folder_id"`
	TotalBlacklisted  types.Int64  `tfsdk:"total_blacklisted"`
	TotalSubscribers  types.Int64  `tfsdk:"total_subscribers"`
	UniqueSubscribers types.Int64  `tfsdk:"unique_subscribers"`
	CreatedAt         types.String `tfsdk:"created_at"`
	DynamicList       types.Bool   `tfsdk:"dynamic_list"`
}

func NewResource() resource.Resource {
	return &listResource{}
}

func (r *listResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (r *listResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a contact list in Brevo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the list.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the list.",
				Required:    true,
			},
			"folder_id": schema.Int64Attribute{
				Description: "ID of the folder containing the list.",
				Required:    true,
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

func (r *listResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*common.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T", req.ProviderData))
		return
	}
	r.client = data.Client
}

func (r *listResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan listResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateList{
		Name:     plan.Name.ValueString(),
		FolderId: plan.FolderID.ValueInt64(),
	}

	result, _, err := r.client.ContactsApi.CreateList(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating list", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.Id)

	r.readListInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *listResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state listResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readListInto(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *listResource) readListInto(ctx context.Context, model *listResourceModel, diagnostics *diag.Diagnostics) {
	l, _, err := r.client.ContactsApi.GetList(ctx, model.ID.ValueInt64(), nil)
	if err != nil {
		diagnostics.AddError("Error reading list", err.Error())
		return
	}

	model.Name = types.StringValue(l.Name)
	model.FolderID = types.Int64Value(l.FolderId)
	model.TotalBlacklisted = types.Int64Value(l.TotalBlacklisted)
	model.TotalSubscribers = types.Int64Value(l.TotalSubscribers)
	model.UniqueSubscribers = types.Int64Value(l.UniqueSubscribers)
	model.CreatedAt = types.StringValue(l.CreatedAt)
	model.DynamicList = types.BoolValue(l.DynamicList)
}

func (r *listResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan listResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state listResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.UpdateList{
		Name:     plan.Name.ValueString(),
		FolderId: plan.FolderID.ValueInt64(),
	}

	_, err := r.client.ContactsApi.UpdateList(ctx, state.ID.ValueInt64(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating list", err.Error())
		return
	}

	plan.ID = state.ID

	r.readListInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *listResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state listResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ContactsApi.DeleteList(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting list", err.Error())
		return
	}
}

func (r *listResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Could not parse %q as int64: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
