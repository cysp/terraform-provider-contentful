package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppKeyResourceModelFromResponse(appKey cm.AppKey) AppKeyModel {
	organizationID := appKey.Sys.Organization.Sys.ID
	appDefinitionID := appKey.Sys.AppDefinition.Sys.ID
	keyKID := appKey.Sys.ID

	model := AppKeyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, appDefinitionID, keyKID),
		AppKeyIdentityModel: AppKeyIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
			KeyKID:          types.StringValue(keyKID),
		},
		JWK: NewTypedObject(NewAppKeyJWKModelFromResponse(appKey.Jwk)),
	}

	model.CreatedAt = timetypes.NewRFC3339TimePointerValue(appKey.Sys.CreatedAt.ValueTimePointer())
	model.UpdatedAt = timetypes.NewRFC3339TimePointerValue(appKey.Sys.UpdatedAt.ValueTimePointer())
	model.LastUsedAt = timetypes.NewRFC3339TimePointerValue(appKey.Sys.LastUsedAt.ValueTimePointer())

	return model
}

func NewAppKeyJWKModelFromResponse(jwk cm.AppKeyJWK) AppKeyJWKModel {
	x5cElements := make([]types.String, 0, len(jwk.X5c))
	for _, publicKey := range jwk.X5c {
		x5cElements = append(x5cElements, types.StringValue(publicKey))
	}

	return AppKeyJWKModel{
		Alg: types.StringValue(string(jwk.Alg)),
		Kty: types.StringValue(string(jwk.Kty)),
		Use: types.StringValue(string(jwk.Use)),
		X5c: NewTypedList(x5cElements),
		Kid: types.StringValue(jwk.Kid),
		X5t: types.StringValue(jwk.X5t),
	}
}
