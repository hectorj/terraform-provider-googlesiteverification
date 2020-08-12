package main

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDnsSiteVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"googlesiteverification": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: `
data "googlesiteverification_dns_token" "example" {
	domain = "test-terraform-provider.hectorj.net"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.googlesiteverification_dns_token.example", "domain", "test-terraform-provider.hectorj.net"),
					resource.TestCheckResourceAttr("data.googlesiteverification_dns_token.example", "record_type", "TXT"),
					resource.TestCheckResourceAttr("data.googlesiteverification_dns_token.example", "record_name", "test-terraform-provider.hectorj.net"),
					resource.TestMatchResourceAttr("data.googlesiteverification_dns_token.example", "record_value", regexp.MustCompile(`^google-site-verification=[A-Za-z0-9_]+$`)),
				),
			},
			{
				Config: `
resource "googlesiteverification_dns" "example" {
	domain = "test-terraform-provider.hectorj.net"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_dns.example", "domain", "test-terraform-provider.hectorj.net"),
				),
			},
			{
				Config: `
resource "googlesiteverification_dns" "example" {
	domain = "test-terraform-provider.hectorj.net"
}`,
				ResourceName:      "googlesiteverification_dns.example",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_dns.example", "domain", "test-terraform-provider.hectorj.net"),
				),
			},
		},
	})
}
