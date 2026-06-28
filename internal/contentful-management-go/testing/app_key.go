package cmtesting

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppKeyFromRequest(organizationID, appDefinitionID string, request cm.AppKeyRequestData) cm.AppKey {
	jwk, ok := request.Jwk.Get()
	if !ok {
		jwk = cm.AppKeyJWK{
			Alg: cm.AppKeyJWKAlgRS256,
			Kty: cm.AppKeyJWKKtyRSA,
			Use: cm.AppKeyJWKUseSig,
			X5c: []string{"generated-certificate"},
			Kid: "generated-key",
			X5t: "generated-key",
		}
	}

	appKey := cm.AppKey{
		Sys: cm.NewAppKeySys(organizationID, appDefinitionID, jwk.Kid),
		Jwk: jwk,
	}

	now := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	appKey.Sys.CreatedAt = cm.NewOptDateTime(now)
	appKey.Sys.UpdatedAt = cm.NewOptDateTime(now)

	if generate, ok := request.Generate.Get(); ok && generate {
		appKey.PrivateKey = cm.NewOptString("generated-private-key")
	}

	return appKey
}
