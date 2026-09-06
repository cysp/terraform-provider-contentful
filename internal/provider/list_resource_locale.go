package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	params, paramsDiags := config.requestParams()
	if paramsDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(paramsDiags)

		return
	}

	fetchPage := func(ctx context.Context, skip int64, limit int64) (cm.GetLocalesRes, error) {
		pageParams := params
		pageParams.Skip = cm.NewOptInt64(skip)
		pageParams.Limit = cm.NewOptInt64(limit)

		return r.providerData.client.GetLocales(ctx, pageParams)
	}
	newResult := func(item cm.Locale) list.ListResult {
		return newListResultFromResponse(
			ctx,
			req,
			item.Name,
			LocaleIdentityModel{
				SpaceID:       types.StringValue(item.Sys.Space.Sys.ID),
				EnvironmentID: types.StringValue(item.Sys.Environment.Sys.ID),
				LocaleID:      types.StringValue(item.Sys.ID),
			},
			func() (LocaleModel, diag.Diagnostics) {
				return NewLocaleResourceModelFromResponse(ctx, item)
			},
		)
	}

	stream.Results = paginateContentfulCollectionItemsAsListResults(
		ctx,
		req,
		"Failed to list locales",
		fetchPage,
		newResult,
	)
}
