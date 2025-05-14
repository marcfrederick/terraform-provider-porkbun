package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tuzzmaniandevil/porkbun-go"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/util"
	"github.com/marcfrederick/terraform-provider-porkbun/internal/validator/enumvalidator"
)

var (
	_ resource.Resource                = &DNSSECRecordResource{}
	_ resource.ResourceWithImportState = &DNSSECRecordResource{}
)

// NewDNSSECRecordResource returns a new instance of the resource.
func NewDNSSECRecordResource() resource.Resource {
	return &DNSSECRecordResource{}
}

// DNSSECRecordResource implements CRUD operations for Porkbun DNSSEC records.
type DNSSECRecordResource struct {
	client *porkbun.Client
}

type DNSSECRecordResourceModel struct {
	Domain     types.String        `tfsdk:"domain"`
	MaxSigLife types.Int64         `tfsdk:"max_sig_life"`
	DSData     *DNSSECDSDataModel  `tfsdk:"ds_data"`
	KeyData    *DNSSECKeyDataModel `tfsdk:"key_data"`
}

type DNSSECDSDataModel struct {
	KeyTag     types.String `tfsdk:"key_tag"`
	Algorithm  types.Int64  `tfsdk:"algorithm"`
	DigestType types.Int64  `tfsdk:"digest_type"`
	Digest     types.String `tfsdk:"digest"`
}

