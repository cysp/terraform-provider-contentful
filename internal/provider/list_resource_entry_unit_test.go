//nolint:testpackage
package provider

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryListResourceListIgnoresQueryPaginationParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	httpServer := httptest.NewServer(server)
	t.Cleanup(httpServer.Close)

	client, err := cm.NewClient(
		httpServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(httpServer.Client()),
	)
	require.NoError(t, err)

	server.SetEntry("space", "environment", "author", "entry-1", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"name": jx.Raw(`{"en-US":"Entry 1"}`),
		}),
	})

	listResource := &entryListResource{
		providerData: ContentfulProviderData{client: client},
	}

	var identitySchemaResponse resource.IdentitySchemaResponse
	(&entryResource{}).IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identitySchemaResponse)

	var stream list.ListResultsStream
	listResource.List(ctx, list.ListRequest{
		Config:                 newEntryListResourceConfig(ctx),
		ResourceSchema:         EntryResourceSchema(ctx),
		ResourceIdentitySchema: identitySchemaResponse.IdentitySchema,
	}, &stream)

	var results []list.ListResult

	stream.Results(func(result list.ListResult) bool {
		results = append(results, result)

		return true
	})

	require.Len(t, results, 1)
	assert.False(t, results[0].Diagnostics.HasError())
	assert.Equal(t, "entry-1", results[0].DisplayName)
}

func TestSetEntryListQueryParamSkipsPaginatorParams(t *testing.T) {
	t.Parallel()

	query := url.Values{}

	setEntryListQueryParam(query, "fields.name[ne]", "nonexistent")
	setEntryListQueryParam(query, "limit", "1")
	setEntryListQueryParam(query, "skip", "100")

	assert.Equal(t, url.Values{
		"fields.name[ne]": []string{"nonexistent"},
	}, query)
}

func newEntryListResourceConfig(ctx context.Context) tfsdk.Config {
	schema := EntryListResourceConfigSchema(ctx)

	return tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"space_id":       tftypes.String,
				"environment_id": tftypes.String,
				"content_type":   tftypes.String,
				"order":          tftypes.List{ElementType: tftypes.String},
				"query":          tftypes.Map{ElementType: tftypes.String},
			},
		}, map[string]tftypes.Value{
			"space_id":       tftypes.NewValue(tftypes.String, "space"),
			"environment_id": tftypes.NewValue(tftypes.String, "environment"),
			"content_type":   tftypes.NewValue(tftypes.String, "author"),
			"order":          tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"query": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"limit": tftypes.NewValue(tftypes.String, "1"),
				"skip":  tftypes.NewValue(tftypes.String, "100"),
			}),
		}),
		Schema: schema,
	}
}
