package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*extensionResource)(nil)
	_ resource.ResourceWithConfigure   = (*extensionResource)(nil)
	_ resource.ResourceWithIdentity    = (*extensionResource)(nil)
	_ resource.ResourceWithImportState = (*extensionResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*extensionResource)(nil)
)

//nolint:ireturn
func NewExtensionResource() resource.Resource {
	return &extensionResource{}
}

type extensionResource struct {
	providerData ContentfulProviderData
}

func (r *extensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extension"
}

func (r *extensionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ExtensionResourceSchema(ctx)
}

func (r *extensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *extensionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":       identityschema.StringAttribute{RequiredForImport: true},
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"extension_id":   identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *extensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("extension_id"),
	}, req, resp)
}

func (r *extensionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	srcPath := path.Root("extension").AtName("src")
	srcdocPath := path.Root("extension").AtName("srcdoc")

	var configSrc, configSrcdoc types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, srcPath, &configSrc)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, srcdocPath, &configSrcdoc)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if configSrc.IsUnknown() {
		if configSrcdoc.IsNull() {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcdocPath, types.StringNull())...)
		}

		return
	}

	if configSrcdoc.IsUnknown() {
		if configSrc.IsNull() {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcPath, types.StringNull())...)
		}

		return
	}

	if !configSrc.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcdocPath, types.StringNull())...)

		return
	}

	if !configSrcdoc.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcPath, types.StringNull())...)

		return
	}

	if req.State.Raw.IsNull() {
		resp.Diagnostics.AddAttributeError(
			srcPath,
			"Missing extension source",
			"Exactly one of extension.src or extension.srcdoc must be configured when creating an extension.",
		)

		return
	}

	var stateSrc, stateSrcdoc types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, srcPath, &stateSrc)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, srcdocPath, &stateSrcdoc)...)

	if resp.Diagnostics.HasError() {
		return
	}

	preserveExtensionSourcePlan(ctx, stateSrc, stateSrcdoc, resp)
}

func preserveExtensionSourcePlan(
	ctx context.Context,
	stateSrc types.String,
	stateSrcdoc types.String,
	resp *resource.ModifyPlanResponse,
) {
	srcPath := path.Root("extension").AtName("src")
	srcdocPath := path.Root("extension").AtName("srcdoc")

	// Older provider versions projected an absent source as an empty string.
	// An empty src is invalid in Contentful, so it can only be that legacy
	// placeholder. A valid src paired with an empty srcdoc is likewise the old
	// absent-sibling encoding; two genuine source values remain invalid.
	stateSrcKnown := !stateSrc.IsNull() && !stateSrc.IsUnknown()
	stateSrcdocKnown := !stateSrcdoc.IsNull() && !stateSrcdoc.IsUnknown()

	switch {
	case stateSrcKnown && stateSrc.ValueString() != "" && (!stateSrcdocKnown || stateSrcdoc.ValueString() == ""):
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcPath, stateSrc)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcdocPath, types.StringNull())...)
	case (!stateSrcKnown || stateSrc.ValueString() == "") && stateSrcdocKnown:
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcPath, types.StringNull())...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, srcdocPath, stateSrcdoc)...)
	case stateSrcKnown && stateSrcdocKnown:
		resp.Diagnostics.AddAttributeError(
			srcdocPath,
			"Conflicting extension sources",
			"The prior state contains both extension.src and extension.srcdoc, but a Contentful extension update requires exactly one source.",
		)
	default:
		resp.Diagnostics.AddAttributeError(
			srcPath,
			"Missing extension source",
			"The prior state does not contain the extension.src or extension.srcdoc value required for a Contentful update.",
		)
	}
}

func (r *extensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   ExtensionModel
		config ExtensionModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToExtensionData(config, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

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

	params := cm.PutExtensionParams{
		SpaceID:       plan.SpaceID.ValueString(),
		EnvironmentID: plan.EnvironmentID.ValueString(),
		ExtensionID:   plan.ExtensionID.ValueString(),
	}

	response, err := r.providerData.client.PutExtension(ctx, &request, params)

	tflog.Info(ctx, "extension.create", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data ExtensionModel

	switch response := response.(type) {
	case *cm.ExtensionStatusCode:
		responseModel, responseModelDiags := NewExtensionModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	var identityModel ExtensionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *extensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExtensionModel

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

	params := cm.GetExtensionParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ExtensionID:   state.ExtensionID.ValueString(),
	}

	response, err := r.providerData.client.GetExtension(ctx, params)

	tflog.Info(ctx, "extension.read", map[string]any{
		"params": params,
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	version := 0

	var data ExtensionModel

	switch response := response.(type) {
	case *cm.Extension:
		responseModel, responseModelDiags := NewExtensionModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	var identityModel ExtensionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *extensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   ExtensionModel
		config ExtensionModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToExtensionData(config, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

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

	params := cm.PutExtensionParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		ExtensionID:        plan.ExtensionID.ValueString(),
		XContentfulVersion: version,
	}

	response, err := r.providerData.client.PutExtension(ctx, &request, params)

	tflog.Info(ctx, "extension.update", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data ExtensionModel

	switch response := response.(type) {
	case *cm.ExtensionStatusCode:
		responseModel, responseModelDiags := NewExtensionModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	var identityModel ExtensionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

//nolint:dupl
func (r *extensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExtensionModel

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

	params := cm.DeleteExtensionParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ExtensionID:   state.ExtensionID.ValueString(),
	}

	response, err := r.providerData.client.DeleteExtension(ctx, params)

	tflog.Info(ctx, "extension.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Extension already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
