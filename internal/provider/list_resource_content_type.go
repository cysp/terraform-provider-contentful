package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	diags := diag.Diagnostics{}

	var config contentTypeListResourceConfig
	diags.Append(req.Config.Get(ctx, &config)...)

	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	params := cm.GetContentTypesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
		Limit:         cm.NewOptInt64(req.Limit),
		Order:         cm.NewOptString("sys.id"),
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		diags := diag.Diagnostics{}

		response, err := r.providerData.client.GetContentTypes(ctx, params)
		if err != nil {
			diags.AddError("Failed to list content types", err.Error())

			yield(list.ListResult{
				Diagnostics: diags,
			})

			return
		}

		switch response := response.(type) {
		case *cm.ContentTypeCollection:
			for _, item := range response.Items {
				result := req.NewListResult(ctx)

				result.DisplayName = item.Name

				result.Identity.SetAttribute(ctx, path.Root("space_id"), item.Sys.Space.Sys.ID)
				result.Identity.SetAttribute(ctx, path.Root("environment_id"), item.Sys.Environment.Sys.ID)
				result.Identity.SetAttribute(ctx, path.Root("content_type_id"), item.Sys.ID)

				responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, item)
				result.Diagnostics.Append(responseModelDiags...)

				result.Diagnostics.Append(result.Resource.Set(ctx, responseModel)...)

				if !yield(result) {
					return
				}
			}

		default:
			diags.AddError("Failed to list content types", util.ErrorDetailFromContentfulManagementResponse(response, err))
			yield(list.ListResult{
				Diagnostics: diags,
			})

			return
		}
	}
}
