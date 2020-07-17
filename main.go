package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/siteverification/v1"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "install" {
		install()
		return
	}
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

const domainKey = "domain"
const recordTypeKey = "record_type"
const recordNameKey = "record_name"
const recordValueKey = "record_value"
const credentialsKey = "credentials"
const siteType = "INET_DOMAIN"
const verificationMethod = "DNS_TXT"

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			credentialsKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ConfigureFunc: configureProvider,
		DataSourcesMap: map[string]*schema.Resource{
			"googlesiteverification_dns_token": {
				Schema: map[string]*schema.Schema{
					domainKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					recordTypeKey: {
						Type:     schema.TypeString,
						Computed: true,
					},
					recordNameKey: {
						Type:     schema.TypeString,
						Computed: true,
					},
					recordValueKey: {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
				Description: "https://developers.google.com/site-verification/v1/webResource/getToken",
				Read:        readDnsSiteVerificationToken,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"googlesiteverification_dns": {
				Schema: map[string]*schema.Schema{
					domainKey: {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
				},
				Create:      createDnsSiteVerification,
				Read:        readDnsSiteVerification,
				Delete:      deleteDnsSiteVerification,
				Description: "https://developers.google.com/site-verification",
				Timeouts: &schema.ResourceTimeout{
					Create: schema.DefaultTimeout(60 * time.Minute),
				},
			},
		},
	}
}

type configuredProvider struct {
	service *siteverification.Service
}

func configureProvider(resourceData *schema.ResourceData) (interface{}, error) {
	ctx := context.Background()

	credentialsClientOption, crendentialsErr := findCredentials(resourceData, ctx)
	if crendentialsErr != nil {
		return nil, crendentialsErr
	}

	service, serviceErr := siteverification.NewService(ctx, credentialsClientOption)
	if serviceErr != nil {
		return nil, serviceErr
	}

	return configuredProvider{
		service: service,
	}, nil
}

func findCredentials(resourceData *schema.ResourceData, ctx context.Context) (option.ClientOption, error) {
	// here we are trying to match the official GCP Provider's behavior https://www.terraform.io/docs/providers/google/guides/provider_reference.html#full-reference
	var credentialsLiteral string
	if credentialsFromConfig, ok := resourceData.GetOk(credentialsKey); ok {
		credentialsLiteral = credentialsFromConfig.(string)
	} else if credentialsFromEnv, ok := os.LookupEnv("GOOGLE_CREDENTIALS"); ok {
		credentialsLiteral = credentialsFromEnv
	} else if credentialsFromEnv, ok := os.LookupEnv("GOOGLE_CLOUD_KEYFILE_JSON"); ok {
		credentialsLiteral = credentialsFromEnv
	} else if credentialsFromEnv, ok := os.LookupEnv("GCLOUD_KEYFILE_JSON"); ok {
		credentialsLiteral = credentialsFromEnv
	}

	var credentialsClientOption option.ClientOption
	if credentialsLiteral != "" {
		if json.Valid([]byte(credentialsLiteral)) {
			credentialsClientOption = option.WithCredentialsJSON([]byte(credentialsLiteral))
		} else {
			_, statErr := os.Stat(credentialsLiteral)
			if statErr != nil {
				return nil, statErr
			}
			credentialsClientOption = option.WithCredentialsFile(credentialsLiteral)
		}
	} else {
		credentials, defaultCredentialsErr := google.FindDefaultCredentials(ctx)
		if defaultCredentialsErr != nil {
			return nil, defaultCredentialsErr
		}
		credentialsClientOption = option.WithCredentials(credentials)
	}
	return credentialsClientOption, nil
}

func readDnsSiteVerificationToken(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
	domain := resourceData.Get(domainKey).(string)

	tokenResource, getTokenErr := service.WebResource.GetToken(&siteverification.SiteVerificationWebResourceGettokenRequest{
		Site: &siteverification.SiteVerificationWebResourceGettokenRequestSite{
			Identifier: domain,
			Type:       siteType,
		},
		VerificationMethod: verificationMethod,
	}).Do()
	if getTokenErr != nil {
		return getTokenErr
	}

	if setErr := resourceData.Set(recordTypeKey, "TXT"); setErr != nil {
		return setErr
	}
	if setErr := resourceData.Set(recordNameKey, domain); setErr != nil {
		return setErr
	}
	if setErr := resourceData.Set(recordValueKey, tokenResource.Token); setErr != nil {
		return setErr
	}
	resourceData.SetId(domain)

	return nil
}

func deleteDnsSiteVerification(_ *schema.ResourceData, _ interface{}) error {
	return nil // no-op, user should just remove the DNS record
}

func readDnsSiteVerification(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service

	_, getErr := service.WebResource.Get(resourceData.Id()).Do()
	return getErr
}

func createDnsSiteVerification(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
	domain := resourceData.Get(domainKey).(string)

	return resource.Retry(resourceData.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, insertErr := service.WebResource.Insert(verificationMethod, &siteverification.SiteVerificationWebResourceResource{
			Site: &siteverification.SiteVerificationWebResourceResourceSite{
				Identifier: domain,
				Type:       siteType,
			},
		}).Do()
		if insertErr != nil {
			return resource.RetryableError(insertErr)
		}

		resourceData.SetId(domain)

		return resource.NonRetryableError(readDnsSiteVerification(resourceData, provider))
	})
}
