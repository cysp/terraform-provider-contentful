package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource              = (*contentTypeListResource)(nil)
	_ list.ListResourceWithConfigure = (*contentTypeListResource)(nil)
)

//nolint:ireturn
func NewContentTypeListResource() list.ListResource {
	return &contentTypeListResource{}
}

type contentTypeListResource struct {
	providerData ContentfulProviderData
}

func (r *contentTypeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_type"
}

func (r *contentTypeListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *contentTypeListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = ContentTypeListResourceConfigSchema(ctx)
}

func (r *contentTypeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config contentTypeListResourceConfig

	configDiags := req.Config.Get(ctx, &config)
	if configDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(configDiags)

		return
	}

	params := cm.GetContentTypesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
		Order:         []string{"sys.id"},
	}

	stream.Results = paginateContentfulCollectionItemsAsListResults(ctx, req,
		"Failed to list content types",
		func(ctx context.Context, skip int64, limit int64) (cm.GetContentTypesRes, error) {
			pageParams := params
			pageParams.Skip = cm.NewOptInt64(skip)
			pageParams.Limit = cm.NewOptInt64(limit)

			return r.providerData.client.GetContentTypes(ctx, pageParams)
		},
		func(item cm.ContentType) list.ListResult {
			result := req.NewListResult(ctx)

			result.DisplayName = item.Name

			result.Diagnostics.Append(result.Identity.Set(ctx, ContentTypeIdentityModel{
				SpaceID:       types.StringValue(item.Sys.Space.Sys.ID),
				EnvironmentID: types.StringValue(item.Sys.Environment.Sys.ID),
				ContentTypeID: types.StringValue(item.Sys.ID),
			})...)

			if req.IncludeResource {
				responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, item)
				result.Diagnostics.Append(responseModelDiags...)

				result.Diagnostics.Append(result.Resource.Set(ctx, responseModel)...)
			}

			return result
		},
	)
}
