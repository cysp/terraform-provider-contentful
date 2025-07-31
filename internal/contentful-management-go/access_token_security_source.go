package contentfulmanagement

import (
	"context"
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
