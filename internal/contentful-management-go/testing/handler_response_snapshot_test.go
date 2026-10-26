package cmtesting_test

import (
	"net/http"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentResponseRemainsStableDuringUpdates(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")
	result, err := handler.GetEnvironment(t.Context(), cm.GetEnvironmentParams{SpaceID: "space", EnvironmentID: "master"})
	require.NoError(t, err)

	response, ok := result.(*cm.Environment)
	require.True(t, ok)
	require.Equal(t, "master", response.Name)
	require.Equal(t, 1, response.Sys.Version)

	var encoders sync.WaitGroup
	t.Cleanup(encoders.Wait)
	encoders.Go(func() {
		for range 1000 {
			encoder := jx.Encoder{}
			response.Encode(&encoder)
		}
	})

	for version := 1; version <= 1000; version++ {
		updated, err := handler.CreateOrUpdateEnvironment(t.Context(), &cm.EnvironmentData{Name: "changed"}, cm.CreateOrUpdateEnvironmentParams{
			SpaceID: "space", EnvironmentID: "master", XContentfulVersion: cm.NewOptInt(version),
		})
		require.NoError(t, err)
		assert.IsType(t, &cm.EnvironmentStatusCode{}, updated)
	}

	encoders.Wait()

	assert.Equal(t, "master", response.Name)
	assert.Equal(t, 1, response.Sys.Version)
}

func TestPreviewEnvironmentResponsesRetainConfigurationsAfterUpdate(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")
	created, err := handler.PutPreviewEnvironment(t.Context(), &cm.PreviewEnvironmentData{
		Name: "Preview", Configurations: []cm.PreviewEnvironmentConfigurationData{{
			EntityType: "ContentType", EntityId: "article", URL: "https://example.invalid/original", Enabled: true,
		}},
	}, cm.PutPreviewEnvironmentParams{SpaceID: "space", PreviewEnvironmentID: "preview"})
	require.NoError(t, err)

	initial, ok := created.(*cm.PreviewEnvironment)
	require.True(t, ok)

	read, err := handler.GetPreviewEnvironment(t.Context(), cm.GetPreviewEnvironmentParams{SpaceID: "space", PreviewEnvironmentID: "preview"})
	require.NoError(t, err)

	readResponse, ok := read.(*cm.PreviewEnvironment)
	require.True(t, ok)

	updated, err := handler.PutPreviewEnvironment(t.Context(), &cm.PreviewEnvironmentData{
		Name: "Preview", Configurations: []cm.PreviewEnvironmentConfigurationData{{
			EntityType: "ContentType", EntityId: "article", URL: "https://example.invalid/updated", Enabled: true,
		}},
	}, cm.PutPreviewEnvironmentParams{SpaceID: "space", PreviewEnvironmentID: "preview", XContentfulVersion: initial.Sys.Version})
	require.NoError(t, err)

	updatedResponse, ok := updated.(*cm.PreviewEnvironment)
	require.True(t, ok)
	require.Len(t, updatedResponse.Configurations, 1)
	assert.Equal(t, "https://example.invalid/updated", updatedResponse.Configurations[0].URL)

	for _, response := range []*cm.PreviewEnvironment{initial, readResponse} {
		require.Len(t, response.Configurations, 1)
		assert.Equal(t, "https://example.invalid/original", response.Configurations[0].URL)
	}
}

func TestEntryResponsesDoNotExposeStoredFields(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")
	created, err := handler.PutEntry(t.Context(), &cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{"title": []byte(testJSON(map[string]any{"en-US": "original"}))}),
	}, cm.PutEntryParams{SpaceID: "space", EnvironmentID: "master", EntryID: "entry", XContentfulContentType: cm.NewOptString("article")})
	require.NoError(t, err)

	createdResponse, ok := created.(*cm.EntryStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusCreated, createdResponse.StatusCode)
	createdResponse.Response.Fields.Value["title"][0] = '!'

	read, err := handler.GetEntry(t.Context(), cm.GetEntryParams{SpaceID: "space", EnvironmentID: "master", EntryID: "entry"})
	require.NoError(t, err)

	readResponse, ok := read.(*cm.Entry)
	require.True(t, ok)
	assert.JSONEq(t, testJSON(map[string]any{"en-US": "original"}), string(readResponse.Fields.Value["title"]))
	readResponse.Fields.Value["title"][0] = '!'

	listed, err := handler.GetEntries(t.Context(), cm.GetEntriesParams{SpaceID: "space", EnvironmentID: "master"})
	require.NoError(t, err)

	listResponse, ok := listed.(*cm.EntryCollection)
	require.True(t, ok)
	require.Len(t, listResponse.Items, 1)
	assert.JSONEq(t, testJSON(map[string]any{"en-US": "original"}), string(listResponse.Items[0].Fields.Value["title"]))
	listResponse.Items[0].Fields.Value["title"][0] = '!'

	again, err := handler.GetEntry(t.Context(), cm.GetEntryParams{SpaceID: "space", EnvironmentID: "master", EntryID: "entry"})
	require.NoError(t, err)

	againResponse, ok := again.(*cm.Entry)
	require.True(t, ok)
	assert.JSONEq(t, testJSON(map[string]any{"en-US": "original"}), string(againResponse.Fields.Value["title"]))
}

func TestTaxonomySchemeResponseRetainsLinksAfterConceptDeletion(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	for _, id := range []string{"first", "second"} {
		created, err := handler.PutTaxonomyConcept(t.Context(), &cm.TaxonomyConceptRequest{
			PrefLabel: cm.LocalizedString{"en-US": id},
		}, cm.PutTaxonomyConceptParams{OrganizationID: "org", TaxonomyConceptID: id})
		require.NoError(t, err)
		require.IsType(t, &cm.TaxonomyConcept{}, created)
	}

	created, err := handler.PutTaxonomyConceptScheme(t.Context(), &cm.TaxonomyConceptSchemeRequest{
		PrefLabel: cm.LocalizedString{"en-US": "Scheme"},
		Concepts:  []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("first"), cm.NewTaxonomyConceptLink("second")},
	}, cm.PutTaxonomyConceptSchemeParams{OrganizationID: "org", TaxonomyConceptSchemeID: "scheme"})
	require.NoError(t, err)

	scheme, ok := created.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)

	deleted, err := handler.DeleteTaxonomyConcept(t.Context(), cm.DeleteTaxonomyConceptParams{
		OrganizationID: "org", TaxonomyConceptID: "first", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deleted)

	require.Len(t, scheme.Concepts, 2)
	assert.Equal(t, "first", scheme.Concepts[0].Sys.ID)
	assert.Equal(t, "second", scheme.Concepts[1].Sys.ID)

	current, err := handler.GetTaxonomyConceptScheme(t.Context(), cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: "org", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)

	currentScheme, ok := current.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)
	require.Len(t, currentScheme.Concepts, 1)
	assert.Equal(t, "second", currentScheme.Concepts[0].Sys.ID)
}

func TestStoredWebhookDefinitionReturnsDetachedTopics(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
		Name: "Webhook", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"},
	})
	stored, ok := server.StoredWebhookDefinition("space", "webhook")
	require.True(t, ok)
	require.Len(t, stored.Topics, 1)
	stored.Topics[0] = "Entry.delete"

	again, ok := server.StoredWebhookDefinition("space", "webhook")
	require.True(t, ok)
	assert.Equal(t, []string{"Entry.publish"}, again.Topics)
}
