package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleReadRequiresParentEnvironment(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetLocale(cm.Locale{
		Sys:  cm.LocaleSys{ID: "locale", Type: cm.LocaleSysTypeLocale, Space: cm.NewSpaceLink("space"), Environment: cm.NewEnvironmentLink("environment")},
		Name: "English", Code: "en-US", Default: true, ContentManagementApi: true, ContentDeliveryApi: true,
	})
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("environment")}))

	for _, environmentID := range []string{"environment", "alias"} {
		response, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: environmentID, LocaleID: "locale"})
		require.NoError(t, err)

		failure, ok := response.(*cm.ErrorStatusCode)
		require.True(t, ok, "orphan fixture must not be readable through %s", environmentID)
		assert.Equal(t, http.StatusNotFound, failure.StatusCode)
	}

	server.RegisterSpaceEnvironment("space", "environment")

	for _, environmentID := range []string{"environment", "alias"} {
		response, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: environmentID, LocaleID: "locale"})
		require.NoError(t, err)

		locale, ok := response.(*cm.Locale)
		require.True(t, ok)
		assert.Equal(t, environmentID, locale.Sys.Environment.Sys.ID)
	}
}

func TestLocaleHTTPZeroLimitUsesDefault(t *testing.T) {
	t.Parallel()

	server := newLocaleValidationServer(t)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/spaces/space/environments/environment/locales?limit=0", nil)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	var collection cm.LocaleCollection
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &collection))
	assert.Equal(t, cm.NewOptInt(100), collection.Limit)
	assert.Len(t, collection.Items, 4)
}
