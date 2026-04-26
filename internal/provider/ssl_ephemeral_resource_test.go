package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccSSLEphemeralResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck: func() {
			if path := os.Getenv("TF_ACC_TERRAFORM_PATH"); strings.HasSuffix(path, "tofu") {
				t.Skipf("OpenTofu does not support ephemeral resources. Skipping test.")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccSSLEphemeralResourceConfig(testAccDomain()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("certificate_chain"),
						knownvalue.StringRegexp(regexp.MustCompile(`\s*-+BEGIN CERTIFICATE-+.+`)),
					),
					// DANGER: This check is useful for local testing, but
					//	 not recommended for CI as it prints the private key
					//	 if the test fails.
					/*statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("private_key"),
						knownvalue.StringRegexp(regexp.MustCompile(`\s*-+BEGIN PRIVATE KEY-+.+`)),
					),*/
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("public_key"),
						knownvalue.StringRegexp(regexp.MustCompile(`\s*-+BEGIN PUBLIC KEY-+.+`)),
					),
				},
			},
		},
	})
}

func testAccSSLEphemeralResourceConfig(domain string) string {
	return fmt.Sprintf(`
ephemeral "porkbun_ssl" "test" {
  domain = %q
}

provider "echo" {
  data = ephemeral.porkbun_ssl.test
}

resource "echo" "test" {}
`, domain)
}
