package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SSLDataSource{}

func NewSSLDataSource() datasource.DataSource {
	return &SSLDataSource{}
}

// SSLDataSource defines the data source implementation.
type SSLDataSource struct {
	client *porkbun.Client
}

// SSLDataSourceModel describes the data source data model.
type SSLDataSourceModel struct {
	Domain           types.String `tfsdk:"domain"`
	CertificateChain types.String `tfsdk:"certificate_chain"`
	PrivateKey       types.String `tfsdk:"private_key"`
	PublicKey        types.String `tfsdk:"public_key"`
}

func (d *SSLDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl"
}

func (d *SSLDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve SSL certificate information for domains using Porkbun's free SSL certificates.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name for which to retrieve SSL certificate information.",
				Required:            true,
			},
			"certificate_chain": schema.StringAttribute{
				MarkdownDescription: "The complete certificate chain.",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The private key.",
				Computed:            true,
				Sensitive:           true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The public key.",
				Computed:            true,
			},
		},
	}
}

func (d *SSLDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (d *SSLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SSLDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sslResp, err := d.client.Ssl.Retrieve(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to retrieve SSL data: %s", err))
		return
	}

	data.CertificateChain = types.StringValue(sslResp.Certificatechain)
	data.PrivateKey = types.StringValue(sslResp.Privatekey)
	data.PublicKey = types.StringValue(sslResp.Publickey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
