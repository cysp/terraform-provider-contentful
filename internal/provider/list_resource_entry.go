package provider

import (
	"context"
	"net/http"
	"net/url"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	params := cm.GetEntriesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
	}

	configContentType := config.ContentType.ValueString()
	if configContentType != "" {
		params.ContentType.SetTo(configContentType)
	}

	configOrder := config.Order.Elements()
	if configOrder != nil {
		order := make([]string, 0, len(configOrder))
		for _, orderElement := range configOrder {
			orderElementString := orderElement.ValueString()
			if orderElementString != "" {
				order = append(order, orderElementString)
			}
		}

		params.Order = order
	}

	getEntriesQueryOption := cm.WithEditRequest(func(req *http.Request) error {
		urlQuery := req.URL.Query()

		for key, value := range config.Query.Elements() {
			setEntryListQueryParam(urlQuery, key, value.ValueString())
		}

		req.URL.RawQuery = urlQuery.Encode()

		return nil
	})

	stream.Results = paginateContentfulCollectionItemsAsListResults(ctx, req,
		"Failed to list entries",
		func(ctx context.Context, skip int64, limit int64) (cm.GetEntriesRes, error) {
			pageParams := params
			pageParams.Skip = cm.NewOptInt64(skip)
			pageParams.Limit = cm.NewOptInt64(limit)

			return r.providerData.client.GetEntries(ctx, pageParams, getEntriesQueryOption)
		},
		func(item cm.Entry) list.ListResult {
			result := req.NewListResult(ctx)

			result.DisplayName = item.Sys.ID

			result.Diagnostics.Append(result.Identity.Set(ctx, NewEntryIdentityModel(
				item.Sys.Space.Sys.ID,
				item.Sys.Environment.Sys.ID,
				item.Sys.ID,
			))...)

			if req.IncludeResource {
				responseModel, responseDiags := NewEntryResourceModelFromResponse(ctx, item)
				result.Diagnostics.Append(responseDiags...)

				result.Diagnostics.Append(result.Resource.Set(ctx, &responseModel)...)
			}

			return result
		},
	)
}

func setEntryListQueryParam(query url.Values, key string, value string) {
	if key == "skip" || key == "limit" {
		return
	}

	query.Set(key, value)
}
