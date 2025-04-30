package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccDomainsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:            testAccDomainsDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{},
			},
		},
	})
}

const testAccDomainsDataSourceConfig = `data "porkbun_domains" "test" {}`
