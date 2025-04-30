package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/tuzzmaniandevil/porkbun-go"
)

func TestAccDNSRecordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSRecordResourceConfig("acctest", "content", porkbun.TXT, 3600, 0),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("id"),
						knownvalue.Int64Func(func(v int64) error {
							if v == 0 {
								return fmt.Errorf("ID must not be 0")
							}
							return nil
						}),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("subdomain"),
						knownvalue.StringExact("acctest"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("TXT"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact("content"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("ttl"),
						knownvalue.Int64Exact(3600),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("prio"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact(""),
					),
				},
			},
			// ImportState testing
			// FIXME: This depends on a fixed ID
			//{
			//	ResourceName:      "porkbun_dns_record.test",
			//	ImportStateId:     fmt.Sprintf("%s:12345", testAccDomain()),
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			// Update and Read testing
			{
				Config: testAccDNSRecordResourceConfig("acctest", "updated content", porkbun.TXT, 3601, 10),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact("updated content"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("ttl"),
						knownvalue.Int64Exact(3601),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dns_record.test",
						tfjsonpath.New("prio"),
						knownvalue.Int64Exact(10),
					),
				},
			},
		},
	})
}

func testAccDNSRecordResourceConfig(subdomain, content string, recordType porkbun.DnsRecordType, ttl, prio int) string {
	return fmt.Sprintf(`
resource "porkbun_dns_record" "test" {
  domain    = %[1]q
  subdomain = %[2]q
  type      = %[3]q	
  content   = %[4]q
  ttl       = %[5]d
  prio      = %[6]d
}
`, testAccDomain(), subdomain, recordType, content, ttl, prio)
}

func TestDNSRecordResource_subdomainFromDomain(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"simple", args{"example.com"}, "", false},
		{"subdomain", args{"foo.example.com"}, "foo", false},
		{"multiple subdomains", args{"foo.bar.example.com"}, "foo.bar", false},
		{"invalid domain", args{"example"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DNSRecordResource{}
			got, err := r.subdomainFromDomain(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("subdomainFromDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("subdomainFromDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}
