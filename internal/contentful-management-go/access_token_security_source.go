package contentfulmanagement

import (
	"context"
	"errors"
)

type AccessTokenSecuritySource struct {
	accessToken string
}

var _ SecuritySource = (*AccessTokenSecuritySource)(nil)

func NewAccessTokenSecuritySource(accessToken string) AccessTokenSecuritySource {
	return AccessTokenSecuritySource{
		accessToken: accessToken,
	}
}

func (c AccessTokenSecuritySource) AccessToken(_ context.Context, _ string, _ *Client) (AccessToken, error) {
	return AccessToken{Token: c.accessToken}, nil
}

var errMissingAppJWT = errors.New("an app JWT is required to issue an app access token")

type appJWTContextKey struct{}

// WithAppJWT supplies the signed JWT used by AccessTokenSecuritySource for CreateAppAccessToken.
func WithAppJWT(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, appJWTContextKey{}, AppJWT{Token: token})
}

func (AccessTokenSecuritySource) AppJWT(ctx context.Context, _ string, _ *Client) (AppJWT, error) {
	jwt, ok := ctx.Value(appJWTContextKey{}).(AppJWT)
	if !ok || jwt.Token == "" {
		return AppJWT{}, errMissingAppJWT
	}

	return jwt, nil
}
