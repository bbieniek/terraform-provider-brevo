package sender

import (
	"context"
	"fmt"
	"strconv"

	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/antihax/optional"
)

var (
	_ resource.Resource                = &senderResource{}
	_ resource.ResourceWithConfigure   = &senderResource{}
	_ resource.ResourceWithImportState = &senderResource{}
)

type senderResource struct {
	client *lib.APIClient
}

type senderResourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func NewResource() resource.Resource {
	return &senderResource{}
}

func (r *senderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sender"
}

func (r *senderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a sender in Brevo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the sender.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "From Name to use for the sender.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "From Email to use for the sender.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *senderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*lib.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *lib.APIClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *senderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan senderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateSender{
		Name:  plan.Name.ValueString(),
		Email: plan.Email.ValueString(),
	}
	opts := &lib.CreateSenderOpts{Sender: optional.NewInterface(body)}
	result, _, err := r.client.SendersApi.CreateSender(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating sender", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *senderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state senderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	senders, _, err := r.client.SendersApi.GetSenders(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading senders", err.Error())
		return
	}

	senderID := state.ID.ValueInt64()
	var found bool
	for _, s := range senders.Senders {
		if s.Id == senderID {
			state.Name = types.StringValue(s.Name)
			state.Email = types.StringValue(s.Email)
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *senderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan senderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state senderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.UpdateSender{
		Name: plan.Name.ValueString(),
	}
	opts := &lib.UpdateSenderOpts{Sender: optional.NewInterface(body)}
	_, err := r.client.SendersApi.UpdateSender(ctx, state.ID.ValueInt64(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating sender", err.Error())
		return
	}

	plan.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *senderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state senderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.SendersApi.DeleteSender(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting sender", err.Error())
		return
	}
}

func (r *senderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Could not parse %q as int64: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
