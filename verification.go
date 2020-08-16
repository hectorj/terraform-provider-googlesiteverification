package main

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/api/siteverification/v1"
)

const verificationMethod = "DNS_TXT"
const recordTypeKey = "record_type"
const recordNameKey = "record_name"
const recordValueKey = "record_value"
const credentialsKey = "credentials"
const siteType = "INET_DOMAIN"

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

	resourceData.SetId(domainKey + "_token")
	if setErr := resourceData.Set(domainKey, domain); setErr != nil {
		return setErr
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

	return nil
}

func deleteDnsSiteVerification(resourceData *schema.ResourceData, provider interface{}) error {
	return nil // no-op, user should remove the DNS token
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
		siteResource, insertErr := service.WebResource.Insert(verificationMethod, &siteverification.SiteVerificationWebResourceResource{
			Site: &siteverification.SiteVerificationWebResourceResourceSite{
				Identifier: domain,
				Type:       siteType,
			},
		}).Do()
		if insertErr != nil {
			return resource.RetryableError(insertErr)
		}

		resourceId, unescapeErr := url.QueryUnescape(siteResource.Id)
		if unescapeErr != nil {
			return resource.NonRetryableError(unescapeErr)
		}
		resourceData.SetId(resourceId)

		return resource.NonRetryableError(readDnsSiteVerification(resourceData, provider))
	})
}
