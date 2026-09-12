package integration_tests_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthenticatedUserUnauthorized(t *testing.T) {
	t.Parallel()

	_, testServer := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testServer.URL, "CFPAT-00000")

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	errorResponse, ok := response.(cm.ErrorResponse)
	require.True(t, ok, "expected an ErrorResponse, got %T", response)
	responseError, ok := errorResponse.GetError()
	require.True(t, ok)
	assert.Equal(t, "AccessTokenInvalid", responseError.Sys.ID)
}

func TestGetAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	server, testServer := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)

	server.SetMe(cm.NewUser("123"))

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	user, ok := response.(*cm.User)
	require.True(t, ok, "expected a User, got %T", response)
	require.NotNil(t, user)
	assert.Equal(t, "123", user.Sys.ID)
}
