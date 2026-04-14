package webhook

import (
	"context"
	"fmt"
	"strconv"

	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

type webhookResource struct {
	client *lib.APIClient
}

type webhookResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Url         types.String `tfsdk:"url"`
	Description types.String `tfsdk:"description"`
	Events      types.List   `tfsdk:"events"`
	Type        types.String `tfsdk:"type"`
	Domain      types.String `tfsdk:"domain"`
	Batched     types.Bool   `tfsdk:"batched"`
}

func NewResource() resource.Resource {
	return &webhookResource{}
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a webhook in Brevo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Description: "URL of the webhook.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the webhook.",
				Optional:    true,
			},
			"events": schema.ListAttribute{
				Description: "Events triggering the webhook.",
				Required:    true,
				ElementType: types.StringType,
			},
			"type": schema.StringAttribute{
				Description: "Type of the webhook (transactional or marketing).",
				Required:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Inbound domain of webhook, required for inbound event type.",
				Optional:    true,
			},
			"batched": schema.BoolAttribute{
				Description: "Whether to send batched webhooks.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateWebhook{
		Url:    plan.Url.ValueString(),
		Events: events,
		Type_:  plan.Type.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		body.Domain = plan.Domain.ValueString()
	}
	if !plan.Batched.IsNull() && !plan.Batched.IsUnknown() {
		body.Batched = plan.Batched.ValueBool()
	}

	result, _, err := r.client.WebhooksApi.CreateWebhook(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.Id)

	// Read back to populate computed fields.
	r.readWebhookInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readWebhookInto(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhookResource) readWebhookInto(ctx context.Context, model *webhookResourceModel, diagnostics *diag.Diagnostics) {
	wh, _, err := r.client.WebhooksApi.GetWebhook(ctx, model.ID.ValueInt64())
	if err != nil {
		diagnostics.AddError("Error reading webhook", err.Error())
		return
	}

	model.Url = types.StringValue(wh.Url)
	model.Type = types.StringValue(wh.Type_)
	model.Batched = types.BoolValue(wh.Batched)

	if wh.Description == "" {
		model.Description = types.StringNull()
	} else {
		model.Description = types.StringValue(wh.Description)
	}

	eventValues := make([]attr.Value, len(wh.Events))
	for i, e := range wh.Events {
		eventValues[i] = types.StringValue(e)
	}
	eventsList, diags := types.ListValue(types.StringType, eventValues)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
	model.Events = eventsList
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.UpdateWebhook{
		Url:    plan.Url.ValueString(),
		Events: events,
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		body.Domain = plan.Domain.ValueString()
	}
	if !plan.Batched.IsNull() && !plan.Batched.IsUnknown() {
		body.Batched = plan.Batched.ValueBool()
	}

	_, err := r.client.WebhooksApi.UpdateWebhook(ctx, state.ID.ValueInt64(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	plan.ID = state.ID

	// Read back to populate computed fields.
	r.readWebhookInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.WebhooksApi.DeleteWebhook(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
		return
	}
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Could not parse %q as int64: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
