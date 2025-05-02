package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
)

// listDomainsBatchSize defines the batch size for listing domains.
//
// This is used to paginate through the list of domains when retrieving them
// from the API. Porkbun returns a maximum of 1000 domains per request, so we
// set this constant to 1000 to match that limit.
//
// https://porkbun.com/api/json/v3/documentation#Domain%20List%20All
const listDomainsBatchSize = 1000

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DomainDataSource{}

// domainLabelObjectAttrs defines the attributes for the domain label object.
var domainLabelObjectAttrs = map[string]attr.Type{
	"id":    types.StringType,
	"title": types.StringType,
	"color": types.StringType,
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

// DomainDataSource defines the data source implementation.
type DomainDataSource struct {
	client *porkbun.Client
}

// DomainDataSourceModel describes the data source data model.
type DomainDataSourceModel struct {
	Domain       types.String `tfsdk:"domain"`
	Status       types.String `tfsdk:"status"`
	TLD          types.String `tfsdk:"tld"`
	SecurityLock types.Bool   `tfsdk:"security_lock"`
	WhoisPrivacy types.Bool   `tfsdk:"whois_privacy"`
	AutoRenew    types.Bool   `tfsdk:"auto_renew"`
	NotLocal     types.Bool   `tfsdk:"not_local"`
	Labels       types.List   `tfsdk:"labels"`
}

func (d *DomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve information about a specific domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name to retrieve information for. Must be a domain registered with or managed through Porkbun.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the domain (e.g., 'ACTIVE', 'EXPIRED', etc.).",
				Computed:            true,
			},
			"tld": schema.StringAttribute{
				MarkdownDescription: "The top-level domain (TLD) of the domain (e.g., 'com', 'org', 'net').",
				Computed:            true,
			},
			"whois_privacy": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether WHOIS privacy protection is enabled for the domain, which hides personal contact information in public WHOIS records.",
				Computed:            true,
			},
			"security_lock": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the domain transfer lock is enabled, which prevents unauthorized domain transfers to other registrars.",
				Computed:            true,
			},
			"auto_renew": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether automatic renewal is enabled for the domain. When enabled, the domain will be automatically renewed before expiration.",
				Computed:            true,
			},
			"not_local": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the domain is registered elsewhere but using Porkbun's DNS (true) or if it's registered with Porkbun (false).",
				Computed:            true,
			},
			"labels": schema.ListAttribute{
				MarkdownDescription: "A list of labels associated with the domain. Labels are used to categorize and organize domains within your Porkbun account.",
				Computed:            true,
				ElementType:         types.ObjectType{AttrTypes: domainLabelObjectAttrs},
			},
		},
	}
}

func (d *DomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.findDomain(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Domain Not Found", fmt.Sprintf("Unable to find domain %s: %s", data.Domain.ValueString(), err))
		return
	}

	data.Domain = types.StringValue(domain.Domain)
	data.Status = types.StringValue(domain.Status)
	data.TLD = types.StringValue(domain.TLD)
	data.SecurityLock = types.BoolValue(bool(domain.SecurityLock))
	data.WhoisPrivacy = types.BoolValue(bool(domain.WhoisPrivacy))
	data.AutoRenew = types.BoolValue(bool(domain.AutoRenew))
	data.NotLocal = types.BoolValue(bool(domain.NotLocal))
	data.Labels = convertDomainLabelsToList(domain.Labels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// findDomain retrieves the domain information from the Porkbun API.
func (d *DomainDataSource) findDomain(ctx context.Context, domainName string) (*porkbun.Domain, error) {
	start := 0
	for {
		opts := porkbun.DomainListOptions{
			Start:         porkbun.String(strconv.Itoa(start)),
			IncludeLabels: porkbun.String("yes"),
		}

		domainResp, err := d.client.Domains.ListDomains(ctx, &opts)
		if err != nil {
			return nil, fmt.Errorf("error listing domains: %w", err)
		}

		for _, domain := range domainResp.Domains {
			if domain.Domain == domainName {
				return &domain, nil
			}
		}

		if len(domainResp.Domains) < listDomainsBatchSize {
			break
		}

		start += listDomainsBatchSize
	}

	return nil, fmt.Errorf("domain %s not found", domainName)
}

// convertDomainLabelsToList converts a slice of porkbun.Label to a types.List of objects.
func convertDomainLabelsToList(labels []porkbun.Label) types.List {
	result := make([]attr.Value, len(labels))

	for i, label := range labels {
		result[i] = types.ObjectValueMust(
			domainLabelObjectAttrs,
			map[string]attr.Value{
				"id":    types.StringValue(label.ID),
				"title": types.StringValue(label.Title),
				"color": types.StringValue(label.Color),
			},
		)
	}

	return types.ListValueMust(
		types.ObjectType{AttrTypes: domainLabelObjectAttrs},
		result,
	)
}
