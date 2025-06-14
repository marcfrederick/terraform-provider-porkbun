package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNameserversDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNameserversDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.porkbun_nameservers.test",
						tfjsonpath.New("nameservers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("maceio.ns.porkbun.com"),
							knownvalue.StringExact("salvador.ns.porkbun.com"),
							knownvalue.StringExact("fortaleza.ns.porkbun.com"),
							knownvalue.StringExact("curitiba.ns.porkbun.com"),
						}),
					),
				},
			},
		},
	})
}

func testAccNameserversDataSourceConfig() string {
	return fmt.Sprintf(`
data "porkbun_nameservers" "test" {
  domain = %q
}
`, testAccDomain())
}
