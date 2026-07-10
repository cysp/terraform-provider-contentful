package cmtesting_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContentTypesPaginates(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	for i := range 3 {
		server.SetContentType("space", "environment", fmt.Sprintf("content-type-%d", i), cm.ContentTypeRequestData{
			Name:   fmt.Sprintf("Content Type %d", i),
			Fields: []cm.ContentTypeRequestDataFieldsItem{},
		})
	}

	firstPageResponse, err := server.Handler().GetContentTypes(context.Background(), cm.GetContentTypesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		Skip:          cm.NewOptInt64(0),
		Limit:         cm.NewOptInt64(1),
	})
	require.NoError(t, err)

	firstPageCollection, firstPageCollectionOk := firstPageResponse.(*cm.ContentTypeCollection)
	require.True(t, firstPageCollectionOk)
	assert.Equal(t, 3, firstPageCollection.Total.Or(0))
	require.Len(t, firstPageCollection.Items, 1)
	assert.Equal(t, "content-type-0", firstPageCollection.Items[0].Sys.ID)
	assert.Equal(t, "Content Type 0", firstPageCollection.Items[0].Name)

	secondPageResponse, err := server.Handler().GetContentTypes(context.Background(), cm.GetContentTypesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		Skip:          cm.NewOptInt64(1),
		Limit:         cm.NewOptInt64(1),
	})
	require.NoError(t, err)

	secondPageCollection, secondPageCollectionOk := secondPageResponse.(*cm.ContentTypeCollection)
	require.True(t, secondPageCollectionOk)
	assert.Equal(t, 3, secondPageCollection.Total.Or(0))
	require.Len(t, secondPageCollection.Items, 1)
	assert.Equal(t, "content-type-1", secondPageCollection.Items[0].Sys.ID)
	assert.Equal(t, "Content Type 1", secondPageCollection.Items[0].Name)
}

func TestGetContentTypesRejectsInvalidPaginationParams(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.SetContentType("space", "environment", "content-type", cm.ContentTypeRequestData{
		Name:   "Content Type",
		Fields: []cm.ContentTypeRequestDataFieldsItem{},
	})

	testCases := map[string]struct {
		params  cm.GetContentTypesParams
		message string
	}{
		"negative skip": {
			params: cm.GetContentTypesParams{
				SpaceID: "space", EnvironmentID: "environment", Skip: cm.NewOptInt64(-1),
			},
			message: `The value provided for "skip" is invalid. Please provide a value larger than or equal to 0`,
		},
		"negative limit": {
			params: cm.GetContentTypesParams{
				SpaceID: "space", EnvironmentID: "environment", Limit: cm.NewOptInt64(-1),
			},
			message: `The value provided for "limit" is invalid. Please provide a value between 0 and 1000`,
		},
		"limit above maximum": {
			params: cm.GetContentTypesParams{
				SpaceID: "space", EnvironmentID: "environment", Limit: cm.NewOptInt64(1001),
			},
			message: `The value provided for "limit" is invalid. Please provide a value between 0 and 1000`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response, err := server.Handler().GetContentTypes(context.Background(), testCase.params)
			require.NoError(t, err)
			requireContentfulError(t, response, http.StatusBadRequest, "InvalidQuery", testCase.message)
		})
	}
}
