package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAPIKeyFromRequestFields(spaceID, apiKeyID string, apiKeyFields cm.ApiKeyRequestData) cm.ApiKey {
	apiKey := cm.ApiKey{
		Sys: cm.NewAPIKeySys(spaceID, apiKeyID),
	}

	setAPIKeyFromRequestFields(&apiKey, apiKeyFields)

	return apiKey
}

func UpdateAPIKeyFromRequestFields(apiKey *cm.ApiKey, apiKeyFields cm.ApiKeyRequestData) {
	apiKey.Sys.Version++
	setAPIKeyFromRequestFields(apiKey, apiKeyFields)
}

func setAPIKeyFromRequestFields(apiKey *cm.ApiKey, apiKeyFields cm.ApiKeyRequestData) {
	apiKey.Name = apiKeyFields.Name
	apiKey.Description = apiKeyFields.Description
	apiKey.Environments = apiKeyFields.Environments
}
