resource "cloudflare_record" "google_search_verification_txt" {
  name    = data.googlesiteverification_dns_token.run_domain.record_name
  proxied = false
  ttl     = 3600
  type    = data.googlesiteverification_dns_token.run_domain.record_type
  value   = data.googlesiteverification_dns_token.run_domain.record_value
  zone_id = var.cloudflare_zone_id

  depends_on = [data.googlesiteverification_dns_token.run_domain]
}