type DNSSECKeyDataModel struct {
	Flags     types.Int64  `tfsdk:"flags"`
	Protocol  types.Int64  `tfsdk:"protocol"`
	Algorithm types.Int64  `tfsdk:"algorithm"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (r *DNSSECRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_record"
}

func (r *DNSSECRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages DNSSEC settings (DS/DNSKEY) for a domain registered with Porkbun.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully‑qualified domain name (FQDN) whose DNSSEC settings will be managed, for example `example.com`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_sig_life": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum lifetime of a DNSSEC signature (RRSIG), in seconds. **Note:** The Porkbun API does not return this value, so it cannot be read back and drift detection for this argument is disabled.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ds_data": schema.SingleNestedAttribute{
				MarkdownDescription: "Delegation‑Signer (DS) record parameters. Many registries require DS data to enable DNSSEC, while some ignore or reject it. If your registry returns an error, omit this block and provide `key_data` instead. **At least one of `ds_data` or `key_data` must be supplied.**",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"key_tag": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Key tag (key ID) calculated from the public key. Provided as a decimal string.",
					},
					"algorithm": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "DNSSEC algorithm identifier, as defined in RFC 8624.",
						Validators: []validator.Int64{
							enumvalidator.Valid(
								porkbun.DnssecAlgorithmRsaMd5,
								porkbun.DnssecAlgorithmDsaSha1,
								porkbun.DnssecAlgorithmRsaSha1,
								porkbun.DnssecAlgorithmDsaNsec3Sha1,
								porkbun.DnssecAlgorithmRsaSha256,
								porkbun.DnssecAlgorithmRsaSha512,
								porkbun.DnssecAlgorithmGostR34111994,
								porkbun.DnssecAlgorithmEcdsaSha256,
								porkbun.DnssecAlgorithmEcdsaSha384,
								porkbun.DnssecAlgorithmEd25519,
								porkbun.DnssecAlgorithmEd448,
							),
						},
					},
					"digest_type": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Hash algorithm identifier used to create the DS digest.",
						Validators: []validator.Int64{
							enumvalidator.Valid(
								porkbun.DnssecDigestTypeSha1,
								porkbun.DnssecDigestTypeSha256,
								porkbun.DnssecDigestTypeGostR34111994,
								porkbun.DnssecDigestTypeSha384,
							),
						},
					},
					"digest": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Digest (hash) of the DNSKEY record, encoded in hexadecimal. Length and contents depend on `digest_type`.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("ds_data"),
						path.MatchRoot("key_data"),
					),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"key_data": schema.SingleNestedAttribute{
				MarkdownDescription: "DNSKEY record data. Some registries accept `key_data` instead of, or in addition to, `ds_data`. If DS records are rejected, try creating DNSSEC with `key_data` only. **At least one of `ds_data` or `key_data` must be supplied.**",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"flags": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "DNSKEY flags field (RFC 4034 §2.1). Common values are `256` (ZSK) and `257` (KSK).",
					},
					"protocol": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "DNSSEC protocol value. Must be `3` (DNSSEC).",
						Validators: []validator.Int64{
							int64validator.OneOf(3),
						},
					},
					"algorithm": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "DNSSEC algorithm identifier, as defined in RFC 8624.",
						Validators: []validator.Int64{
							enumvalidator.Valid(
								porkbun.DnssecAlgorithmRsaMd5,
								porkbun.DnssecAlgorithmDsaSha1,
								porkbun.DnssecAlgorithmRsaSha1,
								porkbun.DnssecAlgorithmDsaNsec3Sha1,
								porkbun.DnssecAlgorithmRsaSha256,
								porkbun.DnssecAlgorithmRsaSha512,
								porkbun.DnssecAlgorithmGostR34111994,
								porkbun.DnssecAlgorithmEcdsaSha256,
								porkbun.DnssecAlgorithmEcdsaSha384,
								porkbun.DnssecAlgorithmEd25519,
								porkbun.DnssecAlgorithmEd448,
							),
						},
					},
					"public_key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Base64‑encoded public key material of the DNSKEY record.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("ds_data"),
						path.MatchRoot("key_data"),
					),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DNSSECRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = getPorkbunClient(req.ProviderData, resp.Diagnostics)
}

func (r *DNSSECRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSSECRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnssecRecord := porkbun.DnssecRecordData{
		MaxSigLife: strconv.FormatInt(data.MaxSigLife.ValueInt64(), 10),
	}

	if data.DSData != nil {
		dnssecRecord.KeyTag = data.DSData.KeyTag.ValueString()
		dnssecRecord.Alg = porkbun.DnssecAlgorithm(strconv.FormatInt(data.DSData.Algorithm.ValueInt64(), 10))
		dnssecRecord.DigestType = porkbun.DnssecDigestType(strconv.FormatInt(data.DSData.DigestType.ValueInt64(), 10))
		dnssecRecord.Digest = data.DSData.Digest.ValueString()
	}

	if data.KeyData != nil {
		flags := strconv.FormatInt(data.KeyData.Flags.ValueInt64(), 10)
		protocol := strconv.FormatInt(data.KeyData.Protocol.ValueInt64(), 10)
		algorithm := porkbun.DnssecAlgorithm(strconv.FormatInt(data.KeyData.Algorithm.ValueInt64(), 10))
		dnssecRecord.KeyDataFlags = &flags
		dnssecRecord.KeyDataProtocol = &protocol
		dnssecRecord.KeyDataAlgo = &algorithm
		dnssecRecord.KeyDataPubKey = data.KeyData.PublicKey.ValueStringPointer()
	}

	if _, err := r.client.Dns.CreateDnssecRecord(ctx, data.Domain.ValueString(), &dnssecRecord); err != nil {
		resp.Diagnostics.AddError("Error Creating DNSSEC", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSSECRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// FIXME: Some TLDs allow DNSSEC without ds_data and thus key_tag. Figure
	//        out how the Porkbun API handles this and update the code accordingly.
	dnssecRecord, err := r.readDNSSECRecord(ctx, data.Domain.ValueString(), data.DSData.KeyTag.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading DNSSEC", err.Error())
		return
	}

	if dnssecRecord.Alg != "" {
		data.DSData = &DNSSECDSDataModel{
			KeyTag:     types.StringValue(dnssecRecord.KeyTag),
			Algorithm:  util.Int64Value(dnssecRecord.Alg, &resp.Diagnostics),
			DigestType: util.Int64Value(dnssecRecord.DigestType, &resp.Diagnostics),
			Digest:     types.StringValue(dnssecRecord.Digest),
		}
	}

	if dnssecRecord.KeyDataFlags != nil {
		data.KeyData = &DNSSECKeyDataModel{
			Flags:     util.Int64PointerValue(dnssecRecord.KeyDataFlags, &resp.Diagnostics),
			Protocol:  util.Int64PointerValue(dnssecRecord.KeyDataProtocol, &resp.Diagnostics),
			Algorithm: util.Int64PointerValue(dnssecRecord.KeyDataProtocol, &resp.Diagnostics),
			PublicKey: types.StringPointerValue(dnssecRecord.KeyDataPubKey),
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSSECRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All attributes require replacement; no update logic necessary.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSSECRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Dns.DeleteDnssecRecord(ctx, data.Domain.ValueString(), data.DSData.KeyTag.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting DNSSEC", err.Error())
		return
	}
}

func (r *DNSSECRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.SplitN(req.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <domain>:<key_tag>")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &DNSSECRecordResourceModel{
		Domain: types.StringValue(idParts[0]),
		DSData: &DNSSECDSDataModel{
			KeyTag: types.StringValue(idParts[1]),
		},
	})...)
}

// readDNSSECRecord retrieves the DNSSEC record for the specified domain and key tag.
func (r *DNSSECRecordResource) readDNSSECRecord(ctx context.Context, domain, keyTag string) (*porkbun.DnssecRecordData, error) {
	resp, err := r.client.Dns.GetDnssecRecords(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("error getting DNSSEC records: %w", err)
	}

	dnssecRecord, ok := resp.Records[keyTag]
	if !ok {
		return nil, fmt.Errorf("DNSSEC record not found for key tag %s", keyTag)
	}

	return &dnssecRecord, nil
}
