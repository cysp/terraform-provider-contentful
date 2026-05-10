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
