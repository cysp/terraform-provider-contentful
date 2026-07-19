package provider

import (
	"context"
	"net/http"
	"net/url"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	spaceID, spaceIDDiags := KnownStringValue(config.SpaceID, path.Root("space_id"))
	configDiags.Append(spaceIDDiags...)

	environmentID, environmentIDDiags := KnownStringValue(config.EnvironmentID, path.Root("environment_id"))
	configDiags.Append(environmentIDDiags...)

	params := cm.GetEntriesParams{SpaceID: spaceID, EnvironmentID: environmentID}

	if config.ContentType.IsUnknown() {
		configDiags.AddAttributeError(path.Root("content_type"), "Unexpected unknown content type", "The content type must be known before entries can be listed.")
	} else if !config.ContentType.IsNull() {
		configContentType := config.ContentType.ValueString()
		if configContentType != "" {
			params.ContentType.SetTo(configContentType)
		}
	}

	if config.Order.IsUnknown() {
		configDiags.AddAttributeError(path.Root("order"), "Unexpected unknown order", "Entry ordering must be known before entries can be listed.")
	} else if !config.Order.IsNull() {
		configOrder := config.Order.Elements()

		order := make([]string, 0, len(configOrder))
		for index, orderElement := range configOrder {
			orderElementString, orderElementDiags := KnownStringValue(orderElement, path.Root("order").AtListIndex(index))
			configDiags.Append(orderElementDiags...)

			if orderElementDiags.HasError() {
				continue
			}

			if orderElementString != "" {
				order = append(order, orderElementString)
			}
		}

		params.Order = order
	}

	if config.Query.IsUnknown() {
		configDiags.AddAttributeError(path.Root("query"), "Unexpected unknown query", "Entry query parameters must be known before entries can be listed.")
	} else if !config.Query.IsNull() {
		for key, value := range config.Query.Elements() {
			_, valueDiags := KnownStringValue(value, path.Root("query").AtMapKey(key))
			configDiags.Append(valueDiags...)
		}
	}

	if configDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(configDiags)

		return
	}

	getEntriesQueryOption := cm.WithEditRequest(func(req *http.Request) error {
		urlQuery := req.URL.Query()

		for key, value := range config.Query.Elements() {
			setEntryListQueryParam(urlQuery, key, value.ValueString())
		}

		req.URL.RawQuery = urlQuery.Encode()

		return nil
	})

	stream.Results = paginateContentfulCollectionItemsAsListResults(
		ctx, req,
		"Failed to list entries",
		func(ctx context.Context, skip int64, limit int64) (cm.GetEntriesRes, error) {
			pageParams := params
			pageParams.Skip = cm.NewOptInt64(skip)
			pageParams.Limit = cm.NewOptInt64(limit)

			return r.providerData.client.GetEntries(ctx, pageParams, getEntriesQueryOption)
		},
		func(item cm.Entry) list.ListResult {
			return newListResultFromResponse(
				ctx,
				req,
				item.Sys.ID,
				NewEntryIdentityModel(item.Sys.Space.Sys.ID, item.Sys.Environment.Sys.ID, item.Sys.ID),
				func() (*EntryModel, diag.Diagnostics) {
					responseModel, responseDiags := NewEntryResourceModelFromResponse(ctx, item)

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
