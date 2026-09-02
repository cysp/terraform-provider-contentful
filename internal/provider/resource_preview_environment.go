package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func previewEnvironmentIdentityAttributeNames() []string {
	return []string{"space_id", "preview_environment_id"}
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
	resp.IdentitySchema = resourceIdentitySchema(previewEnvironmentIdentityAttributeNames())
}

func (r *previewEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, previewEnvironmentIdentityAttributeNames(), req, resp)
}

func (r *previewEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   PreviewEnvironmentModel
		config PreviewEnvironmentModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	request, requestDiagnostics := plan.ToPreviewEnvironmentData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiagnostics...)

	spaceID, spaceIDDiagnostics := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceIDDiagnostics...)

	selectedID := ""
	selectedIDConfigured := !config.PreviewEnvironmentID.IsNull() && !config.PreviewEnvironmentID.IsUnknown()

	if config.PreviewEnvironmentID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("preview_environment_id"),
			"Unexpected unknown configured preview environment ID",
			"The preview environment ID configuration must be known before the provider can choose the Contentful create endpoint.",
		)
	} else if selectedIDConfigured {
		selectedIDValue, selectedIDDiagnostics := requestRequiredString(plan.PreviewEnvironmentID, path.Root("preview_environment_id"))
		resp.Diagnostics.Append(selectedIDDiagnostics...)

		selectedID = selectedIDValue
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		response any
		err      error
	)
	if selectedIDConfigured {
		response, err = r.providerData.client.PutPreviewEnvironment(ctx, &request, cm.PutPreviewEnvironmentParams{
			SpaceID:              spaceID,
			PreviewEnvironmentID: selectedID,
			XContentfulVersion:   0,
		})
	} else {
		createRequest := cm.NewPreviewEnvironmentCreateData(request)
		response, err = r.providerData.client.CreatePreviewEnvironment(ctx, &createRequest, cm.CreatePreviewEnvironmentParams{
			SpaceID: spaceID,
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

	ownedIdentity := PreviewEnvironmentIdentityModel{
		SpaceID:              plan.SpaceID,
		PreviewEnvironmentID: plan.PreviewEnvironmentID,
	}
	if !selectedIDConfigured {
		ownedIdentity.PreviewEnvironmentID = config.PreviewEnvironmentID
	}

	data, dataDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		ctx,
		*previewEnvironment,
		plan,
		ownedIdentity,
	)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(
		ctx,
		resp.Identity,
		&resp.State,
		previewEnvironmentIdentityAttributeNames(),
		&data,
	)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", previewEnvironment.Sys.Version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *previewEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PreviewEnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	if contentfulResponseIsNotFound(response) {
		resp.Diagnostics.AddWarning("Content preview platform not found", util.ErrorDetailFromContentfulManagementResponse(response, err))
		resp.State.RemoveResource(ctx)

		return
	}

	previewEnvironment, ok := response.(*cm.PreviewEnvironment)
	if !ok {
		resp.Diagnostics.AddError("Failed to read content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, dataDiagnostics := NewPreviewEnvironmentModelFromResponse(ctx, *previewEnvironment)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(
		ctx,
		resp.Identity,
		&resp.State,
		previewEnvironmentIdentityAttributeNames(),
		&data,
	)...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	request, requestDiagnostics := ToPreviewEnvironmentUpdateData(ctx, path.Empty(), &state, &plan)
	resp.Diagnostics.Append(requestDiagnostics...)

	spaceID, spaceIDDiagnostics := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceIDDiagnostics...)

	previewEnvironmentID, previewEnvironmentIDDiagnostics := requestRequiredString(
		plan.PreviewEnvironmentID,
		path.Root("preview_environment_id"),
	)
	resp.Diagnostics.Append(previewEnvironmentIDDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion, versionDiagnostics := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutPreviewEnvironmentParams{
		SpaceID:              spaceID,
		PreviewEnvironmentID: previewEnvironmentID,
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

	data, dataDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		ctx,
		*previewEnvironment,
		plan,
		plan.PreviewEnvironmentIdentityModel,
	)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(
		ctx,
		resp.Identity,
		&resp.State,
		previewEnvironmentIdentityAttributeNames(),
		&data,
	)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", previewEnvironment.Sys.Version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *previewEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PreviewEnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.DeletePreviewEnvironment(ctx, cm.DeletePreviewEnvironmentParams{
		SpaceID:              state.SpaceID.ValueString(),
		PreviewEnvironmentID: state.PreviewEnvironmentID.ValueString(),
	})
	switch response := response.(type) {
	case *cm.NoContent:
		return
	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Content preview platform already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete content preview platform", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}
