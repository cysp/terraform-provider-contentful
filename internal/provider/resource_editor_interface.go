package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithIdentity    = (*editorInterfaceResource)(nil)
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
	resp.Schema = EditorInterfaceResourceSchema(ctx)
}

func (r *editorInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *editorInterfaceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":        identityschema.StringAttribute{RequiredForImport: true},
			"environment_id":  identityschema.StringAttribute{RequiredForImport: true},
			"content_type_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *editorInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *editorInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1
	currentVersion += r.providerData.editorInterfaceVersionOffset.Get(plan.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		ContentTypeID:      plan.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := plan.ToEditorInterfaceData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, params)

	tflog.Info(ctx, "editor_interface.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EditorInterfaceModel

	switch response := response.(type) {
	case *cm.EditorInterfaceStatusCode:
		responseModel, responseModelDiags := NewEditorInterfaceResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EditorInterfaceIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)

	r.providerData.editorInterfaceVersionOffset.Reset(plan.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EditorInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetEditorInterfaceParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ContentTypeID: state.ContentTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetEditorInterface(ctx, params)

	tflog.Info(ctx, "editor_interface.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	var data EditorInterfaceModel

	switch response := response.(type) {
	case *cm.EditorInterface:
		responseModel, responseModelDiags := NewEditorInterfaceResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EditorInterfaceIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)

	r.providerData.editorInterfaceVersionOffset.Reset(state.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	currentVersion += r.providerData.editorInterfaceVersionOffset.Get(plan.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		ContentTypeID:      plan.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := plan.ToEditorInterfaceData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, params)

	tflog.Info(ctx, "editor_interface.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EditorInterfaceModel

	switch response := response.(type) {
	case *cm.EditorInterfaceStatusCode:
		responseModel, responseModelDiags := NewEditorInterfaceResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EditorInterfaceIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)

	r.providerData.editorInterfaceVersionOffset.Reset(plan.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete editor interfaces
}
