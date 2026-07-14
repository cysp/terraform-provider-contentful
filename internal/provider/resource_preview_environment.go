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
	_ resource.Resource                = (*previewEnvironmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*previewEnvironmentResource)(nil)
	_ resource.ResourceWithIdentity    = (*previewEnvironmentResource)(nil)
	_ resource.ResourceWithImportState = (*previewEnvironmentResource)(nil)
)

//nolint:ireturn
func NewPreviewEnvironmentResource() resource.Resource {
	return &previewEnvironmentResource{}
}

type previewEnvironmentResource struct {
	providerData ContentfulProviderData
}

func (r *previewEnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_preview_environment"
}

func (r *previewEnvironmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PreviewEnvironmentResourceSchema(ctx)
}

func (r *previewEnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *previewEnvironmentResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":               identityschema.StringAttribute{RequiredForImport: true},
			"preview_environment_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *previewEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("preview_environment_id"),
	}, req, resp)
}

func (r *previewEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PreviewEnvironmentModel
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

	request, requestDiagnostics := plan.ToPreviewEnvironmentData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		response any
		err      error
	)
	if !plan.PreviewEnvironmentID.IsNull() && !plan.PreviewEnvironmentID.IsUnknown() {
		response, err = r.providerData.client.PutPreviewEnvironment(ctx, &request, cm.PutPreviewEnvironmentParams{
			SpaceID:              plan.SpaceID.ValueString(),
			PreviewEnvironmentID: plan.PreviewEnvironmentID.ValueString(),
			XContentfulVersion:   0,
		})
	} else {
		createRequest := cm.NewPreviewEnvironmentCreateData(request)
		response, err = r.providerData.client.CreatePreviewEnvironment(ctx, &createRequest, cm.CreatePreviewEnvironmentParams{
			SpaceID: plan.SpaceID.ValueString(),
		})
	}

	tflog.Info(ctx, "preview_environment.create", map[string]any{
		"request":  request,
		"response": response,
		"err":      err,
	})

	previewEnvironment, ok := response.(*cm.PreviewEnvironment)
	if !ok {
		resp.Diagnostics.AddError("Failed to create content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, dataDiagnostics := NewPreviewEnvironmentModelFromResponse(ctx, *previewEnvironment)
	resp.Diagnostics.Append(dataDiagnostics...)

	data.Timeouts = plan.Timeouts
	r.setCreateState(ctx, data, previewEnvironment.Sys.Version, resp)
}

func (r *previewEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PreviewEnvironmentModel
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

	params := cm.GetPreviewEnvironmentParams{
		SpaceID:              state.SpaceID.ValueString(),
		PreviewEnvironmentID: state.PreviewEnvironmentID.ValueString(),
	}
	response, err := r.providerData.client.GetPreviewEnvironment(ctx, params)
	tflog.Info(ctx, "preview_environment.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	previewEnvironment, ok := response.(*cm.PreviewEnvironment)
	if !ok {
		if statusCodeResponse, statusOK := response.(cm.StatusCodeResponse); statusOK && statusCodeResponse.GetStatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Content preview platform not found", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, dataDiagnostics := NewPreviewEnvironmentModelFromResponse(ctx, *previewEnvironment)
	resp.Diagnostics.Append(dataDiagnostics...)

	data.Timeouts = state.Timeouts

	var identity PreviewEnvironmentIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", previewEnvironment.Sys.Version)...)
}

func (r *previewEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state PreviewEnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var plan PreviewEnvironmentModel
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

	var currentVersion int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)...)
	request, requestDiagnostics := ToPreviewEnvironmentUpdateData(ctx, path.Empty(), &state, &plan)
	resp.Diagnostics.Append(requestDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutPreviewEnvironmentParams{
		SpaceID:              plan.SpaceID.ValueString(),
		PreviewEnvironmentID: plan.PreviewEnvironmentID.ValueString(),
		XContentfulVersion:   currentVersion,
	}
	response, err := r.providerData.client.PutPreviewEnvironment(ctx, &request, params)
	tflog.Info(ctx, "preview_environment.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	previewEnvironment, ok := response.(*cm.PreviewEnvironment)
	if !ok {
		resp.Diagnostics.AddError("Failed to update content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	responseData, dataDiagnostics := NewPreviewEnvironmentModelFromResponse(ctx, *previewEnvironment)
	resp.Diagnostics.Append(dataDiagnostics...)
	resp.Diagnostics.Append(ValidatePreviewEnvironmentUpdateResponse(ctx, path.Empty(), &state, &plan, &responseData)...)

	var identity PreviewEnvironmentIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", previewEnvironment.Sys.Version)...)
}

func (r *previewEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PreviewEnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Delete(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := r.providerData.client.DeletePreviewEnvironment(ctx, cm.DeletePreviewEnvironmentParams{
		SpaceID:              state.SpaceID.ValueString(),
		PreviewEnvironmentID: state.PreviewEnvironmentID.ValueString(),
	})
	switch response := response.(type) {
	case *cm.NoContent:
		return
	default:
		if statusCodeResponse, ok := response.(cm.StatusCodeResponse); ok && statusCodeResponse.GetStatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Content preview platform already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}

func (r *previewEnvironmentResource) setCreateState(ctx context.Context, data PreviewEnvironmentModel, version int, resp *resource.CreateResponse) {
	var identity PreviewEnvironmentIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}
