package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

type aliasTestResponse struct {
	Sys struct {
		ID          string             `json:"id"`
		Environment cm.EnvironmentLink `json:"environment"`
		Version     int                `json:"version"`
	} `json:"sys"`
	Environment cm.EnvironmentLink  `json:"environment"`
	Parameters  json.RawMessage     `json:"parameters"`
	Name        string              `json:"name"`
	Items       []aliasTestResponse `json:"items"`
}

func aliasRequest(t *testing.T, server *cmt.Server, method, path, version, body string, status int) aliasTestResponse {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, "/spaces/"+path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	if body != "" {
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	}

	if version != "" {
		req.Header.Set("X-Contentful-Version", version)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	require.Equal(t, status, rec.Code, "%s %s: %s", method, path, rec.Body.String())

	var result aliasTestResponse
	if rec.Code != http.StatusNoContent {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	}

	return result
}

func TestEnvironmentAliasChildResponses(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "target")
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("target")}))
	server.SetEntry("space", "target", "article", "entry", cm.EntryRequest{})
	server.SetLocale(cmt.NewLocaleFromData("space", "target", "default", cm.LocaleData{Name: "English", Code: "en-US"}, true))
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "Test"})

	const (
		alias  = "space/environments/alias/"
		target = "space/environments/target/"
	)

	put := func(path, version, body string, status int) {
		t.Helper()

		r := aliasRequest(t, server, http.MethodPut, alias+path, version, body, status)
		require.Equal(t, cm.NewEnvironmentLink("alias"), r.Sys.Environment, path)
	}

	put("content_types/article", "", testJSON(map[string]any{
		"name": "Article", "displayField": "title",
		"fields": []any{map[string]any{"id": "title", "name": "Title", "type": "Symbol"}},
	}), http.StatusCreated)
	put("content_types/article/published", "1", "", http.StatusOK)
	put("content_types/article/editor_interface", "1", testJSON(map[string]any{"controls": []any{}}), http.StatusOK)
	put("entries/entry/published", "1", "", http.StatusOK)
	put("extensions/extension", "0", testJSON(map[string]any{"extension": map[string]any{
		"name": "Example", "srcdoc": "sample", "fieldTypes": []any{}, "sidebar": false,
	}}), http.StatusCreated)
	put("tags/tag", "", testJSON(map[string]any{"name": "Tag", "sys": map[string]any{"type": "Tag"}}), http.StatusCreated)

	parameters := map[string]string{"environment": "target"}
	put("app_installations/app", "", testJSON(map[string]any{"parameters": parameters}), http.StatusOK)
	put("live_preview/variables", "0", testJSON(map[string]any{"variables": map[string]any{}}), http.StatusOK)

	installation, ok := server.StoredAppInstallation("space", "target", "app")
	require.True(t, ok)
	require.Equal(t, cm.NewEnvironmentLink("target"), installation.Sys.Environment)
	require.JSONEq(t, testJSON(parameters), string(installation.Parameters))

	viaAlias := aliasRequest(t, server, http.MethodGet, alias+"app_installations/app", "", "", http.StatusOK)
	require.JSONEq(t, testJSON(parameters), string(viaAlias.Parameters))

	paths := []string{
		"entries/entry", "content_types/article", "content_types/article/editor_interface",
		"extensions/extension", "tags/tag", "app_installations/app", "live_preview/variables", "locales/default",
	}
	failure := aliasRequest(t, server, http.MethodDelete, "space/environments/target", "", "", http.StatusUnprocessableEntity)
	require.Equal(t, "ValidationFailed", failure.Sys.ID)

	for _, path := range paths {
		viaAlias := aliasRequest(t, server, http.MethodGet, alias+path, "", "", http.StatusOK)
		direct := aliasRequest(t, server, http.MethodGet, target+path, "", "", http.StatusOK)
		require.Equal(t, cm.NewEnvironmentLink("alias"), viaAlias.Sys.Environment, path)
		require.Equal(t, cm.NewEnvironmentLink("target"), direct.Sys.Environment, path)
		require.Equal(t, direct.Sys.ID, viaAlias.Sys.ID, path)
		require.Equal(t, direct.Sys.Version, viaAlias.Sys.Version, path)
	}

	for _, collection := range []string{"entries", "content_types", "locales"} {
		r := aliasRequest(t, server, http.MethodGet, alias+collection, "", "", http.StatusOK)
		require.Len(t, r.Items, 1, collection)
		require.Equal(t, cm.NewEnvironmentLink("alias"), r.Items[0].Sys.Environment, collection)
	}

	aliasRequest(t, server, http.MethodDelete, "space/environment_aliases/alias", "", "", http.StatusNoContent)
	aliasRequest(t, server, http.MethodDelete, "space/environments/target", "", "", http.StatusNoContent)
	server.RegisterSpaceEnvironment("space", "target")

	for _, path := range paths {
		aliasRequest(t, server, http.MethodGet, target+path, "", "", http.StatusNotFound)
	}
}

