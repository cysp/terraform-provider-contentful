package cmtesting_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetEntryStoresFieldValuesUnchanged(t *testing.T) {
	t.Parallel()

	request := cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"title": jx.Raw(`{"en-US":"Post 1"}`),
		}),
	}

	storedField := storeEntryAndGetField(t, request, "title")

	assert.JSONEq(t, `{"en-US":"Post 1"}`, string(storedField))
}

func TestEntryResponsesProjectOmittedFields(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.SetEntry("space", "environment", "content-type", "entry", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"empty":          jx.Raw(`{"en-US":[]}`),
			"empty-locales":  jx.Raw(`{}`),
			"nonempty-array": jx.Raw(`{"en-US":["value"]}`),
			"raw-null":       jx.Raw(`null`),
			"localized-null": jx.Raw(`{"en-US":null}`),
			"title":          jx.Raw(`{"en-US":"Post 1"}`),
		}),
	})

	getResponse, err := server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	entry, ok := getResponse.(*cm.Entry)
	require.True(t, ok)
	require.True(t, entry.Fields.IsSet())
	assert.NotContains(t, entry.Fields.Value, "empty")
	assert.NotContains(t, entry.Fields.Value, "raw-null")
	assert.JSONEq(t, `{}`, string(entry.Fields.Value["empty-locales"]))
	assert.JSONEq(t, `{"en-US":["value"]}`, string(entry.Fields.Value["nonempty-array"]))
	assert.JSONEq(t, `{"en-US":null}`, string(entry.Fields.Value["localized-null"]))
	assert.JSONEq(t, `{"en-US":"Post 1"}`, string(entry.Fields.Value["title"]))

	listResponse, err := server.Handler().GetEntries(t.Context(), cm.GetEntriesParams{
		SpaceID: "space", EnvironmentID: "environment", ContentType: cm.NewOptString("content-type"),
	})
	require.NoError(t, err)

	entries, ok := listResponse.(*cm.EntryCollection)
	require.True(t, ok)
	require.Len(t, entries.Items, 1)
	assert.NotContains(t, entries.Items[0].Fields.Value, "empty")
	assert.NotContains(t, entries.Items[0].Fields.Value, "raw-null")
	assert.JSONEq(t, `{"en-US":null}`, string(entries.Items[0].Fields.Value["localized-null"]))

	putResponse, err := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"empty":          jx.Raw(`{"en-US":[]}`),
			"raw-null":       jx.Raw(`null`),
			"localized-null": jx.Raw(`{"en-US":null}`),
		}),
	}, cm.PutEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
	})
	require.NoError(t, err)

	updated, ok := putResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	require.True(t, updated.Response.Fields.IsSet())
	assert.NotContains(t, updated.Response.Fields.Value, "empty")
	assert.NotContains(t, updated.Response.Fields.Value, "raw-null")
	assert.JSONEq(t, `{"en-US":null}`, string(updated.Response.Fields.Value["localized-null"]))

	publishResponse, err := server.Handler().PublishEntry(t.Context(), cm.PublishEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
		XContentfulVersion: updated.Response.Sys.Version,
	})
	require.NoError(t, err)

	published, ok := publishResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	require.True(t, published.Response.Fields.IsSet())
	assert.NotContains(t, published.Response.Fields.Value, "empty")
	assert.NotContains(t, published.Response.Fields.Value, "raw-null")
	assert.JSONEq(t, `{"en-US":null}`, string(published.Response.Fields.Value["localized-null"]))
}

func storeEntryAndGetField(t *testing.T, request cm.EntryRequest, fieldID string) jx.Raw {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.SetEntry("space", "environment", "content-type", "entry", request)

	response, err := server.Handler().GetEntry(context.Background(), cm.GetEntryParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		EntryID:       "entry",
	})
	require.NoError(t, err)

	entry, ok := response.(*cm.Entry)
	require.True(t, ok)

	return entry.Fields.Value[fieldID]
}
