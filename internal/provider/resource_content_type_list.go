package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

type contentTypeListConfig struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

func (r *contentTypeListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = ContentTypeListResourceSchema(ctx)
}

func (r *contentTypeListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	diags := diag.Diagnostics{}

	var config contentTypeListConfig

	diags.Append(req.Config.Get(ctx, &config)...)

	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	params := cm.GetContentTypesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
	}

	resp.Results = func(yield func(list.ListResult) bool) {
		diags := diag.Diagnostics{}

		response, err := r.providerData.client.GetContentTypes(ctx, params)
		if err != nil {
			diags.AddError("Unable to list Content Types", err.Error())

			yield(list.ListResult{
				Diagnostics: diags,
			})

			return
		}

		switch response := response.(type) {
		case *cm.GetContentTypesOK:
			// NOTE: we ignore pagination for now as the API seems to return all results anyway.
			// If this becomes an issue we can add pagination support later.
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
