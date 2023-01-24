terraform {
  required_providers {
    googlesiteverification = {
      source  = "hectorj/googlesiteverification"
    }
  }
}

resource "google_project_service" "siteverification" {
  service = "siteverification.googleapis.com"
}

data "googlesiteverification_dns_token" "domain" {
  domain     = var.domain_name
  depends_on = [google_project_service.siteverification]
}

resource "google_dns_managed_zone" "domain" {
  name     = replace(var.domain_name, ".", "-")
  dns_name = "${var.domain_name}."
}

resource "google_dns_record_set" "domain_ns_records" {
  name         = "${var.domain_name}."
  type         = "NS"
  ttl          = 60
  managed_zone = google_dns_managed_zone.domain.name
  rrdatas      = google_dns_managed_zone.domain.name_servers
}

resource "google_dns_record_set" "domain" {
  managed_zone = google_dns_managed_zone.domain.name
  name         = "${data.googlesiteverification_dns_token.domain.record_name}."
  rrdatas      = [data.googlesiteverification_dns_token.domain.record_value]
  type         = data.googlesiteverification_dns_token.domain.record_type
  ttl          = 60
}

resource "googlesiteverification_dns" "domain" {
  domain     = var.domain_name
  token      = data.googlesiteverification_dns_token.domain.record_value
  depends_on = [google_dns_record_set.domain_ns_records]
}