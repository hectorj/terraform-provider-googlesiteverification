package main

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const ownersKey = "owners"

func importSiteOwners(resourceData *schema.ResourceData, provider interface{}) ([]*schema.ResourceData, error) {
	domain := resourceData.Id()
	resourceId := resourceData.Id()
	if !strings.Contains(domain, "://") {
		resourceId = resourceIdFromDomain(resourceId)
	} else {
		domain = strings.SplitN(resourceId, "://", 2)[1]
	}

	resourceData.SetId(resourceId)
	if setErr := resourceData.Set(domainKey, domain); setErr != nil {
		return nil, setErr
	}

	readErr := readSiteOwners(resourceData, provider)
	if readErr != nil {
		return nil, readErr
	}

	return []*schema.ResourceData{resourceData}, nil
}

func deleteSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service

	deleteErr := service.WebResource.Delete(resourceData.Id()).Do()
	return deleteErr
}

func readSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service

	siteResource, getErr := service.WebResource.Get(resourceData.Id()).Do()
	if getErr != nil {
		return getErr
	}

	ownersSet := schema.NewSet(schema.HashString, nil)
	for _, owner := range siteResource.Owners {
		ownersSet.Add(owner)
	}
	setOwnersErr := resourceData.Set(ownersKey, ownersSet)

	return setOwnersErr
}

func createSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	domain := resourceData.Get(domainKey).(string)
	resourceData.SetId(resourceIdFromDomain(domain))

	return updateSiteOwners(resourceData, provider)
}

func updateSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
	domain := resourceData.Get(domainKey).(string)

	owners := resourceData.Get(ownersKey).(*schema.Set)
	ownersList := make([]string, 0, owners.Len())
	for _, owner := range owners.List() {
		ownersList = append(ownersList, owner.(string))
	}

	siteResource, getErr := service.WebResource.Get(resourceData.Id()).Do()
	if getErr != nil {
		return getErr
	}
	siteResource.Owners = ownersList

	return resource.Retry(resourceData.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		updatedSiteResource, updateErr := service.WebResource.Update(resourceData.Id(), siteResource).Do()
		if updateErr != nil {
			return resource.RetryableError(updateErr)
		}

		resourceData.SetId(domain)
		_ = resourceData.Set(ownersKey, updatedSiteResource.Owners)

		return resource.NonRetryableError(readDnsSiteVerification(resourceData, provider))
	})
}
