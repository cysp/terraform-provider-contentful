package cmtesting

import (
	"context"
	"fmt"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContentTypesPaginates(t *testing.T) {
	t.Parallel()

	server, err := NewContentfulManagementServer()
	require.NoError(t, err)

	for i := range 3 {
		server.SetContentType("space", "environment", fmt.Sprintf("content-type-%d", i), cm.ContentTypeRequestData{
			Name:   fmt.Sprintf("Content Type %d", i),
			Fields: []cm.ContentTypeRequestDataFieldsItem{},
		})
	}

	response, err := server.Handler().GetContentTypes(context.Background(), cm.GetContentTypesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		Skip:          cm.NewOptInt64(1),
		Limit:         cm.NewOptInt64(1),
	})
	require.NoError(t, err)

	collection, ok := response.(*cm.ContentTypeCollection)
	require.True(t, ok)
	assert.Equal(t, 3, collection.Total.Or(0))
	assert.Len(t, collection.Items, 1)
}
