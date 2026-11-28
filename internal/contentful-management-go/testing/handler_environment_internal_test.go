package cmtesting

import (
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func TestDeleteEnvironmentDiscardsContentTypePublicationHistory(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	handler.RegisterSpaceEnvironment("space", "blue", "ready")

	for _, scope := range [][2]string{{"space", "blue"}, {"space", "green"}, {"other", "blue"}} {
		handler.contentTypePublications.Set(scope[0], scope[1], "article", &contentTypePublicationSnapshot{"title": false})
	}

	deleted, err := handler.DeleteEnvironment(t.Context(), cm.DeleteEnvironmentParams{SpaceID: "space", EnvironmentID: "blue"})
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deleted)
	require.Nil(t, handler.contentTypePublications.Get("space", "blue", "article"))
	require.NotNil(t, handler.contentTypePublications.Get("space", "green", "article"))
	require.NotNil(t, handler.contentTypePublications.Get("other", "blue", "article"))
}

func TestDeleteEnvironmentDiscardsOnlyItsIssuedTokens(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	handler.RegisterSpaceEnvironment("space", "blue", "ready")

	for name, installation := range map[string]cm.CreateAppAccessTokenParams{
		"first":             {SpaceID: "space", EnvironmentID: "blue", AppDefinitionID: "a"},
		"second":            {SpaceID: "space", EnvironmentID: "blue", AppDefinitionID: "b"},
		"other environment": {SpaceID: "space", EnvironmentID: "green", AppDefinitionID: "a"},
		"other space":       {SpaceID: "other", EnvironmentID: "blue", AppDefinitionID: "a"},
	} {
		handler.appAccessTokens[name] = issuedAppAccessToken{installation: installation, expiresAt: time.Now().Add(time.Minute)}
	}

	alias := NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("blue")})
	handler.environmentAliases.Set("space", "alias", &alias)

	params := cm.DeleteEnvironmentParams{SpaceID: "space", EnvironmentID: "blue"}
	blocked, err := handler.DeleteEnvironment(t.Context(), params)
	require.NoError(t, err)
	require.IsType(t, &cm.ErrorStatusCode{}, blocked)
	require.Len(t, handler.appAccessTokens, 4)

	handler.environmentAliases.Delete("space", "alias")
	deleted, err := handler.DeleteEnvironment(t.Context(), params)
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deleted)
	require.Len(t, handler.appAccessTokens, 2)
	require.Contains(t, handler.appAccessTokens, "other environment")
	require.Contains(t, handler.appAccessTokens, "other space")
}
