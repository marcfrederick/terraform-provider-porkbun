package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"porkbun": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccDomain returns the domain to be used for acceptance tests.
func testAccDomain() string {
	return os.Getenv("PORKBUN_ACCTEST_DOMAIN")
}

func testAccPreCheck(t *testing.T) {
	checkEnv(t, "PORKBUN_API_KEY")
	checkEnv(t, "PORKBUN_SECRET_API_KEY")
	checkEnv(t, "PORKBUN_ACCTEST_DOMAIN")
}

// checkEnv checks if the given environment variable is set.
func checkEnv(t *testing.T, env string) {
	if v := os.Getenv(env); v == "" {
		t.Fatalf("%s must be set for acceptance tests", env)
	}
}
