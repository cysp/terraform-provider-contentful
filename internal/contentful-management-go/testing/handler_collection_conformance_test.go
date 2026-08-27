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

func TestGetEntriesReturnsStableFilteredPagination(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	for _, entry := range []struct {
		id            string
		contentTypeID string
	}{
		{id: "charlie", contentTypeID: "article"},
		{id: "alpha", contentTypeID: "article"},
		{id: "ignored", contentTypeID: "author"},
		{id: "bravo", contentTypeID: "article"},
	} {
		response, err := handler.PutEntry(context.Background(), &cm.EntryRequest{}, cm.PutEntryParams{
			SpaceID:                "space",
			EnvironmentID:          "environment",
			EntryID:                entry.id,
			XContentfulContentType: cm.NewOptString(entry.contentTypeID),
		})
		require.NoError(t, err)

		statusCode, ok := response.(*cm.EntryStatusCode)
		require.True(t, ok)
		assert.Equal(t, http.StatusCreated, statusCode.StatusCode)
	}

	firstResponse, err := handler.GetEntries(context.Background(), cm.GetEntriesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		ContentType:   cm.NewOptString("article"),
		Limit:         cm.NewOptInt64(2),
	})
	require.NoError(t, err)

	first, ok := firstResponse.(*cm.EntryCollection)
	require.True(t, ok)
	assert.Equal(t, 3, first.Total.Or(-1))
	assert.Equal(t, 0, first.Skip.Or(-1))
	assert.Equal(t, 2, first.Limit.Or(-1))
	require.Len(t, first.Items, 2)
	assert.Equal(t, []string{"alpha", "bravo"}, []string{first.Items[0].Sys.ID, first.Items[1].Sys.ID})

	secondResponse, err := handler.GetEntries(context.Background(), cm.GetEntriesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		ContentType:   cm.NewOptString("article"),
		Skip:          cm.NewOptInt64(2),
		Limit:         cm.NewOptInt64(2),
	})
	require.NoError(t, err)

	second, ok := secondResponse.(*cm.EntryCollection)
	require.True(t, ok)
	assert.Equal(t, 3, second.Total.Or(-1))
	assert.Equal(t, 2, second.Skip.Or(-1))
	assert.Equal(t, 2, second.Limit.Or(-1))
	require.Len(t, second.Items, 1)
	assert.Equal(t, "charlie", second.Items[0].Sys.ID)

	beyondEndResponse, err := handler.GetEntries(context.Background(), cm.GetEntriesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		ContentType:   cm.NewOptString("article"),
		Skip:          cm.NewOptInt64(100),
		Limit:         cm.NewOptInt64(2),
	})
	require.NoError(t, err)

	beyondEnd, ok := beyondEndResponse.(*cm.EntryCollection)
	require.True(t, ok)
	assert.Equal(t, 3, beyondEnd.Total.Or(-1))
	assert.Equal(t, 100, beyondEnd.Skip.Or(-1))
	assert.Equal(t, 2, beyondEnd.Limit.Or(-1))
	assert.Empty(t, beyondEnd.Items)
}

func TestGetEntriesRejectsInvalidPaginationWithoutPanicking(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	tests := map[string]cm.GetEntriesParams{
		"negative skip": {
			SpaceID: "space", EnvironmentID: "environment", Skip: cm.NewOptInt64(-1),
		},
		"negative limit": {
			SpaceID: "space", EnvironmentID: "environment", Limit: cm.NewOptInt64(-1),
		},
		"limit above maximum": {
			SpaceID: "space", EnvironmentID: "environment", Limit: cm.NewOptInt64(1001),
		},
	}

	for name, params := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response, err := handler.GetEntries(context.Background(), params)
			require.NoError(t, err)

			statusCode, ok := response.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusBadRequest, statusCode.StatusCode)

			errorResponse, ok := statusCode.Response.GetError()
			require.True(t, ok)
			assert.Equal(t, "InvalidQuery", errorResponse.Sys.ID)
			assert.NotEmpty(t, errorResponse.Message.Or(""))
		})
	}
}
