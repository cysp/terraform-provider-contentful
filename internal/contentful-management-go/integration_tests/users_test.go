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

	_, testserver := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testserver.URL, "CFPAT-00000")

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	switch response := response.(type) {
	case cm.ErrorResponse:
		require.NotNil(t, response)

		responseError, responseErrorOk := response.GetError()
		require.True(t, responseErrorOk)
		assert.Equal(t, "AccessTokenInvalid", responseError.Sys.ID)
	default:
		t.Fatal("unexpected type")
	}
}

func TestGetAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	server, testserver := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	server.SetMe(cm.NewUser("123"))

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	switch response := response.(type) {
	case *cm.User:
		require.NotNil(t, response)
	default:
		t.Fatal("unexpected type")
	}
}
