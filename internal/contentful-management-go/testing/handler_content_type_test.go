package cmtesting_test

import (
	"context"
	"fmt"
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
