package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAPIKeyFromRequestFields(spaceID, apiKeyID string, apiKeyFields cm.ApiKeyRequestData) cm.ApiKey {
	apiKey := cm.ApiKey{
		Sys: NewAPIKeySys(spaceID, apiKeyID),
	}

	UpdateAPIKeyFromRequestFields(&apiKey, apiKeyFields)

	return apiKey
}

func NewAPIKeySys(spaceID string, apiKeyID string) cm.ApiKeySys {
	return cm.ApiKeySys{
		Type: cm.ApiKeySysTypeApiKey,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		ID: apiKeyID,
	}
}

func UpdateAPIKeyFromRequestFields(apiKey *cm.ApiKey, apiKeyFields cm.ApiKeyRequestData) {
	apiKey.Name = apiKeyFields.Name
	apiKey.Description = apiKeyFields.Description
	apiKey.Environments = apiKeyFields.Environments
}
