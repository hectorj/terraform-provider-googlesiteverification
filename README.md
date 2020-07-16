# Terraform provider for Google's site verification

A simple provider hitting this API: https://developers.google.com/site-verification

## Usage

```hcl-terraform
# We get the verification token from Google
data "googlesiteverification_dns_token" "example" {
  domain = "yourdomain.example.com"
}

# We put it in our DNS records
# Here is an example with Cloudflare, but you should be able to adapt it to any DNS provider
resource "cloudflare_record" "verification" {
  zone_id = "{your zone ID}"
  name    = data.googlesiteverification_dns_token.example.record_name
  value   = data.googlesiteverification_dns_token.example.record_value
  type    = data.googlesiteverification_dns_token.example.record_type # "TXT" only for now, but it's better to use the data value for future-proofing.
}

# *After* that, we submit our verification request to Google.
# Might take some times, depending on Google's DNS caching.
resource "googlesiteverification_dns" "example" {
  domain = "yourdomain.example.com"
  depends_on = [cloudflare_record.verification] 
}
```
