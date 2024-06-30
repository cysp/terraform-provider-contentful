package integration_tests_test

import (
	"context"
	"os"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContentfulManagementClient(t *testing.T) contentfulManagement.Client {
	t.Helper()

	client, err := contentfulManagement.NewClient(
		contentfulManagement.DefaultServerURL,
		contentfulManagement.NewAccessTokenSecuritySource("CFPAT-12345"),
		contentfulManagement.WithUserAgent(contentfulManagement.DefaultUserAgent),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	return *client
}

func testAuthorizedContentfulManagementClient(t *testing.T) contentfulManagement.Client {
	t.Helper()

	accessToken := os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if accessToken == "" {
		t.Skip("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN is not set")
	}

	client, err := contentfulManagement.NewClient(
		contentfulManagement.DefaultServerURL,
		contentfulManagement.NewAccessTokenSecuritySource(accessToken),
		contentfulManagement.WithUserAgent(contentfulManagement.DefaultUserAgent),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	return *client
}

func TestGetAuthenticatedUserUnauthorized(t *testing.T) {
	t.Parallel()

	client := testContentfulManagementClient(t)

	response, err := client.GetAuthenticatedUser(context.Background())
	require.NoError(t, err)

	switch response := response.(type) {
	case *contentfulManagement.Error:
		require.NotNil(t, response)
		assert.EqualValues(t, "AccessTokenInvalid", response.Sys.ID)
	default:
		t.Fatal("unexpected type")
	}
}

func TestGetAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	client := testAuthorizedContentfulManagementClient(t)

	response, err := client.GetAuthenticatedUser(context.Background())
	require.NoError(t, err)

	switch response := response.(type) {
	case *contentfulManagement.User:
		require.NotNil(t, response)
	default:
		t.Fatal("unexpected type")
	}
}
