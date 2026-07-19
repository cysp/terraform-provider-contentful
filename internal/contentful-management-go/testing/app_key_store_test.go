package cmtesting_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppKeyStoreSetReplacesOnlyMatchingIdentity(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "First App"})
	server.SetAppDefinition("other-organization", "other-app-definition", cm.AppDefinitionData{Name: "Second App"})

	server.SetAppKey("organization", "app-definition", testAppKeyStoreRequest("first"))
	server.SetAppKey("other-organization", "other-app-definition", testAppKeyStoreRequest("second"))

	first := getStoredAppKey(t, server, "organization", "app-definition")
	assert.Equal(t, []string{"first"}, first.Jwk.X5c)

	second := getStoredAppKey(t, server, "other-organization", "other-app-definition")
	assert.Equal(t, []string{"second"}, second.Jwk.X5c)

	server.SetAppKey("organization", "app-definition", testAppKeyStoreRequest("replacement"))

	replaced := getStoredAppKey(t, server, "organization", "app-definition")
	assert.Equal(t, []string{"replacement"}, replaced.Jwk.X5c)

	unchanged := getStoredAppKey(t, server, "other-organization", "other-app-definition")
	assert.Equal(t, []string{"second"}, unchanged.Jwk.X5c)
}

func testAppKeyStoreRequest(x5c string) cm.AppKeyRequestData {
	return cm.NewAppKeyRequestData(cm.AppKeyJWK{
		Kid: "key",
		X5c: []string{x5c},
	})
}

func getStoredAppKey(t *testing.T, server *cmt.Server, organizationID, appDefinitionID string) *cm.AppKey {
	t.Helper()

	response, err := server.Handler().GetAppKey(context.Background(), cm.GetAppKeyParams{
		OrganizationID:  organizationID,
		AppDefinitionID: appDefinitionID,
		KeyKid:          "key",
	})
	require.NoError(t, err)

	appKey, ok := response.(*cm.AppKey)
	require.True(t, ok)

	return appKey
}
