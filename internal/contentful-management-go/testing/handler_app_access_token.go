package cmtesting

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"maps"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/golang-jwt/jwt/v5"
)

//nolint:ireturn
func (h *Handler) CreateAppAccessToken(ctx context.Context, params cm.CreateAppAccessTokenParams) (cm.CreateAppAccessTokenRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	raw, _ := ctx.Value(appJWTContextKey{}).(string)

	definition := h.appDefinitions[params.AppDefinitionID]
	if definition == nil || !validAppJWT(raw, params.AppDefinitionID, h.appKeys.List(definition.Sys.Organization.Sys.ID, params.AppDefinitionID)) {
		return nil, ErrAccessTokenInvalid
	}

	if h.environments.Get(params.SpaceID, params.EnvironmentID) == nil || h.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	const lifetime = 10 * time.Minute

	now := time.Now().UTC().Truncate(time.Second)

	maps.DeleteFunc(h.appAccessTokens, func(_ string, token issuedAppAccessToken) bool { return !now.Before(token.expiresAt) })

	token := rand.Text()
	expiresAt := now.Add(lifetime)
	h.appAccessTokens[token] = issuedAppAccessToken{installation: params, expiresAt: expiresAt}

	return &cm.CreateAppAccessTokenCreated{Token: token, Sys: cm.AppAccessTokenSys{ExpiresAt: expiresAt}}, nil
}

func validAppJWT(raw, appDefinitionID string, appKeys []cm.AppKey) bool {
	const maxLifetime = 10 * time.Minute

	var claims jwt.RegisteredClaims

	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		var candidates jwt.VerificationKeySet

		for _, key := range appKeys {
			if kid, ok := token.Header["kid"]; (ok && kid != key.Jwk.Kid) || len(key.Jwk.X5c) == 0 {
				continue
			}

			der, err := base64.StdEncoding.DecodeString(key.Jwk.X5c[0])
			if err != nil {
				continue
			}

			publicKey, err := x509.ParsePKIXPublicKey(der)
			if rsaKey, ok := publicKey.(*rsa.PublicKey); err == nil && ok {
				candidates.Keys = append(candidates.Keys, rsaKey)
			}
		}

		return candidates, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(appDefinitionID), jwt.WithExpirationRequired(), jwt.WithIssuedAt())

	if err != nil || token.Header["typ"] != "JWT" || claims.IssuedAt == nil {
		return false
	}

	return claims.ExpiresAt.Sub(claims.IssuedAt.Time) <= maxLifetime
}
