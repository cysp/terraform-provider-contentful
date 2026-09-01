package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func editorInterfaceIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id", "content_type_id"}
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
	resp.IdentitySchema = resourceIdentitySchema(editorInterfaceIdentityAttributeNames())
}

func (r *editorInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, editorInterfaceIdentityAttributeNames(), req, resp)
}

func (r *editorInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := plan.Timeouts.Create(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	version := 1
	version += r.providerData.editorInterfaceVersionOffset.Get(plan.SpaceID.ValueString(), plan.EnvironmentID.ValueString(), plan.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		ContentTypeID:      plan.ContentTypeID.ValueString(),
		XContentfulVersion: version,
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

	var (
		data             EditorInterfaceModel
		consistencyDiags diag.Diagnostics
	)

	switch response := response.(type) {
	case *cm.EditorInterfaceStatusCode:
		var responseDiags diag.Diagnostics

		data, responseDiags, consistencyDiags = ReconcileEditorInterfaceMutationResponse(ctx, response.Response, plan)
		resp.Diagnostics.Append(responseDiags...)

		version = response.Response.Sys.Version

	default:
		if contentfulResponseIsVersionMismatch(response) {
			resp.Diagnostics.AddError("Editor Interface requires import", "Contentful rejected the request because the Editor Interface version did not match. Import the Editor Interface into this resource before applying again. If Terraform already tracks this resource, remove its existing state entry before importing it.")
		} else {
			resp.Diagnostics.AddError("Failed to create editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, editorInterfaceIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.providerData.editorInterfaceVersionOffset.Reset(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.ContentTypeID.ValueString())
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *editorInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EditorInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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

	version := 0

	var data EditorInterfaceModel

	switch response := response.(type) {
	case *cm.EditorInterface:
		responseModel, responseModelDiags := NewEditorInterfaceResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

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

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, editorInterfaceIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)

	r.providerData.editorInterfaceVersionOffset.Reset(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EditorInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := plan.Timeouts.Update(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version += r.providerData.editorInterfaceVersionOffset.Get(plan.SpaceID.ValueString(), plan.EnvironmentID.ValueString(), plan.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		ContentTypeID:      plan.ContentTypeID.ValueString(),
		XContentfulVersion: version,
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

	var (
		data             EditorInterfaceModel
		consistencyDiags diag.Diagnostics
	)

	switch response := response.(type) {
	case *cm.EditorInterfaceStatusCode:
		var responseDiags diag.Diagnostics

		data, responseDiags, consistencyDiags = ReconcileEditorInterfaceMutationResponse(ctx, response.Response, plan)

		resp.Diagnostics.Append(responseDiags...)

		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, editorInterfaceIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.providerData.editorInterfaceVersionOffset.Reset(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.ContentTypeID.ValueString())
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *editorInterfaceResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete editor interfaces
}
