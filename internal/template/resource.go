package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bbieniek/terraform-provider-brevo/internal/common"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &templateResource{}
	_ resource.ResourceWithConfigure   = &templateResource{}
	_ resource.ResourceWithImportState = &templateResource{}
)

type templateResource struct {
	client *lib.APIClient
	apiKey string
}

type templateResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Subject     types.String `tfsdk:"subject"`
	HtmlContent types.String `tfsdk:"html_content"`
	SenderName  types.String `tfsdk:"sender_name"`
	SenderEmail types.String `tfsdk:"sender_email"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	ReplyTo     types.String `tfsdk:"reply_to"`
	Tag         types.String `tfsdk:"tag"`
}

func NewResource() resource.Resource {
	return &templateResource{}
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an email template in Brevo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the template.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the template.",
				Required:    true,
			},
			"subject": schema.StringAttribute{
				Description: "Subject of the template.",
				Required:    true,
			},
			"html_content": schema.StringAttribute{
				Description: "HTML content of the template.",
				Required:    true,
			},
			"sender_name": schema.StringAttribute{
				Description: "Name of the sender.",
				Required:    true,
			},
			"sender_email": schema.StringAttribute{
				Description: "Email of the sender.",
				Required:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the template is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"reply_to": schema.StringAttribute{
				Description: "Email address for replies.",
				Optional:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag of the template.",
				Optional:    true,
			},
		},
	}
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.apiKey = data.APIKey
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateSmtpTemplate{
		TemplateName: plan.Name.ValueString(),
		Subject:      plan.Subject.ValueString(),
		HtmlContent:  plan.HtmlContent.ValueString(),
		Sender: &lib.CreateSmtpTemplateSender{
			Name:  plan.SenderName.ValueString(),
			Email: plan.SenderEmail.ValueString(),
		},
		IsActive: plan.IsActive.ValueBool(),
	}

	if !plan.ReplyTo.IsNull() && !plan.ReplyTo.IsUnknown() {
		body.ReplyTo = plan.ReplyTo.ValueString()
	}
	if !plan.Tag.IsNull() && !plan.Tag.IsUnknown() {
		body.Tag = plan.Tag.ValueString()
	}

	result, _, err := r.client.TransactionalEmailsApi.CreateSmtpTemplate(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating email template", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, _, err := r.client.TransactionalEmailsApi.GetSmtpTemplate(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading email template", err.Error())
		return
	}

	state.Name = types.StringValue(tmpl.Name)
	state.Subject = types.StringValue(tmpl.Subject)
	state.HtmlContent = types.StringValue(tmpl.HtmlContent)
	state.IsActive = types.BoolValue(tmpl.IsActive)

	if tmpl.Sender != nil {
		state.SenderName = types.StringValue(tmpl.Sender.Name)
		state.SenderEmail = types.StringValue(tmpl.Sender.Email)
	}

	if tmpl.ReplyTo == "" || tmpl.ReplyTo == "[DEFAULT_REPLY_TO]" {
		state.ReplyTo = types.StringNull()
	} else {
		state.ReplyTo = types.StringValue(tmpl.ReplyTo)
	}

	if tmpl.Tag == "" {
		state.Tag = types.StringNull()
	} else {
		state.Tag = types.StringValue(tmpl.Tag)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use raw HTTP because the SDK uses omitempty on bool fields,
	// which means IsActive=false is never sent (it's the zero value).
	updateMap := map[string]interface{}{
		"templateName": plan.Name.ValueString(),
		"subject":      plan.Subject.ValueString(),
		"htmlContent":  plan.HtmlContent.ValueString(),
		"sender": map[string]string{
			"name":  plan.SenderName.ValueString(),
			"email": plan.SenderEmail.ValueString(),
		},
		"isActive": plan.IsActive.ValueBool(),
	}
	if !plan.ReplyTo.IsNull() && !plan.ReplyTo.IsUnknown() {
		updateMap["replyTo"] = plan.ReplyTo.ValueString()
	}
	if !plan.Tag.IsNull() && !plan.Tag.IsUnknown() {
		updateMap["tag"] = plan.Tag.ValueString()
	}

	jsonBody, err := json.Marshal(updateMap)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding update request", err.Error())
		return
	}

	url := fmt.Sprintf("https://api.brevo.com/v3/smtp/templates/%d", state.ID.ValueInt64())
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError("Error building update request", err.Error())
		return
	}
	httpReq.Header.Set("api-key", r.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating email template", err.Error())
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("Error updating email template",
			fmt.Sprintf("API returned %d: %s", httpResp.StatusCode, string(respBody)))
		return
	}

	plan.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Brevo requires templates to be deactivated before deletion.
	// Use raw HTTP because the SDK sends all zero-value fields which the API rejects.
	url := fmt.Sprintf("https://api.brevo.com/v3/smtp/templates/%d", state.ID.ValueInt64())
	body := bytes.NewBufferString(`{"isActive":false}`)
	httpReq, err := http.NewRequest("PUT", url, body)
	if err != nil {
		resp.Diagnostics.AddError("Error building deactivation request", err.Error())
		return
	}
	httpReq.Header.Set("api-key", r.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error deactivating email template", err.Error())
		return
	}
	_ = httpResp.Body.Close()

	_, err = r.client.TransactionalEmailsApi.DeleteSmtpTemplate(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting email template", err.Error())
		return
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Could not parse %q as int64: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
