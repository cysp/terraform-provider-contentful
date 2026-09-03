package provider

import (
	"context"
	"net/http"
	"net/url"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ list.ListResource              = (*entryListResource)(nil)
	_ list.ListResourceWithConfigure = (*entryListResource)(nil)
)

//nolint:ireturn
func NewEntryListResource() list.ListResource {
	return &entryListResource{}
}

type entryListResource struct {
	providerData ContentfulProviderData
}

func (r *entryListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *entryListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = EntryListResourceConfigSchema(ctx)
}

func (r *entryListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *entryListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config entryListResourceConfig

	configDiags := req.Config.Get(ctx, &config)
	if configDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(configDiags)

		return
	}

	request, requestDiags := config.request()
	if requestDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(requestDiags)

		return
	}

	getEntriesQueryOption := cm.WithEditRequest(func(req *http.Request) error {
		urlQuery := req.URL.Query()

		for key, value := range request.query {
			setEntryListQueryParam(urlQuery, key, value)
		}

		req.URL.RawQuery = urlQuery.Encode()

		return nil
	})

	stream.Results = paginateContentfulCollectionItemsAsListResults(
		ctx, req,
		"Failed to list entries",
		func(ctx context.Context, skip int64, limit int64) (cm.GetEntriesRes, error) {
			pageParams := request.params
			pageParams.Skip = cm.NewOptInt64(skip)
			pageParams.Limit = cm.NewOptInt64(limit)

			response, err := r.providerData.client.GetEntries(ctx, pageParams, getEntriesQueryOption)

			tflog.Info(ctx, "entry.list", map[string]any{
				"params": pageParams,
				"query":  request.query,
				// "response": response, omitted to avoid logging complete entry payloads
				"err": err,
			})

			return response, err //nolint:wrapcheck // preserve generated CMA client errors for list diagnostics.
		},
		func(item cm.Entry) list.ListResult {
			return newListResultFromResponse(
				ctx,
				req,
				item.Sys.ID,
				NewEntryIdentityModel(item.Sys.Space.Sys.ID, item.Sys.Environment.Sys.ID, item.Sys.ID),
				func() (*EntryModel, diag.Diagnostics) {
					responseModel, responseDiags := NewEntryResourceModelFromResponse(ctx, item)
					responseModel.Fields = mergeEntryFieldsWithFallback(
						responseModel.Fields,
						NewTypedMap(map[string]jsontypes.Normalized{}),
					)

					return &responseModel, responseDiags
				},
			)
		},
	)
}

func setEntryListQueryParam(query url.Values, key string, value string) {
	if key == "skip" || key == "limit" {
		return
	}

	query.Set(key, value)
}
