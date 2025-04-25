package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
)

var (
	_ resource.Resource                = &DomainNameserversResource{}
	_ resource.ResourceWithImportState = &DomainNameserversResource{}
)

func NewDomainNameserversResource() resource.Resource {
	return &DomainNameserversResource{}
}

type DomainNameserversResource struct {
	client *porkbun.Client
}

type DomainNameserversResourceModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

func (r *DomainNameserversResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameservers"
}

func (r *DomainNameserversResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage custom nameservers for domains registered through Porkbun.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name to manage nameservers for. Must be a domain registered with Porkbun.",
				Required:            true,
			},
			"nameservers": schema.ListAttribute{
				MarkdownDescription: "The list of nameservers to set for the domain. ", // Use Porkbun default nameservers by deleting this resource.",
				ElementType:         types.StringType,
				Required:            true,
			},
		},
	}
}

func (r *DomainNameserversResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = ConfigureResource(req, resp)
}

func (r *DomainNameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameservers, err := extractNameservers(data.Nameservers)
	if err != nil {
		resp.Diagnostics.AddError("Error Extracting Nameservers", err.Error())
		return
	}

	if _, err := r.client.Domains.UpdateNameServers(ctx, data.Domain.ValueString(), &nameservers); err != nil {
		resp.Diagnostics.AddError("Error Setting Nameservers", err.Error())
		return
	}

	data.Nameservers = listOfStringsToList(nameservers)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nsResp, err := r.client.Domains.GetNameServers(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Nameservers", err.Error())
		return
	}

	data.Nameservers = listOfStringsToList(nsResp.NS)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameservers, err := extractNameservers(data.Nameservers)
	if err != nil {
		resp.Diagnostics.AddError("Error Extracting Nameservers", err.Error())
		return
	}

	if _, err := r.client.Domains.UpdateNameServers(ctx, data.Domain.ValueString(), &nameservers); err != nil {
		resp.Diagnostics.AddError("Error Updating Nameservers", err.Error())
		return
	}

	data.Nameservers = listOfStringsToList(nameservers)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Domains.UpdateNameServers(ctx, data.Domain.ValueString(), &porkbun.NameServers{}); err != nil {
		resp.Diagnostics.AddError("Error Deleting Nameservers", err.Error())
	}
}

func (r *DomainNameserversResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}

// extractNameservers converts a types.List to a porkbun.NameServers slice.
func extractNameservers(list types.List) (porkbun.NameServers, error) {
	elements := list.Elements()
	result := make(porkbun.NameServers, 0, len(elements))
	for _, elem := range elements {
		stringElem, ok := elem.(types.String)
		if !ok {
			return nil, fmt.Errorf("error converting element to string: %v", elem)
		}
		result = append(result, stringElem.ValueString())
	}
	return result, nil
}

// listOfStringsToList converts a slice of strings to a types.List.
func listOfStringsToList(values []string) types.List {
	attrVals := make([]attr.Value, 0, len(values))
	for _, v := range values {
		attrVals = append(attrVals, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, attrVals)
}
