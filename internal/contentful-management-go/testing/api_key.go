package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAPIKeyFromRequestFields(spaceID, apiKeyID string, apiKeyFields cm.ApiKeyRequestData) cm.ApiKey {
	apiKey := cm.ApiKey{
		Sys: cm.NewAPIKeySys(spaceID, apiKeyID),
	}

	UpdateAPIKeyFromRequestFields(&apiKey, apiKeyFields)

	return apiKey
}

func UpdateAPIKeyFromRequestFields(apiKey *cm.ApiKey, apiKeyFields cm.ApiKeyRequestData) {
	apiKey.Name = apiKeyFields.Name
	apiKey.Description = apiKeyFields.Description
	apiKey.Environments = apiKeyFields.Environments
}
