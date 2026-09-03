//nolint:testpackage
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryListResourceListSendsFiltersAndOwnsPagination(t *testing.T) {
	t.Parallel()

	const (
		queryValueSentinel = "ENTRY_QUERY_VALUE_SENTINEL"
		fieldValueSentinel = "ENTRY_FIELD_VALUE_SENTINEL"
	)

	var logOutput bytes.Buffer

	ctx := tflogtest.RootLogger(t.Context(), &logOutput)

	var requestCount atomic.Int64

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/spaces/space/environments/environment/entries", r.URL.Path)
		assert.Equal(t, "Bearer "+cmt.ValidAccessToken, r.Header.Get("Authorization"))
		assert.Equal(t, url.Values{
			"content_type":    []string{"author"},
			"fields.name[ne]": []string{queryValueSentinel},
			"limit":           []string{"100"},
			"order":           []string{"sys.createdAt"},
			"skip":            []string{"0"},
		}, r.URL.Query())

		entry := cmt.NewEntryFromRequest("space", "environment", "author", "entry-1", &cm.EntryRequest{
			Fields: cm.NewOptEntryFields(cm.EntryFields{
				"name": jx.Raw(`{"en-US":"` + fieldValueSentinel + `"}`),
			}),
		})

		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(cm.EntryCollection{
			Sys:   cm.EntryCollectionSys{Type: cm.EntryCollectionSysTypeArray},
			Total: cm.NewOptInt(1),
			Skip:  cm.NewOptInt(0),
			Limit: cm.NewOptInt(100),
			Items: []cm.Entry{entry},
		}))
	}))
	t.Cleanup(httpServer.Close)

	client, err := cm.NewClient(
		httpServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(httpServer.Client()),
	)
	require.NoError(t, err)

	listResource := &entryListResource{
		providerData: ContentfulProviderData{client: client},
	}

	var identitySchemaResponse resource.IdentitySchemaResponse
	(&entryResource{}).IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identitySchemaResponse)

	var stream list.ListResultsStream
	listResource.List(ctx, list.ListRequest{
		Config:                 newEntryListResourceConfig(ctx, "fields.name[ne]", queryValueSentinel),
		ResourceSchema:         EntryResourceSchema(ctx),
		ResourceIdentitySchema: identitySchemaResponse.IdentitySchema,
	}, &stream)

	var results []list.ListResult

	stream.Results(func(result list.ListResult) bool {
		results = append(results, result)

		return true
	})

	require.Len(t, results, 1)
	assert.False(t, results[0].Diagnostics.HasError(), results[0].Diagnostics)
	assert.Equal(t, "entry-1", results[0].DisplayName)
	assert.Equal(t, int64(1), requestCount.Load())

	logEntries, err := tflogtest.MultilineJSONDecode(bytes.NewReader(logOutput.Bytes()))
	require.NoError(t, err)
	require.Len(t, logEntries, 1)
	assert.Equal(t, "entry.list", logEntries[0]["@message"])
	assert.Equal(t, "info", logEntries[0]["@level"])
	assert.Contains(t, logEntries[0], "params")
	assert.Contains(t, logEntries[0], "query")
	assert.Contains(t, logEntries[0], "err")
	assert.NotContains(t, logEntries[0], "response")
	assert.Contains(t, logOutput.String(), queryValueSentinel)
	assert.NotContains(t, logOutput.String(), fieldValueSentinel)
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

func newEntryListResourceConfig(ctx context.Context, queryKey string, queryValue string) tfsdk.Config {
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
			"order": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "sys.createdAt"),
			}),
			"query": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				queryKey: tftypes.NewValue(tftypes.String, queryValue),
				"limit":  tftypes.NewValue(tftypes.String, "1"),
				"skip":   tftypes.NewValue(tftypes.String, "100"),
			}),
		}),
		Schema: schema,
	}
}