func TestEnvironmentAliasRetargetUsesCurrentTargetVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetTag("space", "blue", "tag", cm.TagRequest{Name: "Blue"})
	server.SetTag("space", "green", "tag", cm.TagRequest{Name: "Green"})
	server.SetTag("other", "blue", "tag", cm.TagRequest{Name: "Other space"})

	for _, id := range []string{"alias", "second"} {
		server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", id, cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("blue")}))
	}

	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("other", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("blue")}))

	const tag = "space/environments/alias/tags/tag"

	update := testJSON(map[string]any{"name": "Updated", "sys": map[string]any{"type": "Tag"}})

	before := aliasRequest(t, server, http.MethodPut, tag, "1", update, http.StatusOK)
	require.Equal(t, 2, before.Sys.Version)
	aliasRequest(t, server, http.MethodPut, "space/environment_aliases/alias", "1", testJSON(map[string]any{"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "green"}}}), http.StatusOK)
	aliasRequest(t, server, http.MethodPut, tag, "2", update, http.StatusConflict)
	after := aliasRequest(t, server, http.MethodGet, tag, "", "", http.StatusOK)
	require.Equal(t, "Green", after.Name)
	require.Equal(t, 1, after.Sys.Version)
	require.Equal(t, cm.NewEnvironmentLink("alias"), after.Sys.Environment)
	require.Equal(t, "Updated", aliasRequest(t, server, http.MethodGet, "space/environments/blue/tags/tag", "", "", http.StatusOK).Name)

	aliasRequest(t, server, http.MethodDelete, "space/environments/blue", "", "", http.StatusUnprocessableEntity)
	aliasRequest(t, server, http.MethodDelete, "space/environment_aliases/second", "", "", http.StatusNoContent)
	aliasRequest(t, server, http.MethodDelete, "space/environments/blue", "", "", http.StatusNoContent)
	require.Equal(t, "Green", aliasRequest(t, server, http.MethodGet, tag, "", "", http.StatusOK).Name)
	require.Equal(t, "Other space", aliasRequest(t, server, http.MethodGet, "other/environments/alias/tags/tag", "", "", http.StatusOK).Name)
	aliasRequest(t, server, http.MethodDelete, tag, "", "", http.StatusNoContent)
	aliasRequest(t, server, http.MethodGet, "space/environments/green/tags/tag", "", "", http.StatusNotFound)
}

func TestMasterEnvironmentDeletion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	const (
		environment = "space/environments/master"
		alias       = "space/environment_aliases/master"
	)

	aliasRequest(t, server, http.MethodDelete, environment, "", "", http.StatusNotFound)
	aliasRequest(t, server, http.MethodDelete, alias, "", "", http.StatusNotFound)
	server.RegisterSpaceEnvironment("space", "master")
	aliasRequest(t, server, http.MethodDelete, environment, "", "", http.StatusUnprocessableEntity)
	aliasRequest(t, server, http.MethodGet, environment, "", "", http.StatusOK)
	aliasRequest(t, server, http.MethodDelete, alias, "", "", http.StatusNotFound)
}

func TestMasterEnvironmentAliasDeletion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterMasterEnvironmentAlias("space")

	aliasRequest(t, server, http.MethodDelete, "space/environment_aliases/master", "", "", http.StatusUnprocessableEntity)
	aliasRequest(t, server, http.MethodDelete, "space/environments/master", "", "", http.StatusUnprocessableEntity)
	aliasRequest(t, server, http.MethodDelete, "space/environments/master-1970-01-01", "", "", http.StatusUnprocessableEntity)
	aliasRequest(t, server, http.MethodGet, "space/environment_aliases/master", "", "", http.StatusOK)
	listed := aliasRequest(t, server, http.MethodGet, "space/environments", "", "", http.StatusOK)
	require.Len(t, listed.Items, 1)
	require.Equal(t, "master-1970-01-01", listed.Items[0].Sys.ID)
}

