# Enable IAM api
resource "google_project_service" "iam_api" {
  service = "iam.googleapis.com"
}

# Create new service account
resource "google_service_account" "siteverifier" {
  account_id   = "google-site-verifier"
  display_name = "Google Site verification account"

  depends_on = [google_project_service.iam_api]
}

# Generate service account key
resource "google_service_account_key" "siteverifier" {
  service_account_id = google_service_account.siteverifier.name
}

# Initialise provider with service account key
provider "googlesiteverification" {
  credentials = base64decode(google_service_account_key.siteverifier.private_key)
}

# Enable site verification api
resource "google_project_service" "siteverification" {
  service = "siteverification.googleapis.com"
}

# Request for DNS token from site verification API
data "googlesiteverification_dns_token" "run_domain" {
  domain = var.domain_name

  depends_on = [google_project_service.siteverification]
}

# Request google to verify the newly added verification record
resource "googlesiteverification_dns" "run_domain" {
  domain = var.domain_name
  token  = data.googlesiteverification_dns_token.run_domain.record_value

  depends_on = [cloudflare_record.google_search_verification_txt]
}

# Get the details of the verification
data "googlesiteverification_site_details" "test" {
  resource_id = googlesiteverification_dns.run_domain.id

  depends_on = [googlesiteverification_dns.run_domain]
}

# Add the details of the new owner
resource "googlesiteverification_add_owner" "test" {
  resource_id             = data.googlesiteverification_site_details.test.id
  site_details_type       = data.googlesiteverification_site_details.test.site_details_type
  site_details_identifier = data.googlesiteverification_site_details.test.site_details_identifier
  owner_details           = concat(data.googlesiteverification_site_details.test.owner_details, ["emailAddress"])

  depends_on = [data.googlesiteverification_site_details.test]
}
