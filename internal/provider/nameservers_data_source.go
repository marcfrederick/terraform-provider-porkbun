package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
)

var _ datasource.DataSource = &DomainNameserversDataSource{}

func NewNameserversDataSource() datasource.DataSource {
	return &DomainNameserversDataSource{}
}

// DomainNameserversDataSource defines the data source implementation.
type DomainNameserversDataSource struct {
	client *porkbun.Client
}

// DomainNameserversDataSourceModel describes the data source data model.
type DomainNameserversDataSourceModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

func (d *DomainNameserversDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameservers"
}

func (d *DomainNameserversDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve the current nameservers for a domain registered with Porkbun.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name to retrieve nameserver information for. Must be a domain registered with or managed through Porkbun.",
				Required:            true,
			},
			"nameservers": schema.ListAttribute{
				MarkdownDescription: "A list of name server host names.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *DomainNameserversDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (d *DomainNameserversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainNameserversDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nsResp, err := d.client.Domains.GetNameServers(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Nameservers", err.Error())
		return
	}

	data.Nameservers = listOfStringsToList(nsResp.NS)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
