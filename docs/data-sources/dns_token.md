---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "googlesiteverification_dns_token Data Source - terraform-provider-googlesiteverification"
subcategory: ""
description: |-
  https://developers.google.com/site-verification/v1/webResource/getToken
---

# googlesiteverification_dns_token (Data Source)

https://developers.google.com/site-verification/v1/webResource/getToken



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) The domain you want to verify.

### Read-Only

- `id` (String) The ID of this resource.
- `record_name` (String) The name of the record you should create.
- `record_type` (String) The type of DNS record you should create.
- `record_value` (String) The value of the record you should create.

