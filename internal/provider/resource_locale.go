package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*localeResource)(nil)
	_ resource.ResourceWithConfigure   = (*localeResource)(nil)
	_ resource.ResourceWithIdentity    = (*localeResource)(nil)
	_ resource.ResourceWithImportState = (*localeResource)(nil)
)

//nolint:ireturn
func NewLocaleResource() resource.Resource {
	return &localeResource{}
}

type localeResource struct {
	providerData ContentfulProviderData
}

func localeIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id", "locale_id"}
}

func (r *localeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locale"
}

func (r *localeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = LocaleResourceSchema(ctx)
}

func (r *localeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *localeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(localeIdentityAttributeNames())
}

func (r *localeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, localeIdentityAttributeNames(), req, resp)
}

func (r *localeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LocaleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	spaceID, spaceDiags := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceDiags...)

	environmentID, environmentDiags := requestRequiredString(plan.EnvironmentID, path.Root("environment_id"))
	resp.Diagnostics.Append(environmentDiags...)

	request, requestDiagnostics := plan.ToLocaleData()
	resp.Diagnostics.Append(requestDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateLocaleParams{SpaceID: spaceID, EnvironmentID: environmentID}
	response, err := r.providerData.client.CreateLocale(ctx, &request, params)

	tflog.Info(ctx, "locale.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	locale, ok := response.(*cm.LocaleStatusCode)
	if !ok {
		resp.Diagnostics.AddError("Failed to create locale", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, consistencyDiagnostics := ReconcileLocaleMutationResponse(locale.Response, plan)

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, localeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setLocalePrivateVersion(ctx, resp.Private, locale.Response.Sys.Version)...)
	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *localeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LocaleModel

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

	params := cm.GetLocaleParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		LocaleID:      state.LocaleID.ValueString(),
	}

	response, err := r.providerData.client.GetLocale(ctx, params)

	tflog.Info(ctx, "locale.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	if contentfulResponseIsNotFound(response) {
		resp.Diagnostics.AddWarning("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))
		resp.State.RemoveResource(ctx)

		return
	}

	locale, ok := response.(*cm.Locale)
	if !ok {
		resp.Diagnostics.AddError("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data := NewLocaleResourceModelFromResponse(*locale)

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, localeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setLocalePrivateVersion(ctx, resp.Private, locale.Sys.Version)...)
}

func (r *localeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LocaleModel

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

	spaceID, spaceDiags := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceDiags...)

	environmentID, environmentDiags := requestRequiredString(plan.EnvironmentID, path.Root("environment_id"))
	resp.Diagnostics.Append(environmentDiags...)

	localeID, localeDiags := requestRequiredString(plan.LocaleID, path.Root("locale_id"))
	resp.Diagnostics.Append(localeDiags...)

	request, requestDiagnostics := plan.ToLocaleData()
	resp.Diagnostics.Append(requestDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiagnostics := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutLocaleParams{
		SpaceID:            spaceID,
		EnvironmentID:      environmentID,
		LocaleID:           localeID,
		XContentfulVersion: version,
	}

	response, err := r.providerData.client.PutLocale(ctx, &request, params)

	tflog.Info(ctx, "locale.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	locale, ok := response.(*cm.LocaleStatusCode)
	if !ok {
		resp.Diagnostics.AddError("Failed to update locale", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, consistencyDiagnostics := ReconcileLocaleMutationResponse(locale.Response, plan)

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, localeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setLocalePrivateVersion(ctx, resp.Private, locale.Response.Sys.Version)...)
	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *localeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LocaleModel

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

	params := cm.DeleteLocaleParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		LocaleID:      state.LocaleID.ValueString(),
	}

	response, err := r.providerData.client.DeleteLocale(ctx, params)

	tflog.Info(ctx, "locale.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	if _, ok := response.(*cm.NoContent); ok || contentfulResponseIsNotFound(response) {
		return
	}

	resp.Diagnostics.AddError("Failed to delete locale", util.ErrorDetailFromContentfulManagementResponse(response, err))
}

func setLocalePrivateVersion(ctx context.Context, private PrivateProviderData, version cm.OptInt) diag.Diagnostics {
	if value, ok := version.Get(); ok {
		return SetPrivateProviderData(ctx, private, "version", value)
	}

	// Discard the old version so a partial mutation cannot reuse it.
	diagnostics := private.SetKey(ctx, "version", nil)
	diagnostics.AddError("Missing locale version", "Contentful returned a locale without sys.version. Refresh the resource successfully before updating it.")

	return diagnostics
}
