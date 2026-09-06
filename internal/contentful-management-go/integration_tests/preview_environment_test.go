package integration_tests_test

import (
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func previewEnvironmentData(name string, configurations ...cm.PreviewEnvironmentConfigurationData) cm.PreviewEnvironmentData {
	return cm.PreviewEnvironmentData{
		Name:           name,
		Description:    "description",
		Configurations: configurations,
	}
}

func previewEnvironmentConfiguration(contentTypeID, url string, enabled bool) cm.PreviewEnvironmentConfigurationData {
	return cm.PreviewEnvironmentConfigurationData{
		URL:        url,
		EntityType: "ContentType",
		EntityId:   contentTypeID,
		Enabled:    enabled,
	}
}

func TestPreviewEnvironmentLifecycle(t *testing.T) {
	t.Parallel()

	server, testserver := testContentfulManagementHTTPTestServer(t, cmt.WithRateLimitPerSecond(100))
	defer testserver.Close()

	server.RegisterSpaceEnvironment("space", "master")

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	createResponse, err := client.CreatePreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"preview",
		previewEnvironmentConfiguration("page", "https://preview.invalid/pages/{entry.sys.id}", true),
	)), cm.CreatePreviewEnvironmentParams{SpaceID: "space"})
	require.NoError(t, err)

	created, ok := createResponse.(*cm.PreviewEnvironment)
	require.True(t, ok)
	require.Equal(t, 0, created.Sys.Version)
	require.Equal(t, "page", created.Configurations[0].ContentType.Or(""))
	require.Equal(t, "ContentType", created.Configurations[0].EntityType.Or(""))

	updateResponse, err := client.PutPreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"renamed",
		previewEnvironmentConfiguration("page", "https://preview.invalid/pages/{entry.sys.id}", true),
		previewEnvironmentConfiguration("author", "https://preview.invalid/authors/{entry.sys.id}", false),
	)), cm.PutPreviewEnvironmentParams{
		SpaceID:              "space",
		PreviewEnvironmentID: created.Sys.ID,
		XContentfulVersion:   created.Sys.Version,
	})
	require.NoError(t, err)

	updated, ok := updateResponse.(*cm.PreviewEnvironment)
	require.True(t, ok)
	require.Equal(t, 1, updated.Sys.Version)
	require.Equal(t, []string{"page", "author"}, []string{
		updated.Configurations[0].EntityId.Or(""),
		updated.Configurations[1].EntityId.Or(""),
	})
	require.False(t, updated.Configurations[1].Enabled)

	staleResponse, err := client.PutPreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"stale update",
		previewEnvironmentConfiguration("page", "https://preview.invalid/pages/{entry.sys.id}", true),
	)), cm.PutPreviewEnvironmentParams{
		SpaceID:              "space",
		PreviewEnvironmentID: created.Sys.ID,
		XContentfulVersion:   0,
	})
	require.NoError(t, err)

	staleError, ok := staleResponse.(cm.ErrorStatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, 409, staleError.GetStatusCode())
	staleErrorBody, ok := staleError.GetError()
	require.True(t, ok)
	require.Equal(t, "Conflict", staleErrorBody.Sys.ID)

	deleteResponse, err := client.DeletePreviewEnvironment(t.Context(), cm.DeletePreviewEnvironmentParams{
		SpaceID:              "space",
		PreviewEnvironmentID: created.Sys.ID,
	})
	require.NoError(t, err)

	_, ok = deleteResponse.(*cm.NoContent)
	require.True(t, ok)

	getResponse, err := client.GetPreviewEnvironment(t.Context(), cm.GetPreviewEnvironmentParams{
		SpaceID:              "space",
		PreviewEnvironmentID: created.Sys.ID,
	})
	require.NoError(t, err)

	notFound, ok := getResponse.(*cm.ErrorStatusCode)
	require.True(t, ok)
	require.Equal(t, 404, notFound.StatusCode)
}

func TestPreviewEnvironmentConfigurationSemantics(t *testing.T) {
	t.Parallel()

	server, testserver := testContentfulManagementHTTPTestServer(t, cmt.WithRateLimitPerSecond(100))
	defer testserver.Close()

	server.RegisterSpaceEnvironment("space", "master")

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	createResponse, err := client.CreatePreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"preview",
		previewEnvironmentConfiguration("page", "https://preview.invalid/page", true),
		previewEnvironmentConfiguration("author", "https://preview.invalid/author", true),
	)), cm.CreatePreviewEnvironmentParams{SpaceID: "space"})
	require.NoError(t, err)

	created, ok := createResponse.(*cm.PreviewEnvironment)
	require.True(t, ok)

	updateResponse, err := client.PutPreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"preview",
		previewEnvironmentConfiguration("author", "https://preview.invalid/author-v2", false),
	)), cm.PutPreviewEnvironmentParams{
		SpaceID:              "space",
		PreviewEnvironmentID: created.Sys.ID,
		XContentfulVersion:   0,
	})
	require.NoError(t, err)

	updated, ok := updateResponse.(*cm.PreviewEnvironment)
	require.True(t, ok)
	require.Equal(t, 0, updated.Sys.Version, "configuration-only updates do not increment the live API version")
	require.Equal(t, []string{"page", "author"}, []string{
		updated.Configurations[0].EntityId.Or(""),
		updated.Configurations[1].EntityId.Or(""),
	}, "the live API preserves existing order and does not remove omitted configurations")
	require.Equal(t, "https://preview.invalid/author-v2", updated.Configurations[1].URL)
	require.False(t, updated.Configurations[1].Enabled)

	duplicateResponse, err := client.CreatePreviewEnvironment(t.Context(), new(previewEnvironmentData(
		"duplicates",
		previewEnvironmentConfiguration("page", "https://one.invalid", true),
		previewEnvironmentConfiguration("page", "https://two.invalid", true),
	)), cm.CreatePreviewEnvironmentParams{SpaceID: "space"})
	require.NoError(t, err)

	duplicateError, ok := duplicateResponse.(*cm.ErrorStatusCode)
	require.True(t, ok)
	require.Equal(t, 400, duplicateError.StatusCode)

	requestBody := `{
		"name":"unsafe update response round-trip",
		"description":"description",
		"configurations":[{
			"url":"https://unsafe.invalid",
			"entityType":"ContentType",
			"entityId":"page",
			"enabled":true,
			"example":false,
			"contentType":"page"
		}]
	}`
	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		testserver.URL+"/spaces/space/preview_environments/"+created.Sys.ID,
		strings.NewReader(requestBody),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Version", "0")

	rawUpdateResponse, err := testserver.Client().Do(request)
	require.NoError(t, err)

	defer rawUpdateResponse.Body.Close()

	require.Equal(t, http.StatusBadRequest, rawUpdateResponse.StatusCode)
}
