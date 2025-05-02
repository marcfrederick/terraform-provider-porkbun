package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tuzzmaniandevil/porkbun-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ ephemeral.EphemeralResourceWithConfigure = &SSLEphemeralResource{}

func NewSSLEphemeralResource() ephemeral.EphemeralResource {
	return &SSLEphemeralResource{}
}

// SSLEphemeralResource defines the ephemeral resource implementation.
type SSLEphemeralResource struct {
	client *porkbun.Client
}

// SSLEphemeralResourceModel describes the ephemeral resource data model.
type SSLEphemeralResourceModel struct {
	Domain           types.String `tfsdk:"domain"`
	CertificateChain types.String `tfsdk:"certificate_chain"`
	PrivateKey       types.String `tfsdk:"private_key"`
	PublicKey        types.String `tfsdk:"public_key"`
}

func (r *SSLEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl"
}

func (r *SSLEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve SSL certificate information for domains using Porkbun's free SSL certificates.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name for which to retrieve SSL certificate information.",
				Required:            true,
			},
			"certificate_chain": schema.StringAttribute{
				MarkdownDescription: "The certificate chain for the SSL certificate, which includes intermediate certificates needed for validation.",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The private key for the SSL certificate.",
				Computed:            true,
				Sensitive:           true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The public key (certificate) for the SSL certificate, containing the domain's identity and public key.",
				Computed:            true,
			},
		},
	}
}

func (r *SSLEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	r.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (r *SSLEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data SSLEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sslResp, err := r.client.Ssl.Retrieve(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to retrieve SSL data: %s", err))
		return
	}

	data.CertificateChain = types.StringValue(sslResp.Certificatechain)
	data.PrivateKey = types.StringValue(sslResp.Privatekey)
	data.PublicKey = types.StringValue(sslResp.Publickey)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
