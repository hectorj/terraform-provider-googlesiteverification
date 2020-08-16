package main

import (
	"fmt"
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
				ImportState:   true,
				ImportStateId: "test-terraform-provider.hectorj.net",
				ResourceName:  "googlesiteverification_owners.example",
				ImportStateCheck: func(instanceStates []*terraform.InstanceState) error {
					if want, got := 1, len(instanceStates); want != got {
						return fmt.Errorf("expected exactly %d imported resource, got %d", want, got)
					}
					if want, got := "dns://test-terraform-provider.hectorj.net", instanceStates[0].ID; want != got {
						return fmt.Errorf("expected ID %q, got %q", want, got)
					}
					if want, got := "test-terraform-provider.hectorj.net", instanceStates[0].Attributes["domain"]; want != got {
						return fmt.Errorf("expected domain %q, got %q", want, got)
					}
					// TODO: find a better way to check a set's values
					if want, got := "1", instanceStates[0].Attributes["owners.#"]; want != got {
						return fmt.Errorf("expected owners count %q, got %q", want, got)
					}
					if want, got := "provider-ci-runner@terraform-siteverification.iam.gserviceaccount.com", instanceStates[0].Attributes["owners.464778800"]; want != got {
						return fmt.Errorf("expected owner %q, got %q", want, got)
					}
					return nil
				},
			},
			{
				Config: `
resource "googlesiteverification_owners" "example" {
	domain = "test-terraform-provider.hectorj.net"
	owners = list("provider-ci-runner@terraform-siteverification.iam.gserviceaccount.com")
}`,
				ResourceName: "googlesiteverification_owners.example",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "id", "dns://test-terraform-provider.hectorj.net"),
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "domain", "test-terraform-provider.hectorj.net"),
					// TODO: find a better way to check a set's values
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "owners.#", "1"),
					resource.TestCheckResourceAttr("googlesiteverification_owners.example", "owners.464778800", "provider-ci-runner@terraform-siteverification.iam.gserviceaccount.com"),
				),
			},
		},
	})
}
