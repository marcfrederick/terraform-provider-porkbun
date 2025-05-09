package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tuzzmaniandevil/porkbun-go"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/validator/enumvalidator"
)

var (
	_ resource.Resource                = &DNSRecordResource{}
	_ resource.ResourceWithImportState = &DNSRecordResource{}
)

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

type DNSRecordResource struct {
	client *porkbun.Client
}

type DNSRecordResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
	Type      types.String `tfsdk:"type"`
	Content   types.String `tfsdk:"content"`
	TTL       types.Int64  `tfsdk:"ttl"`
	Prio      types.Int64  `tfsdk:"prio"`
	Notes     types.String `tfsdk:"notes"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage DNS records for domains registered through Porkbun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the DNS record. This is assigned by Porkbun and used for record management.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name for which to create the DNS record (e.g., example.com).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "The subdomain for the record being created, not including the domain itself. Leave blank to create a record on the root domain. Use * to create a wildcard record.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of DNS record to create (A, AAAA, CNAME, MX, TXT, NS, ALIAS, SRV, TLSA, CAA, HTTPS, SVCB).",
				Required:            true,
				Validators: []validator.String{
					enumvalidator.Valid(
						porkbun.A,
						porkbun.MX,
						porkbun.CNAME,
						porkbun.ALIAS,
						porkbun.TXT,
						porkbun.NS,
						porkbun.AAAA,
						porkbun.SRV,
						porkbun.TLSA,
						porkbun.CAA,
						porkbun.HTTPS,
						porkbun.SVCB,
					),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The answer content for the record. Please see the DNS management popup from the domain management console for proper formatting of each record type.",
				Required:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The time to live in seconds for the record. The minimum and the default is 600 seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(600),
				Validators: []validator.Int64{
					int64validator.AtLeast(600),
				},
			},
			"prio": schema.Int64Attribute{
				MarkdownDescription: "The priority of the record for those that support it.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Notes for the DNS record. This is read-only and can only be set from the Porkbun web interface.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := porkbun.DnsRecord{
		Name:    data.Subdomain.ValueString(),
		Type:    porkbun.DnsRecordType(data.Type.ValueString()), // guaranteed to be valid by schema validation
		Content: data.Content.ValueString(),
		TTL:     strconv.FormatInt(data.TTL.ValueInt64(), 10),
		Prio:    strconv.FormatInt(data.Prio.ValueInt64(), 10),
	}

	apiResp, err := r.client.Dns.CreateRecord(ctx, data.Domain.ValueString(), &record)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating DNS Record", err.Error())
		return
	}

	data.ID = types.Int64Value(apiResp.ID)
	data.Notes = types.StringValue("") // empty on create
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("Invalid ID", "ID cannot be null or unknown.")
		return
	}

	record, err := r.getDNSRecord(ctx, data.Domain.ValueString(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Fetching DNS Record", err.Error())
		return
	}

	subdomain, err := r.subdomainFromDomain(record.Name)
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing Subdomain", err.Error())
		return
	}

	data.ID = types.Int64Value(*record.ID)
	data.Subdomain = types.StringValue(subdomain)
	data.Type = types.StringValue(string(record.Type))
	data.Content = types.StringValue(record.Content)
	data.Notes = types.StringValue(record.Notes)

	ttl, _ := strconv.ParseInt(record.TTL, 10, 64)
	data.TTL = types.Int64Value(ttl)

	prio, _ := strconv.ParseInt(record.Prio, 10, 64)
	data.Prio = types.Int64Value(prio)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Dns.EditRecord(ctx, data.Domain.ValueString(), data.ID.ValueInt64(), &porkbun.EditRecord{
		Name:    data.Subdomain.ValueString(),
		Type:    porkbun.DnsRecordType(data.Type.ValueString()), // guaranteed to be valid by schema validation
		Content: data.Content.ValueString(),
		TTL:     strconv.FormatInt(data.TTL.ValueInt64(), 10),
		Prio:    strconv.FormatInt(data.Prio.ValueInt64(), 10),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating DNS Record", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Dns.DeleteRecord(ctx, data.Domain.ValueString(), data.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error Deleting DNS Record", err.Error())
		return
	}
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.SplitN(req.ID, ":", 2)

	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <domain>:<record_id>")
		return
	}

	domain := strings.TrimSpace(idParts[0])
	recordID := strings.TrimSpace(idParts[1])
	if domain == "" || recordID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Domain and record ID cannot be empty. Expected format: <domain>:<record_id>")
		return
	}

	id, err := strconv.ParseInt(recordID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Failed to parse record ID as integer. Expected format: <domain>:<record_id>. Error: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &DNSRecordResourceModel{
		Domain: types.StringValue(domain),
		ID:     types.Int64Value(id),
	})...)
}

// getDNSRecord fetches the DNS record from Porkbun using the domain and record ID.
func (r *DNSRecordResource) getDNSRecord(ctx context.Context, domain string, id int64) (*porkbun.DnsRecord, error) {
	apiResp, err := r.client.Dns.GetRecords(ctx, domain, &id)
	if err != nil {
		return nil, fmt.Errorf("error fetching DNS record: %w", err)
	}

	if len(apiResp.Records) == 0 {
		return nil, fmt.Errorf("no DNS records found for ID %q in domain %q", id, domain)
	} else if len(apiResp.Records) > 1 {
		return nil, fmt.Errorf("multiple DNS records found for ID %q in domain %q", id, domain)
	}

	return &apiResp.Records[0], nil
}

// subdomainFromDomain extracts the subdomain from the full domain name.
func (r *DNSRecordResource) subdomainFromDomain(name string) (string, error) {
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain name: %q", name)
	}

	subdomain := strings.Join(parts[:len(parts)-2], ".")
	if subdomain == "" {
		return "", nil
	}

	return subdomain, nil
}
