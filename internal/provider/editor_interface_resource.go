package provider

import (
	"context"
	"net/http"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_editor_interface"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithImportState = (*editorInterfaceResource)(nil)
)

//nolint:ireturn
func NewEditorInterfaceResource() resource.Resource {
	return &editorInterfaceResource{}
}

type editorInterfaceResource struct {
	providerData ContentfulProviderData
}

func (r *editorInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_editor_interface"
}

func (r *editorInterfaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_editor_interface.EditorInterfaceResourceSchema(ctx)
}

func (r *editorInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	util.ProviderDataFromResourceConfigureRequest(req, &r.providerData, resp)
}

func (r *editorInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	util.ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *editorInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_editor_interface.EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1
	request := data.ToPutEditorInterfaceReq(ctx, &resp.Diagnostics)

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, contentfulManagement.PutEditorInterfaceParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	})

	switch response := response.(type) {
	case *contentfulManagement.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(ReadEditorInterfaceModel(ctx, &data, *response)...)

	default:
		resp.Diagnostics.AddError("Failed to create editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetInt(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *editorInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_editor_interface.EditorInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 0

	response, err := r.providerData.client.GetEditorInterface(ctx, contentfulManagement.GetEditorInterfaceParams{
		SpaceID:       data.SpaceId.ValueString(),
		EnvironmentID: data.EnvironmentId.ValueString(),
		ContentTypeID: data.ContentTypeId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(ReadEditorInterfaceModel(ctx, &data, *response)...)

	case *contentfulManagement.ErrorStatusCode:
		if response.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Failed to read editor interface", response.Response.Message)
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read editor interface", response.Response.Message)

	default:
		resp.Diagnostics.AddError("Failed to read editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetInt(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *editorInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_editor_interface.EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion, currentVersionDiags := util.PrivateDataGetInt(ctx, req.Private, "version")
	resp.Diagnostics.Append(currentVersionDiags...)

	request := data.ToPutEditorInterfaceReq(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, contentfulManagement.PutEditorInterfaceParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	})

	switch response := response.(type) {
	case *contentfulManagement.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(ReadEditorInterfaceModel(ctx, &data, *response)...)

	default:
		resp.Diagnostics.AddError("Failed to update editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetInt(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *editorInterfaceResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete editor interfaces
}