package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
)

const (
	providerType = "porkbun"

	// Provider argument names.
	argAPIKey       = "api_key"
	argSecretAPIKey = "secret_api_key"
	argIPv4Only     = "ipv4_only"
	argMaxRetries   = "max_retries"

	// Default values for provider arguments.
	argIPV4OnlyDefault   = false
	argMaxRetriesDefault = 3
)

// Ensure PorkbunProvider satisfies various provider interfaces.
var (
	_ provider.Provider                       = &PorkbunProvider{}
	_ provider.ProviderWithFunctions          = &PorkbunProvider{}
	_ provider.ProviderWithEphemeralResources = &PorkbunProvider{}
)

// PorkbunProvider defines the provider implementation.
type PorkbunProvider struct {
	version string
}

// PorkbunProviderModel describes the provider data model.
type PorkbunProviderModel struct {
	APIKey       types.String `tfsdk:"api_key"`
	SecretAPIKey types.String `tfsdk:"secret_api_key"`
	IPv4Only     types.Bool   `tfsdk:"ipv4_only"`
	MaxRetries   types.Int64  `tfsdk:"max_retries"`
}

func (p *PorkbunProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerType
	resp.Version = p.version
}

func (p *PorkbunProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing domains, DNS records, URL forwarding, and nameserver configurations for domains registered with Porkbun.",
		Attributes: map[string]schema.Attribute{
			argAPIKey: schema.StringAttribute{
				MarkdownDescription: "API key for authentication. Can also be set using the `PORKBUN_API_KEY` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			argSecretAPIKey: schema.StringAttribute{
				MarkdownDescription: "Secret API key for authentication. Can also be set using the `PORKBUN_SECRET_API_KEY` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			argIPv4Only: schema.BoolAttribute{
				MarkdownDescription: "Use IPv4 only for API requests. Defaults to false.",
				Optional:            true,
			},
			argMaxRetries: schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("Maximum number of retries for API requests. Defaults to %d.", argMaxRetriesDefault),
				Optional:            true,
			},
		},
	}
}

func (p *PorkbunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PorkbunProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.validateUnknownAttribute(resp, data.APIKey, path.Root(argAPIKey), "Porkbun API Key")
	p.validateUnknownAttribute(resp, data.SecretAPIKey, path.Root(argSecretAPIKey), "Porkbun Secret API Key")
	p.validateUnknownAttribute(resp, data.IPv4Only, path.Root(argIPv4Only), "Porkbun IPv4 Flag")
	p.validateUnknownAttribute(resp, data.MaxRetries, path.Root(argMaxRetries), "Max Retries Count")
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("PORKBUN_API_KEY")
	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}
	p.validateMissingAttribute(resp, apiKey, "Porkbun API Key", path.Root(argAPIKey))

	secretAPIKey := os.Getenv("PORKBUN_SECRET_API_KEY")
	if !data.SecretAPIKey.IsNull() {
		secretAPIKey = data.SecretAPIKey.ValueString()
	}
	p.validateMissingAttribute(resp, secretAPIKey, "Porkbun Secret API Key", path.Root(argSecretAPIKey))

	ipv4Only := argIPV4OnlyDefault
	if !data.IPv4Only.IsNull() {
		ipv4Only = data.IPv4Only.ValueBool()
	}

	maxRetries := argMaxRetriesDefault
	if !data.MaxRetries.IsNull() {
		maxRetries = int(data.MaxRetries.ValueInt64())
		if maxRetries < 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root(argMaxRetries),
				"Invalid Max Retries Count",
				"The maximum number of retries for API requests must be a non-negative integer.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	httpClient := p.newRetryableHttpClient(maxRetries)
	client := porkbun.NewClient(&porkbun.Options{
		ApiKey:       apiKey,
		SecretApiKey: secretAPIKey,
		IPv4Only:     ipv4Only,
		HttpClient:   &httpClient,
	})

	resp.ResourceData = client
	resp.EphemeralResourceData = client
	resp.DataSourceData = client
}

func (p *PorkbunProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
		NewDNSSECRecordResource,
		NewDomainNameserversResource,
		NewURLForwardResource,
	}
}

func (p *PorkbunProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewSSLEphemeralResource,
	}
}

func (p *PorkbunProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
		NewDomainsDataSource,
		NewNameserversDataSource,
		NewSSLDataSource,
	}
}

func (p *PorkbunProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

// validateUnknownAttribute checks if the attribute is unknown and adds an error to the response.
func (p *PorkbunProvider) validateUnknownAttribute(resp *provider.ConfigureResponse, attr attr.Value, attrPath path.Path, attrName string) {
	if attr.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			attrPath,
			fmt.Sprintf("Unknown %s", attrName),
			fmt.Sprintf("The provider cannot create the Porkbun API client as there is an unknown configuration value for the %s. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the corresponding environment variable.", attrName),
		)
	}
}

// validateMissingAttribute checks if the attribute is missing and adds an error to the response.
func (p *PorkbunProvider) validateMissingAttribute(resp *provider.ConfigureResponse, value, attrName string, attrPath path.Path) {
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			attrPath,
			fmt.Sprintf("Missing %s", attrName),
			fmt.Sprintf("The provider cannot create the Porkbun API client as there is a missing or empty value for the %s. "+
				"Set the value in the configuration or use the corresponding environment variable. "+
				"If either is already set, ensure the value is not empty.", attrName),
		)
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PorkbunProvider{
			version: version,
		}
	}
}

// getPorkbunClient retrieves the Porkbun client from the provider data.
//
// It returns nil if the provider data is nil or if the type assertion fails.
// In case of an error, it adds an error to the diagnostics.
func getPorkbunClient(providerData any, diagnostics diag.Diagnostics) *porkbun.Client {
	if providerData == nil {
		return nil
	}

	client, ok := providerData.(*porkbun.Client)
	if !ok {
		diagnostics.AddError(
			"Unexpected ProviderData",
			fmt.Sprintf("Expected *porkbun.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}

	return client
}

// newRetryableHttpClient creates a porkbun.HTTPClient with retry capabilities.
func (p *PorkbunProvider) newRetryableHttpClient(maxRetries int) porkbun.HTTPClient {
	retryableHttpClient := retryablehttp.NewClient()
	retryableHttpClient.RetryMax = maxRetries
	return retryableHttpClient.StandardClient()
}
