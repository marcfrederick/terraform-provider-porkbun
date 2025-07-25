package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/util"
	"github.com/marcfrederick/terraform-provider-porkbun/internal/validator/enumvalidator"
)

var (
	_ resource.Resource                = &URLForwardResource{}
	_ resource.ResourceWithImportState = &URLForwardResource{}
)

func NewURLForwardResource() resource.Resource {
	return &URLForwardResource{}
}

type URLForwardResource struct {
	client *porkbun.Client
}

type URLForwardResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Domain      types.String `tfsdk:"domain"`
	Subdomain   types.String `tfsdk:"subdomain"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
	IncludePath types.Bool   `tfsdk:"include_path"`
	Wildcard    types.Bool   `tfsdk:"wildcard"`
}

func (r *URLForwardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_url_forward"
}

func (r *URLForwardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage URL forwarding rules for domains registered through Porkbun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the URL forward. Automatically generated by Porkbun.",
				Computed:            true,
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain name for which to configure URL forwarding (e.g., example.com).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A subdomain that you would like to add URL forwarding for. Leave this blank to forward the root domain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Where you'd like to forward the domain to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of forward (temporary, permanent).",
				Validators: []validator.String{
					enumvalidator.Valid(porkbun.Temporary, porkbun.Permanent),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"include_path": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether or not to include the URI path in the redirection.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"wildcard": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Also forward all subdomains of the domain.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *URLForwardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (r *URLForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data URLForwardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.createURLForward(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating URL Forward", err.Error())
		return
	}
	data.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *URLForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data URLForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forward, err := r.readURLForward(ctx, data.Domain.ValueString(), data.Subdomain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading URL Forward", err.Error())
		return
	}

	data.ID = types.StringValue(forward.Id)
	data.Subdomain = types.StringValue(forward.Subdomain)
	data.Location = types.StringValue(forward.Location)
	data.Type = types.StringValue(string(forward.Type))
	data.IncludePath = util.BoolValue(forward.IncludePath, &resp.Diagnostics)
	data.Wildcard = util.BoolValue(forward.Wildcard, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *URLForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data URLForwardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Domains.DeleteDomainUrlForward(ctx, data.Domain.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting URL Forward for Update", err.Error())
		return
	}

	id, err := r.createURLForward(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating URL Forward", err.Error())
		return
	}
	data.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *URLForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data URLForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Domains.DeleteDomainUrlForward(ctx, data.Domain.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting URL Forward", err.Error())
	}
}

func (r *URLForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.SplitN(req.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <domain>:<forward_id>")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &URLForwardResourceModel{
		ID:     types.StringValue(idParts[1]),
		Domain: types.StringValue(idParts[0]),
	})...)
}

// createURLForward creates a new URL forward for the specified domain and subdomain.
func (r *URLForwardResource) createURLForward(ctx context.Context, data *URLForwardResourceModel) (string, error) {
	opts := porkbun.UrlForward{
		Subdomain:   data.Subdomain.ValueString(),
		Location:    data.Location.ValueString(),
		Type:        porkbun.ForwardType(data.Type.ValueString()),
		IncludePath: encodeBool(data.IncludePath.ValueBool()),
		Wildcard:    encodeBool(data.Wildcard.ValueBool()),
	}

	if _, err := r.client.Domains.AddDomainUrlForward(ctx, data.Domain.ValueString(), &opts); err != nil {
		return "", err
	}

	forward, err := r.readURLForward(ctx, data.Domain.ValueString(), data.Subdomain.ValueString())
	if err != nil {
		return "", err
	}

	return forward.Id, nil
}

// readURLForward retrieves the URL forward for the specified domain and subdomain.
func (r *URLForwardResource) readURLForward(ctx context.Context, domain, subdomain string) (*porkbun.UrlForwardData, error) {
	resp, err := r.client.Domains.GetDomainURLForwarding(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed fetching URL forwards for domain %s: %w", domain, err)
	}

	for _, forward := range resp.Forwards {
		if forward.Subdomain == subdomain {
			return &forward, nil
		}
	}

	return nil, fmt.Errorf("URL forward not found for domain %s and subdomain %s", domain, subdomain)
}

// encodeBool converts a boolean value to a string representation.
func encodeBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
