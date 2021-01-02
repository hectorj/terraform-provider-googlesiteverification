package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/cloudflare/terraform-provider-cloudflare/cloudflare"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDnsSiteVerification(t *testing.T) {
	domain := fmt.Sprintf("%s-test-terraform-provider.hectorj.net", uuid.New())

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"googlesiteverification": Provider(),
			"cloudflare":             cloudflare.Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: `
data "googlesiteverification_dns_token" "example" {
	domain = "` + domain + `"
}

resource "cloudflare_record" "verification" {
  zone_id = "` + os.Getenv("CLOUDFLARE_ZONE_ID") + `"
  name    = data.googlesiteverification_dns_token.example.record_name
  value   = data.googlesiteverification_dns_token.example.record_value
  type    = data.googlesiteverification_dns_token.example.record_type
}

resource "googlesiteverification_dns" "example" {
	domain     = "` + domain + `"
	token      = data.googlesiteverification_dns_token.example.record_value
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googlesiteverification_dns.example", "domain", domain),
					resource.TestMatchResourceAttr("googlesiteverification_dns.example", "token", regexp.MustCompile(`^google-site-verification=[A-Za-z0-9_-]+$`)),
				),
			},
		},
	})
}
