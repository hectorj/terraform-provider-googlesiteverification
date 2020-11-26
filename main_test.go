package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/cloudflare/terraform-provider-cloudflare/cloudflare"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDnsSiteVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"googlesiteverification": Provider(),
			"cloudflare":             cloudflare.Provider(),
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
data "googlesiteverification_dns_token" "example" {
	domain = "test-terraform-provider.hectorj.net"
}

resource "cloudflare_record" "verification" {
  zone_id = "` + os.Getenv("CLOUDFLARE_ZONE_ID") + `"
  name    = data.googlesiteverification_dns_token.example.record_name
  value   = data.googlesiteverification_dns_token.example.record_value
  type    = data.googlesiteverification_dns_token.example.record_type
}

resource "googlesiteverification_dns" "example" {
	domain     = "test-terraform-provider.hectorj.net"
	token      = data.googlesiteverification_dns_token.example.record_value
	depends_on = [cloudflare_record.verification]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_dns.example", "domain", "test-terraform-provider.hectorj.net"),
					resource.TestMatchResourceAttr("googlesiteverification_dns.example", "token", regexp.MustCompile(`^google-site-verification=[A-Za-z0-9_]+$`)),
				),
			},
			{
				Config: `
data "googlesiteverification_dns_token" "example" {
	domain = "test-terraform-provider.hectorj.net"
}

resource "cloudflare_record" "verification" {
  zone_id = "` + os.Getenv("CLOUDFLARE_ZONE_ID") + `"
  name    = data.googlesiteverification_dns_token.example.record_name
  value   = data.googlesiteverification_dns_token.example.record_value
  type    = data.googlesiteverification_dns_token.example.record_type
}

resource "googlesiteverification_dns" "example" {
	domain     = "test-terraform-provider.hectorj.net"
	token      = data.googlesiteverification_dns_token.example.record_value
	depends_on = [cloudflare_record.verification]
}`,
				ResourceName:      "googlesiteverification_dns.example",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_dns.example", "domain", "test-terraform-provider.hectorj.net"),
					resource.TestMatchResourceAttr("googlesiteverification_dns.example", "token", regexp.MustCompile(`^google-site-verification=[A-Za-z0-9_]+$`)),
				),
			},
		},
	})
}
