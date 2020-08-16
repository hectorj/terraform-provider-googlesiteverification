package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccSiteOwners(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"googlesiteverification": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "googlesiteverification_owners" "example" {
	domain = "test-terraform-provider.hectorj.net"
	owners = list("provider-ci-runner@terraform-siteverification.iam.gserviceaccount.com")
}`,
				ResourceName:  "googlesiteverification_owners.example",
				ImportState:   true,
				ImportStateId: "test-terraform-provider.hectorj.net",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "domain", "test-terraform-provider.hectorj.net"),
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "owners", ""),
				),
			},
			// 			{
			// 				Config: `
			// resource "googlesiteverification_owners" "example" {
			// 	domain = "test-terraform-provider.hectorj.net"
			// }`,
			// 				ResourceName:      "googlesiteverification_owners.example",
			// 				ImportState:       true,
			// 				ImportStateVerify: true,
			// 				Check: resource.ComposeTestCheckFunc(
			// 					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "domain", "test-terraform-provider.hectorj.net"),
			// 					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "owners", ""),
			// 				),
			// 			},
		},
	})
}
