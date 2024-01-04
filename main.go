package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/siteverification/v1"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	if len(os.Args) > 1 && os.Args[1] == "install" {
		install()
		return
	}
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

const tokenKey = "token"
const domainKey = "domain"
const recordTypeKey = "record_type"
const recordNameKey = "record_name"
const recordValueKey = "record_value"
const credentialsKey = "credentials"
const resourceId = "resource_id"
const webResource_site_type = "site_details_type"
const webResource_site_identifier = "site_details_identifier"
const webResource_owner = "owner_details"
const siteType = "INET_DOMAIN"
const verificationMethod = "DNS_TXT"
const tokenStillExists = "You cannot unverify your ownership of this site until your verification token (meta tag, HTML file, Google Analytics tracking code, Google Tag Manager container code, or DNS record) has been removed."

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			credentialsKey: {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CREDENTIALS",
					"GOOGLE_CLOUD_KEYFILE_JSON",
					"GCLOUD_KEYFILE_JSON",
				}, ""),
				Description: "Either the path to or the contents of a [service account key file](https://cloud.google.com/iam/docs/creating-managing-service-account-keys) in JSON format. If not provided, the [application default credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default) will be used.",
			},
		},
		ConfigureFunc: configureProvider,
		DataSourcesMap: map[string]*schema.Resource{
			"googlesiteverification_dns_token": {
				Schema: map[string]*schema.Schema{
					domainKey: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The domain you want to verify.",
					},
					recordTypeKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of DNS record you should create.",
					},
					recordNameKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The name of the record you should create.",
					},
					recordValueKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The value of the record you should create.",
					},
				},
				Description: "https://developers.google.com/site-verification/v1/webResource/getToken",
				Read:        readDnsSiteVerificationToken,
			},
			"googlesiteverification_site_details": {
				Schema: map[string]*schema.Schema{
					resourceId: {
						Type:     schema.TypeString,
						Required: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_owner: {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Computed: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_site_identifier: {
						Type:     schema.TypeString,
						Computed: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_site_type: {
						Type:     schema.TypeString,
						Computed: true,
						// Description: "The id of the resource you want to get details for.",
					},
				},
				Description: "https://developers.google.com/site-verification/v1/webResource/get",
				Read:        readDnsSiteInformation,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"googlesiteverification_dns": {
				Schema: map[string]*schema.Schema{
					domainKey: {
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
						Description: "The domain you want to verify.",
					},
					tokenKey: {
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
						Description: "The token you got from data.googlesiteverification_dns_token. This forces a new verification in case the token changes.",
					},
				},
				Create:      createDnsSiteVerification,
				Read:        readDnsSiteVerification,
				Delete:      deleteDnsSiteVerification,
				Description: "https://developers.google.com/site-verification",
				Timeouts: &schema.ResourceTimeout{
					Create: schema.DefaultTimeout(60 * time.Minute),
				},
				Importer: &schema.ResourceImporter{
					State: importSiteVerification,
				},
			},
			"googlesiteverification_add_owner": {
				Schema: map[string]*schema.Schema{
					resourceId: {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_owner: {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Required: true,
						ForceNew: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_site_identifier: {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
						// Description: "The id of the resource you want to get details for.",
					},
					webResource_site_type: {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
						// Description: "The id of the resource you want to get details for.",
					},
				},
				Create:      createDnsSiteInformation,
				Read:        readDnsSiteInformation,
				Delete:      deleteDnsSiteVerification,
				Description: "https://developers.google.com/site-verification",
				Timeouts: &schema.ResourceTimeout{
					Create: schema.DefaultTimeout(60 * time.Minute),
				},
				// Importer: &schema.ResourceImporter{
				// 	State: importSiteVerification,
				// },
			},
		},
	}
}

func importSiteVerification(resourceData *schema.ResourceData, provider interface{}) ([]*schema.ResourceData, error) {
	service := provider.(configuredProvider).service
	domain := strings.TrimPrefix(resourceData.Id(), "dns://")

	if setErr := resourceData.Set(domainKey, domain); setErr != nil {
		return nil, setErr
	}

	_, getErr := service.WebResource.Get(resourceData.Id()).Do()
	if getErr != nil {
		return nil, getErr
	}

	// fetch and set the token's value
	tokenResource, getTokenErr := service.WebResource.GetToken(&siteverification.SiteVerificationWebResourceGettokenRequest{
		Site: &siteverification.SiteVerificationWebResourceGettokenRequestSite{
			Identifier: domain,
			Type:       siteType,
		},
		VerificationMethod: verificationMethod,
	}).Do()
	if getTokenErr != nil {
		return nil, getTokenErr
	}
	if setErr := resourceData.Set(tokenKey, tokenResource.Token); setErr != nil {
		return nil, setErr
	}

	return []*schema.ResourceData{resourceData}, nil
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
		scopes := []string{
			"https://www.googleapis.com/auth/siteverification",
		}
		credentials, defaultCredentialsErr := google.FindDefaultCredentials(ctx, scopes...)
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

func readDnsSiteInformation(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
	id := resourceData.Get(resourceId).(string)

	slog.Info("raw ID: ", "id", id)
	decodedInputId, err := url.QueryUnescape(id)
	if err != nil {
		errString := fmt.Errorf(
			"failed to urldecode id %s, %s", id, err)
		log.Fatalln(errString)
		return errString
	}
	slog.Info("decoded ID: ", "id", decodedInputId)

	siteData, getErr := service.WebResource.Get(decodedInputId).Do()
	if getErr != nil {
		return getErr
	}

	if setErr := resourceData.Set(webResource_site_identifier, siteData.Site.Identifier); setErr != nil {
		return setErr
	}
	if setErr := resourceData.Set(webResource_site_type, siteData.Site.Type); setErr != nil {
		return setErr
	}

	if setErr := resourceData.Set(webResource_owner, siteData.Owners); setErr != nil {
		return setErr
	}

	decodedOutputId, err := url.QueryUnescape(siteData.Id)
	if err != nil {
		errString := fmt.Errorf(
			"failed to urldecode id %s, %s", siteData.Id, err)
		log.Fatalln(errString)
		return errString
	}
	slog.Info("decoded ID: ", "id", decodedOutputId)

	resourceData.SetId(decodedOutputId)

	return nil
}

func deleteDnsSiteVerification(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service

	id := resourceData.Id()
	if !strings.HasPrefix(resourceData.Id(), "dns://") {
		// the provider 0.3.1 and earlier stored the domain as
		// the id, which is incorrect.
		id = fmt.Sprintf("dns://%s", id)
	}

	return resource.Retry(resourceData.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.WebResource.Delete(id).Do()
		if err != nil {
			if strings.Contains(err.Error(), tokenStillExists) {
				log.Printf("retry: %s", err)
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})
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
		r, insertErr := service.WebResource.Insert(verificationMethod, &siteverification.SiteVerificationWebResourceResource{
			Site: &siteverification.SiteVerificationWebResourceResourceSite{
				Identifier: domain,
				Type:       siteType,
			},
		}).Do()
		if insertErr != nil {
			log.Printf("retrying failed site verification request, %s", insertErr)
			return resource.RetryableError(insertErr)
		}

		id, err := url.QueryUnescape(r.Id)
		if err != nil {
			return resource.NonRetryableError(
				fmt.Errorf(
					"failed to urldecode id %s, %s", r.Id, err))
		}

		resourceData.SetId(id)

		return resource.NonRetryableError(readDnsSiteVerification(resourceData, provider))
	})
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func createDnsSiteInformation(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
	id := resourceData.Get(resourceId).(string)
	domain := resourceData.Get(webResource_site_identifier).(string)
	site_type := resourceData.Get(webResource_site_type).(string)

	ownerInferfaceArray := resourceData.Get(webResource_owner).([]interface{})

	var owners []string

	for _, v := range ownerInferfaceArray {
		owner := v.(string)
		log.Println(owner)
		owners = append(owners, owner)
	}

	return resource.Retry(resourceData.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, insertErr := service.WebResource.Update(
			id, &siteverification.SiteVerificationWebResourceResource{
				Site: &siteverification.SiteVerificationWebResourceResourceSite{
					Identifier: domain,
					Type:       site_type,
				},
				Owners: owners,
			}).Do()
		if insertErr != nil {
			log.Printf("retrying failed site verification request, %s", insertErr)
			return resource.RetryableError(insertErr)
		}

		id, err := url.QueryUnescape(r.Id)
		if err != nil {
			return resource.NonRetryableError(
				fmt.Errorf(
					"failed to urldecode id %s, %s", r.Id, err))
		}

		resourceData.SetId(id)

		return resource.NonRetryableError(readDnsSiteInformation(resourceData, provider))
	})
}