func TestSpaceParentRequiresConcreteMasterOrLocalAliasTarget(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetSpace(cm.Space{Sys: cm.SpaceSys{ID: "space"}})
	server.RegisterSpaceEnvironment("space", "unrelated")
	server.SetSpaceEnablements("space", cm.SpaceEnablementData{})
	aliasRequest(t, server, http.MethodGet, "space/roles", "", "", http.StatusNotFound)

	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "master", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("blue")}))
	server.RegisterSpaceEnvironment("other", "blue")
	aliasRequest(t, server, http.MethodGet, "space/roles", "", "", http.StatusNotFound)
	aliasRequest(t, server, http.MethodGet, "space/environments", "", "", http.StatusNotFound)

	server.RegisterSpaceEnvironment("space", "blue")
	aliasRequest(t, server, http.MethodGet, "space/roles", "", "", http.StatusOK)
	created := aliasRequest(t, server, http.MethodPut, "space/environments/sandbox", "", testJSON(map[string]string{"name": "Sandbox"}), http.StatusCreated)
	require.Equal(t, "sandbox", created.Sys.ID)
}

func TestEnvironmentAndAliasIDsCannotCollide(t *testing.T) {
	t.Parallel()

	aliasBody := testJSON(cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("master-1970-01-01")})

	for _, test := range []struct {
		name, path, version, body string
	}{
		{name: "create environment at alias ID", path: "space/environments/master", body: testJSON(map[string]string{"name": "New"})},
		{name: "update environment at alias ID", path: "space/environments/active", version: "1", body: testJSON(map[string]string{"name": "Renamed"})},
		{name: "create alias at environment ID", path: "space/environment_aliases/staging", version: "0", body: aliasBody},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterMasterEnvironmentAlias("space")
			server.RegisterSpaceEnvironment("space", "staging")
			server.RegisterSpaceEnvironment("other", "active")
			aliasRequest(t, server, http.MethodPut, "space/environment_aliases/active", "0", aliasBody, http.StatusCreated)

			environments := aliasRequest(t, server, http.MethodGet, "space/environments", "", "", http.StatusOK)
			aliases := aliasRequest(t, server, http.MethodGet, "space/environment_aliases", "", "", http.StatusOK)
			conflict := aliasRequest(t, server, http.MethodPut, test.path, test.version, test.body, http.StatusConflict)
			require.Equal(t, "Conflict", conflict.Sys.ID)
			require.Equal(t, environments.Items, aliasRequest(t, server, http.MethodGet, "space/environments", "", "", http.StatusOK).Items)
			require.Equal(t, aliases.Items, aliasRequest(t, server, http.MethodGet, "space/environment_aliases", "", "", http.StatusOK).Items)
		})
	}
}
