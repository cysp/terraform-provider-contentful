//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource              = (*localeListResource)(nil)
	_ list.ListResourceWithConfigure = (*localeListResource)(nil)
)

//nolint:ireturn
func NewLocaleListResource() list.ListResource {
	return &localeListResource{}
}

type localeListResource struct {
	providerData ContentfulProviderData
}

func (r *localeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locale"
}

func (r *localeListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *localeListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = LocaleListResourceConfigSchema(ctx)
}

func (r *localeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config localeListResourceConfig

	configDiags := req.Config.Get(ctx, &config)
	if configDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(configDiags)

		return
	}

	params := cm.GetLocalesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
		Order:         []string{"sys.id"},
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

	stream.Results = paginateContentfulCollectionItemsAsListResults(ctx, req,
		"Failed to list locales",
		func(ctx context.Context, skip int64, limit int64) (cm.GetLocalesRes, error) {
			pageParams := params
			pageParams.Skip = cm.NewOptInt64(skip)
			pageParams.Limit = cm.NewOptInt64(limit)

			return r.providerData.client.GetLocales(ctx, pageParams)
		},
		func(item cm.Locale) list.ListResult {
			result := req.NewListResult(ctx)

			result.DisplayName = item.Name

			result.Diagnostics.Append(result.Identity.Set(ctx, LocaleIdentityModel{
				SpaceID:       types.StringValue(item.Sys.Space.Sys.ID),
				EnvironmentID: types.StringValue(item.Sys.Environment.Sys.ID),
				LocaleID:      types.StringValue(item.Sys.ID),
			})...)

			if req.IncludeResource {
				responseModel, responseModelDiags := NewLocaleResourceModelFromResponse(ctx, item)
				result.Diagnostics.Append(responseModelDiags...)

				result.Diagnostics.Append(result.Resource.Set(ctx, responseModel)...)
			}

			return result
		},
	)
}
