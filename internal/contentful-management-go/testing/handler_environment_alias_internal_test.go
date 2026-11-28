package cmtesting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/ogen-go/ogen/middleware"
	"github.com/stretchr/testify/require"
)

func TestAppRequestRetainsTargetSelectedBeforeDispatch(t *testing.T) {
	t.Parallel()

	server, err := NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppInstallation("space", "blue", "app", cm.AppInstallationData{})
	server.SetTag("space", "blue", "tag", cm.TagRequest{Name: "Blue"})
	server.SetTag("space", "green", "tag", cm.TagRequest{Name: "Green"})

	setAlias := func(target string) {
		server.SetEnvironmentAlias(NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink(target)}))
	}
	setAlias("blue")

	handler := server.Handler()
	handler.appAccessTokens["test-app-token"] = issuedAppAccessToken{
		installation: cm.CreateAppAccessTokenParams{SpaceID: "space", EnvironmentID: "blue", AppDefinitionID: "app"},
		expiresAt:    time.Now().Add(time.Minute),
	}

	retarget := func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
		setAlias("green")

		return next(req)
	}
	api, err := cm.NewServer(handler, handler, cm.WithMiddleware(handler.authorizeAppRequest, retarget))
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/spaces/space/environments/alias/tags/tag", nil)
	req.Header.Set("Authorization", "Bearer test-app-token")

	response := httptest.NewRecorder()
	api.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)

	var tag cm.Tag

	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &tag))
	require.Equal(t, "Blue", tag.Name)
	require.Equal(t, "alias", tag.Sys.Environment.Sys.ID)

	// A subsequent direct Handler call resolves the alias at the time of that read.
	next, err := handler.GetTag(t.Context(), cm.GetTagParams{SpaceID: "space", EnvironmentID: "alias", TagID: "tag"})
	require.NoError(t, err)

	current, ok := next.(*cm.Tag)
	require.True(t, ok)
	require.Equal(t, "Green", current.Name)
	require.Equal(t, "alias", current.Sys.Environment.Sys.ID)
}
