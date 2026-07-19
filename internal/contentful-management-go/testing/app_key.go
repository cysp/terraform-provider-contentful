package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const appKeyMockUserID = "mock-user"

func NewAppKeyFromRequest(organizationID, appDefinitionID string, request cm.AppKeyRequestData) cm.AppKey {
	return cm.AppKey{
		Sys: cm.NewAppKeySys(organizationID, appDefinitionID, request.Jwk.Kid, appKeyMockUserID),
		Jwk: request.Jwk,
	}
}
