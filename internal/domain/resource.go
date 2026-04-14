package domain

import (
	"context"
	"fmt"

	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &domainResource{}
	_ resource.ResourceWithConfigure   = &domainResource{}
	_ resource.ResourceWithImportState = &domainResource{}
)

type domainResource struct {
	client *lib.APIClient
}

type domainResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DkimRecord1 types.String `tfsdk:"dkim_record_1"`
	DkimRecord2 types.String `tfsdk:"dkim_record_2"`
	BrevoCode   types.String `tfsdk:"brevo_code"`
	Verified    types.Bool   `tfsdk:"verified"`
}

func NewResource() resource.Resource {
	return &domainResource{}
}

func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a sender domain in Brevo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Domain name to register with Brevo.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateDomain{Name: plan.Name.ValueString()}
	opts := &lib.CreateDomainOpts{DomainName: optional.NewInterface(body)}
	result, _, err := r.client.DomainsApi.CreateDomain(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.Id)

	// Read back the full configuration to get DNS records.
	r.readDomainConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readDomainConfig(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *domainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported",
		"Domain resources are immutable. Change the name to trigger replacement.")
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DomainsApi.DeleteDomain(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting domain", err.Error())
		return
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)

	// Look up the numeric ID from the domains list.
	domains, _, err := r.client.DomainsApi.GetDomains(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing domain", err.Error())
		return
	}
	for _, d := range domains.Domains {
		if d.DomainName == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), d.Id)...)
			return
		}
	}
	resp.Diagnostics.AddError("Domain not found",
		fmt.Sprintf("Domain %q not found in Brevo account", req.ID))
}

func (r *domainResource) readDomainConfig(ctx context.Context, model *domainResourceModel, diagnostics *diag.Diagnostics) {
	config, _, err := r.client.DomainsApi.GetDomainConfiguration(ctx, model.Name.ValueString())
	if err != nil {
		diagnostics.AddError("Error reading domain configuration", err.Error())
		return
	}

	model.Verified = types.BoolValue(config.Verified)

	if config.DnsRecords != nil {
		if config.DnsRecords.DkimRecord != nil {
			model.DkimRecord1 = types.StringValue(config.DnsRecords.DkimRecord.Value)
		}
		if config.DnsRecords.BrevoCode != nil {
			model.BrevoCode = types.StringValue(config.DnsRecords.BrevoCode.Value)
		}
	}

	// The SDK returns a single DKIM record struct; Brevo actually provides two DKIM keys.
	// The second DKIM record may need to be derived from the domain name pattern.
	// If the API only returns one, we populate dkim_record_2 as empty.
	if model.DkimRecord2.IsNull() || model.DkimRecord2.IsUnknown() {
		model.DkimRecord2 = types.StringValue("")
	}
}
