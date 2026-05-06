package contactattribute

import (
	"context"
	"fmt"
	"strings"

	"github.com/bbieniek/terraform-provider-brevo/internal/common"
	lib "github.com/getbrevo/brevo-go/lib"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &contactAttributeResource{}
	_ resource.ResourceWithConfigure   = &contactAttributeResource{}
	_ resource.ResourceWithImportState = &contactAttributeResource{}
)

type contactAttributeResource struct {
	client *lib.APIClient
}

type contactAttributeResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Category        types.String `tfsdk:"category"`
	Type            types.String `tfsdk:"type"`
	Value           types.String `tfsdk:"value"`
	CalculatedValue types.String `tfsdk:"calculated_value"`
}

func NewResource() resource.Resource {
	return &contactAttributeResource{}
}

func (r *contactAttributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact_attribute"
}

func (r *contactAttributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a contact attribute in Brevo. Use this to declare custom contact fields like FIRMA, FIRSTNAME, etc. Brevo's REST API treats name and category as the identifier; changing either forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form 'category/name'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the attribute (uppercase recommended, e.g. FIRMA).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"category": schema.StringAttribute{
				Description: "Category of the attribute: 'normal', 'transactional', 'category', 'calculated', or 'global'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the attribute. Allowed values depend on category. For 'normal': text, date, float, boolean, multiple-choice. For 'transactional': id. For 'category': category.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value or formula for the attribute. Only used when category is 'calculated' or 'global'.",
				Optional:    true,
			},
			"calculated_value": schema.StringAttribute{
				Description: "Calculated value formula as reported by the API.",
				Computed:    true,
			},
		},
	}
}

func (r *contactAttributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *contactAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan contactAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.CreateAttribute{}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		body.Type_ = plan.Type.ValueString()
	}
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		body.Value = plan.Value.ValueString()
	}

	_, err := r.client.ContactsApi.CreateAttribute(ctx, plan.Category.ValueString(), plan.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact attribute", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Category.ValueString() + "/" + plan.Name.ValueString())

	r.readAttributeInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state contactAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readAttributeInto(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *contactAttributeResource) readAttributeInto(ctx context.Context, model *contactAttributeResourceModel, diagnostics *diag.Diagnostics) {
	attrs, _, err := r.client.ContactsApi.GetAttributes(ctx)
	if err != nil {
		diagnostics.AddError("Error reading contact attributes", err.Error())
		return
	}

	wantName := model.Name.ValueString()
	wantCategory := model.Category.ValueString()

	for _, a := range attrs.Attributes {
		if strings.EqualFold(a.Name, wantName) && strings.EqualFold(a.Category, wantCategory) {
			if a.Type_ == "" {
				model.Type = types.StringNull()
			} else {
				model.Type = types.StringValue(a.Type_)
			}
			if a.CalculatedValue == "" {
				model.CalculatedValue = types.StringNull()
			} else {
				model.CalculatedValue = types.StringValue(a.CalculatedValue)
			}
			return
		}
	}

	diagnostics.AddError("Contact attribute not found",
		fmt.Sprintf("Attribute %q in category %q was not found.", wantName, wantCategory))
}

func (r *contactAttributeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan contactAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state contactAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := lib.UpdateAttribute{}
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		body.Value = plan.Value.ValueString()
	}

	_, err := r.client.ContactsApi.UpdateAttribute(ctx, state.Category.ValueString(), state.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact attribute", err.Error())
		return
	}

	plan.ID = state.ID

	r.readAttributeInto(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state contactAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ContactsApi.DeleteAttribute(ctx, state.Category.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting contact attribute", err.Error())
		return
	}
}

func (r *contactAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected import ID in form 'category/name', got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("category"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
