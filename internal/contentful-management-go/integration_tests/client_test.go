package integration_tests_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func testContentfulManagementClient(t *testing.T, serverURL string, accessToken string) cm.Client {
	t.Helper()

	client, err := cm.NewClient(
		serverURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(cm.NewClientWithUserAgent(http.DefaultClient, cm.DefaultUserAgent)),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	return *client
}
