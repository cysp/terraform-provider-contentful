package integration_tests_test

import (
	"net/http"
	"os"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContentfulManagementClient(t *testing.T) cm.Client {
	t.Helper()

	client, err := cm.NewClient(
		cm.DefaultServerURL,
		cm.NewAccessTokenSecuritySource("CFPAT-12345"),
		cm.WithClient(cm.NewClientWithUserAgent(http.DefaultClient, cm.DefaultUserAgent)),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	return *client
}

func testAuthorizedContentfulManagementClient(t *testing.T) cm.Client {
	t.Helper()

	accessToken := os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if accessToken == "" {
		t.Skip("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN is not set")
	}

	client, err := cm.NewClient(
		cm.DefaultServerURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(cm.NewClientWithUserAgent(http.DefaultClient, cm.DefaultUserAgent)),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	return *client
}

func TestGetAuthenticatedUserUnauthorized(t *testing.T) {
	t.Parallel()

	client := testContentfulManagementClient(t)

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	switch response := response.(type) {
	case *cm.Error:
		require.NotNil(t, response)
		assert.EqualValues(t, "AccessTokenInvalid", response.Sys.ID)
	default:
		t.Fatal("unexpected type")
	}
}

func TestGetAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	client := testAuthorizedContentfulManagementClient(t)

	response, err := client.GetAuthenticatedUser(t.Context())
	require.NoError(t, err)

	switch response := response.(type) {
	case *cm.User:
		require.NotNil(t, response)
	default:
		t.Fatal("unexpected type")
	}
}
