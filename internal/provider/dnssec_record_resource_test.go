package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDNSSECRecordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSSECRecordResourceConfig(86400, "64087", 13, 2, "15E445BD08128BDC213E25F1C8227DF4CB35186CAC701C1C335B2C406D5530DC"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("max_sig_life"),
						knownvalue.Int64Exact(86400),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("key_tag"),
						knownvalue.StringExact("64087"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("algorithm"),
						knownvalue.Int64Exact(13),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("digest_type"),
						knownvalue.Int64Exact(2),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("digest"),
						knownvalue.StringExact("15E445BD08128BDC213E25F1C8227DF4CB35186CAC701C1C335B2C406D5530DC"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName: "porkbun_dnssec_record.test",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					record, ok := state.RootModule().Resources["porkbun_dnssec_record.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}

					keyTag, ok := record.Primary.Attributes["ds_data.key_tag"]
					if !ok {
						return "", fmt.Errorf("key_tag not found in resource attributes")
					}

					return fmt.Sprintf("%s:%s", testAccDomain(), keyTag), nil
				},
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ds_data.key_tag",
				ImportStateVerifyIgnore:              []string{"max_sig_life"}, // can't be imported
			},
			// Update and Read testing
			{
				Config: testAccDNSSECRecordResourceConfig(3600, "64087", 13, 2, "15E445BD08128BDC213E25F1C8227DF4CB35186CAC701C1C335B2C406D5530DC"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("max_sig_life"),
						knownvalue.Int64Exact(3600),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("key_tag"),
						knownvalue.StringExact("64087"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("algorithm"),
						knownvalue.Int64Exact(13),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("digest_type"),
						knownvalue.Int64Exact(2),
					),
					statecheck.ExpectKnownValue(
						"porkbun_dnssec_record.test",
						tfjsonpath.New("ds_data").AtMapKey("digest"),
						knownvalue.StringExact("15E445BD08128BDC213E25F1C8227DF4CB35186CAC701C1C335B2C406D5530DC"),
					),
				},
			},
		},
	})
}

func testAccDNSSECRecordResourceConfig(maxSigLife int64, keyTag string, algorithm, digestType int64, digest string) string {
	return fmt.Sprintf(`
resource "porkbun_dnssec_record" "test" {
  domain       = %q
  max_sig_life = %d

  ds_data = {
	key_tag      = %q
	algorithm    = %d
	digest_type  = %d
	digest       = %q
  }
}
`, testAccDomain(), maxSigLife, keyTag, algorithm, digestType, digest)
}
