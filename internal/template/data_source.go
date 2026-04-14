package template

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
	_ datasource.DataSource              = &templateDataSource{}
	_ datasource.DataSourceWithConfigure = &templateDataSource{}
)

type templateDataSource struct {
	client *lib.APIClient
}

type templateDataSourceModel struct {
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

func NewDataSource() datasource.DataSource {
	return &templateDataSource{}
}

func (d *templateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (d *templateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an email template in Brevo by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the template.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the template.",
				Computed:    true,
			},
			"subject": schema.StringAttribute{
				Description: "Subject of the template.",
				Computed:    true,
			},
			"html_content": schema.StringAttribute{
				Description: "HTML content of the template.",
				Computed:    true,
			},
			"sender_name": schema.StringAttribute{
				Description: "Name of the sender.",
				Computed:    true,
			},
			"sender_email": schema.StringAttribute{
				Description: "Email of the sender.",
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the template is active.",
				Computed:    true,
			},
			"reply_to": schema.StringAttribute{
				Description: "Email address for replies.",
				Computed:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag of the template.",
				Computed:    true,
			},
		},
	}
}

func (d *templateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config templateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, _, err := d.client.TransactionalEmailsApi.GetSmtpTemplate(ctx, config.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading email template", err.Error())
		return
	}

	config.Name = types.StringValue(tmpl.Name)
	config.Subject = types.StringValue(tmpl.Subject)
	config.HtmlContent = types.StringValue(tmpl.HtmlContent)
	config.IsActive = types.BoolValue(tmpl.IsActive)

	if tmpl.Sender != nil {
		config.SenderName = types.StringValue(tmpl.Sender.Name)
		config.SenderEmail = types.StringValue(tmpl.Sender.Email)
	}

	if tmpl.ReplyTo == "" || tmpl.ReplyTo == "[DEFAULT_REPLY_TO]" {
		config.ReplyTo = types.StringNull()
	} else {
		config.ReplyTo = types.StringValue(tmpl.ReplyTo)
	}

	if tmpl.Tag == "" {
		config.Tag = types.StringNull()
	} else {
		config.Tag = types.StringValue(tmpl.Tag)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
