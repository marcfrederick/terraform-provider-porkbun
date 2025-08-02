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

func TestAccURLForwardResource(t *testing.T) {
	// FIXME: Get a dedicated Porkbun domain for testing URL forwarding.
	t.Skipf("The ACCTEST domain is currently not using Porkbuns' nameservers, which is required for URL forwarding to work.")

	const (
		initialSubdomain = "acctest-url-forward"
		updatedSubdomain = "acctest-url-forward-updated"
		initialLocation  = "https://example.com"
		updatedLocation  = "https://test.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccURLForwardResourceConfig(initialSubdomain, initialLocation, "temporary", true, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(regexp.MustCompile(`^\d+$`)),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(testAccDomain()),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("subdomain"),
						knownvalue.StringExact(initialSubdomain),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact(initialLocation),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("temporary"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("include_path"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("wildcard"),
						knownvalue.Bool(false),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:        "porkbun_url_forward.test",
				ImportStateIdPrefix: fmt.Sprintf("%s:", testAccDomain()),
				ImportState:         true,
				// ImportStateVerify:   true,
			},
			// Update and Read testing
			{
				Config: testAccURLForwardResourceConfig(updatedSubdomain, updatedLocation, "permanent", false, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(regexp.MustCompile(`^\d+$`)),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("subdomain"),
						knownvalue.StringExact(updatedSubdomain),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact(updatedLocation),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("permanent"),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("include_path"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"porkbun_url_forward.test",
						tfjsonpath.New("wildcard"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccURLForwardResourceConfig(subdomain, location, redirectType string, includePath, wildcard bool) string {
	return fmt.Sprintf(`
resource "porkbun_url_forward" "test" {
  domain       = %q
  subdomain    = %q
  location     = %q
  type         = %q
  include_path = %t
  wildcard     = %t
}
`, testAccDomain(), subdomain, location, redirectType, includePath, wildcard)
}
