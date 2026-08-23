package cmtesting_test

import (
	"context"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishEntryRequiresExactVersion(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	createdResponse, err := handler.PutEntry(context.Background(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:                "space",
		EnvironmentID:          "environment",
		EntryID:                "entry",
		XContentfulContentType: cm.NewOptString("article"),
		XContentfulVersion:     1,
	})
	require.NoError(t, err)

	createdStatus, ok := createdResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusCreated, createdStatus.StatusCode)
	assert.Equal(t, 1, createdStatus.Response.Sys.Version)

	staleResponse, err := handler.PublishEntry(context.Background(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: 0,
	})
	require.NoError(t, err)
	requireContentfulConflictWithNonemptyMessage(t, staleResponse, cm.ErrorSysIDVersionMismatch)

	publishedResponse, err := handler.PublishEntry(context.Background(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: createdStatus.Response.Sys.Version,
	})
	require.NoError(t, err)

	publishedStatus, ok := publishedResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusOK, publishedStatus.StatusCode)
	assert.Equal(t, 2, publishedStatus.Response.Sys.Version)
	assert.Equal(t, 1, publishedStatus.Response.Sys.PublishedVersion.Or(0))
}
