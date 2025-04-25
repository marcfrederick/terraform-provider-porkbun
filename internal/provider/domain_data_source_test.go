package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDomainDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("status"),
						knownvalue.StringExact("ACTIVE"),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("tld"),
						knownvalue.StringExact("com"),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("security_lock"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("whois_privacy"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("auto_renew"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("not_local"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_domain.test",
						tfjsonpath.New("labels"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

func testAccDomainDataSourceConfig() string {
	return fmt.Sprintf(`
data "porkbun_domain" "test" {
  domain = %q
}
`, testAccDomain())
}
