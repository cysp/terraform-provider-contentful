package testing

import (
	"context"
	"net/http"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreatePersonalAccessToken(_ context.Context, req *cm.PersonalAccessTokenRequestFields) (cm.CreatePersonalAccessTokenRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	err := ValidatePersonalAccessTokenRequestFields(req)
	if err != nil {
		return NewContentfulManagementErrorStatusCodeBadRequest(pointerTo(err.Error()), nil), nil
	}

	personalAccessTokenID := generateResourceID()

	personalAccessToken := NewPersonalAccessTokenFromRequestFields(personalAccessTokenID, *req)
	ts.personalAccessTokens[personalAccessToken.Sys.ID] = &personalAccessToken

	personalAccessTokenWithToken := personalAccessToken
	personalAccessTokenWithToken.Token.SetTo(generateResourceID())

	return &cm.PersonalAccessTokenStatusCode{
		StatusCode: http.StatusCreated,
		Response:   personalAccessTokenWithToken,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetPersonalAccessToken(_ context.Context, params cm.GetPersonalAccessTokenParams) (cm.GetPersonalAccessTokenRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	personalAccessToken := ts.personalAccessTokens[params.AccessTokenID]
	if personalAccessToken == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("PersonalAccessToken not found"), nil), nil
	}

	return personalAccessToken, nil
}

//nolint:ireturn
func (ts *Handler) RevokePersonalAccessToken(_ context.Context, params cm.RevokePersonalAccessTokenParams) (cm.RevokePersonalAccessTokenRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	personalAccessToken := ts.personalAccessTokens[params.AccessTokenID]
	if personalAccessToken == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("PersonalAccessToken not found"), nil), nil
	}

	personalAccessToken.RevokedAt.SetTo(time.Now())

	return &cm.PersonalAccessTokenStatusCode{
		StatusCode: http.StatusOK,
		Response:   *personalAccessToken,
	}, nil
}
