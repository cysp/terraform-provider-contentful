package cmtesting_test

import (
	"context"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePersonalAccessTokenReturnsBadRequestForInvalidRequest(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	response, err := handler.CreatePersonalAccessToken(context.Background(), &cm.PersonalAccessTokenRequestData{})

	require.NoError(t, err)

	errorStatusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, errorStatusCode.StatusCode)
	errResponse, ok := errorStatusCode.Response.GetError()
	require.True(t, ok)
	assert.Equal(t, "BadRequest", errResponse.Sys.ID)
}

func TestPutTeamReturnsVersionMismatchForStaleVersion(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	_, err := handler.PutTeam(context.Background(), &cm.TeamData{
		Name:        "Test Team",
		Description: cm.NewNilString(""),
	}, cm.PutTeamParams{
		OrganizationID:     "organization-id",
		TeamID:             "team-id",
		XContentfulVersion: 0,
	})
	require.NoError(t, err)

	response, err := handler.PutTeam(context.Background(), &cm.TeamData{
		Name:        "Updated Test Team",
		Description: cm.NewNilString(""),
	}, cm.PutTeamParams{
		OrganizationID:     "organization-id",
		TeamID:             "team-id",
		XContentfulVersion: 1,
	})

	require.NoError(t, err)

	errorStatusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusConflict, errorStatusCode.StatusCode)
	errResponse, ok := errorStatusCode.Response.GetError()
	require.True(t, ok)
	assert.Equal(t, cm.ErrorSysIDVersionMismatch, errResponse.Sys.ID)
}
