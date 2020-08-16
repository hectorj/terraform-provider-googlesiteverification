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

	setOwnersErr := setOwners(resourceData, siteResource.Owners)
	return setOwnersErr
}

func createSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	domain := resourceData.Get(domainKey).(string)
	resourceData.SetId(resourceIdFromDomain(domain))

	return updateSiteOwners(resourceData, provider)
}

func updateSiteOwners(resourceData *schema.ResourceData, provider interface{}) error {
	service := provider.(configuredProvider).service
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

		setOwnersErr := setOwners(resourceData, updatedSiteResource.Owners)
		if setOwnersErr != nil {
			resource.NonRetryableError(setOwnersErr)
		}

		return resource.NonRetryableError(readDnsSiteVerification(resourceData, provider))
	})
}

func setOwners(resourceData *schema.ResourceData, owners []string) error {
	ownersSet := schema.NewSet(schema.HashString, nil)
	for _, owner := range owners {
		ownersSet.Add(owner)
	}
	return resourceData.Set(ownersKey, ownersSet)
}
