# Terraform Provider for Google Site Verification

This repository hosts the Terraform provider for Google Site Verification. The provider enables Terraform to manage Google Site Verification records programmatically.

## Features

- Verify ownership of websites and domains via Google Site Verification.
- Supports multiple verification methods, including HTML file, meta tag, DNS TXT record, and Google Analytics.

A simple provider hitting this API: https://developers.google.com/site-verification

See https://registry.terraform.io/providers/hectorj/googlesiteverification

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0.0
- [Go](https://golang.org/doc/install) >= 1.17 (for development)
- Google Cloud account with Site Verification API enabled.

## Installation

### From Terraform Registry

The provider can be installed directly from the Terraform Registry by including it in your Terraform configuration file:

```hcl
terraform {
  required_providers {
    googlesiteverification = {
      source  = "hectorj/googlesiteverification"
      version = "1.0.0" # Replace with the latest version
    }
  }
}
```

### From Source

To build and install the provider from source:

1. Clone this repository:

   ```bash
   git clone https://github.com/hectorj/terraform-provider-googlesiteverification.git
   cd terraform-provider-googlesiteverification
   ```

2. Build the provider:

   ```bash
   go build -o terraform-provider-googlesiteverification
   ```

3. Move the binary to your Terraform plugins directory:

   ```bash
   mkdir -p ~/.terraform.d/plugins/hectorj/googlesiteverification/1.0.0/linux_amd64/
   mv terraform-provider-googlesiteverification ~/.terraform.d/plugins/hectorj/googlesiteverification/1.0.0/linux_amd64/
   ```

4. Update your Terraform configuration file to use the local provider binary.

## Usage

```hcl
provider "googlesiteverification" {
  # Add provider-specific configuration here
}

resource "googlesiteverification" "example" {
  method = "DNS"
  domain = "example.com"
}
```

## Authentication

To use this provider, you must authenticate with Google Cloud:

1. Create a Google Cloud project and enable the Site Verification API.
2. Generate a service account key:
   - Go to the [Google Cloud Console](https://console.cloud.google.com/).
   - Navigate to IAM & Admin > Service Accounts.
   - Create a new service account and download the key file in JSON format.
3. Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of your service account key file:

   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/keyfile.json"
   ```

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## References

- [Terraform Documentation](https://www.terraform.io/docs/index.html)
- [Google Site Verification API](https://developers.google.com/site-verification/)

## Support

For support, please open an issue in the [GitHub repository](https://github.com/hectorj/terraform-provider-googlesiteverification/issues).
