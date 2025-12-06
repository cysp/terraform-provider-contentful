package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		Limit:         cm.NewOptInt64(req.Limit),
		Order:         []string{"sys.id"},
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		response, err := r.providerData.client.GetContentTypes(ctx, params)
		if err != nil {
			yield(list.ListResult{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("Failed to list content types", util.ErrorDetailFromContentfulManagementResponse(response, err)),
				},
			})

			return
		}

		switch response := response.(type) {
		case *cm.ContentTypeCollection:
			for _, item := range response.Items {
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

				if !yield(result) {
					return
				}
			}

		default:
			yield(list.ListResult{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("Failed to list content types", util.ErrorDetailFromContentfulManagementResponse(response, err)),
				},
			})

			return
		}
	}
}
