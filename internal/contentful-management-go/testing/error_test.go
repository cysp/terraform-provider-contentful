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
	assert.Equal(t, cm.ErrorSysTypeError, errResponse.Sys.Type)
	assert.NotEmpty(t, errResponse.Message.Or(""))
}

func TestErrorHelpersAlwaysReturnMessages(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response *cm.ErrorStatusCode
		status   int
		id       string
	}{
		"not found": {
			response: cmt.NewContentfulManagementErrorStatusCodeNotFound(nil, nil),
			status:   http.StatusNotFound,
			id:       cm.ErrorSysIDNotFound,
		},
		"version mismatch": {
			response: cmt.NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil),
			status:   http.StatusConflict,
			id:       cm.ErrorSysIDVersionMismatch,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.status, test.response.StatusCode)
			errorResponse, ok := test.response.Response.GetError()
			require.True(t, ok)
			assert.Equal(t, cm.ErrorSysTypeError, errorResponse.Sys.Type)
			assert.Equal(t, test.id, errorResponse.Sys.ID)
			assert.NotEmpty(t, errorResponse.Message.Or(""))
		})
	}
}
