package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSSLDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSLDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.porkbun_ssl.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_ssl.test",
						tfjsonpath.New("certificate_chain"),
						knownvalue.StringRegexp(regexp.MustCompile(`\s*-+BEGIN CERTIFICATE-+.+`)),
					),
					statecheck.ExpectSensitiveValue(
						"data.porkbun_ssl.test",
						tfjsonpath.New("private_key"),
					),
					statecheck.ExpectKnownValue(
						"data.porkbun_ssl.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringRegexp(regexp.MustCompile(`\s*-+BEGIN PUBLIC KEY-+.+`)),
					),
				},
			},
		},
	})
}

func testAccSSLDataSourceConfig() string {
	return fmt.Sprintf(`
data "porkbun_ssl" "test" {
  domain = %q
}
`, testAccDomain())
}
