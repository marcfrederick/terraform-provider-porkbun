package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"
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

	var domainValues []attr.Value
	start := 0

	for {
		opts := porkbun.DomainListOptions{
			Start:         porkbun.String(strconv.Itoa(start)),
			IncludeLabels: porkbun.String("yes"),
		}

		listDomainsResp, err := d.client.Domains.ListDomains(ctx, &opts)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Domains", err.Error())
			return
		}

		for _, domain := range listDomainsResp.Domains {
			domainValues = append(domainValues, convertDomainToObjectValue(domain))
		}

		if len(listDomainsResp.Domains) < listDomainsBatchSize {
			break
		}
		start += listDomainsBatchSize
	}

	data.Domains = types.ListValueMust(
		types.ObjectType{AttrTypes: domainObjectAttrs},
		domainValues,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// convertDomainToObjectValue converts a porkbun.Domain to an attr.Value.
func convertDomainToObjectValue(domain porkbun.Domain) attr.Value {
	return types.ObjectValueMust(
		domainObjectAttrs,
		map[string]attr.Value{
			"domain":        types.StringValue(domain.Domain),
			"status":        types.StringValue(domain.Status),
			"tld":           types.StringValue(domain.TLD),
			"security_lock": types.BoolValue(bool(domain.SecurityLock)),
			"whois_privacy": types.BoolValue(bool(domain.WhoisPrivacy)),
			"auto_renew":    types.BoolValue(bool(domain.AutoRenew)),
			"not_local":     types.BoolValue(bool(domain.NotLocal)),
			"labels":        convertDomainLabelsToList(domain.Labels),
		},
	)
}
