package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/util"
)

var _ datasource.DataSource = &DomainsDataSource{}

// domainObjectAttrs defines the attributes for the domain object.
var domainObjectAttrs = map[string]attr.Type{
	"domain":        types.StringType,
	"status":        types.StringType,
	"tld":           types.StringType,
	"security_lock": types.BoolType,
	"whois_privacy": types.BoolType,
	"auto_renew":    types.BoolType,
	"not_local":     types.BoolType,
	"labels": types.ListType{
		ElemType: types.ObjectType{AttrTypes: domainLabelObjectAttrs},
	},
}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

// DomainsDataSource defines the data source implementation.
type DomainsDataSource struct {
	client *porkbun.Client
}

// DomainsDataSourceModel describes the data source data model.
type DomainsDataSourceModel struct {
	Domains       types.List `tfsdk:"domains"`
	IncludeLabels types.Bool `tfsdk:"include_labels"`
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about all domains registered with Porkbun.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListAttribute{
				MarkdownDescription: "A list of domains registered with Porkbun.",
				Computed:            true,
				ElementType: types.ObjectType{
					AttrTypes: domainObjectAttrs,
				},
			},
			"include_labels": schema.BoolAttribute{
				MarkdownDescription: "Whether to include labels in the response. Defaults to false.",
				Optional:            true,
			},
		},
	}
}

func (d *DomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsDataSourceModel

	// Load config into data
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domains, err := listDomains(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Domains", err.Error())
		return
	}

	data.Domains = convertDomainsToList(domains, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// listDomains retrieves the list of domains from the Porkbun API.
func listDomains(ctx context.Context, client *porkbun.Client) ([]porkbun.Domain, error) {
	start := 0
	opts := porkbun.DomainListOptions{
		IncludeLabels: porkbun.String("yes"),
	}

	var result []porkbun.Domain
	for {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}

		opts.Start = porkbun.String(strconv.Itoa(start))

		resp, err := client.Domains.ListDomains(ctx, &opts)
		if err != nil {
			return nil, fmt.Errorf("error listing domains at start=%d: %w", start, err)
		}

		if len(resp.Domains) == 0 {
			break
		}

		result = append(result, resp.Domains...)
		start += len(resp.Domains)
	}

	return result, nil
}

// convertDomainLabelsToList converts a slice of porkbun.Domain to a types.List.
func convertDomainsToList(domains []porkbun.Domain, diagnostics *diag.Diagnostics) types.List {
	return util.MustMapToList(domains, types.ObjectType{AttrTypes: domainObjectAttrs}, func(domain porkbun.Domain) attr.Value {
		return convertDomainToObjectValue(domain, diagnostics)
	})
}

// convertDomainToObjectValue converts a porkbun.Domain to an attr.Value.
func convertDomainToObjectValue(domain porkbun.Domain, diagnostics *diag.Diagnostics) attr.Value {
	return types.ObjectValueMust(
		domainObjectAttrs,
		map[string]attr.Value{
			"domain":        types.StringValue(domain.Domain),
			"status":        types.StringValue(domain.Status),
			"tld":           types.StringValue(domain.TLD),
			"security_lock": util.BoolValue(bool(domain.SecurityLock), diagnostics),
			"whois_privacy": util.BoolValue(bool(domain.WhoisPrivacy), diagnostics),
			"auto_renew":    util.BoolValue(bool(domain.AutoRenew), diagnostics),
			"not_local":     util.BoolValue(bool(domain.NotLocal), diagnostics),
			"labels":        util.MustMapToList(domain.Labels, types.ObjectType{AttrTypes: domainLabelObjectAttrs}, convertDomainLabelToObjectValue),
		},
	)
}
