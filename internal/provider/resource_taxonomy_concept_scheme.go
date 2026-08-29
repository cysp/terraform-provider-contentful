//nolint:dupl
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
	_ resource.Resource                = (*taxonomyConceptSchemeResource)(nil)
	_ resource.ResourceWithConfigure   = (*taxonomyConceptSchemeResource)(nil)
	_ resource.ResourceWithIdentity    = (*taxonomyConceptSchemeResource)(nil)
	_ resource.ResourceWithImportState = (*taxonomyConceptSchemeResource)(nil)
)

//nolint:ireturn
func NewTaxonomyConceptSchemeResource() resource.Resource { return &taxonomyConceptSchemeResource{} }

type taxonomyConceptSchemeResource struct{ providerData ContentfulProviderData }

func (r *taxonomyConceptSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_taxonomy_concept_scheme"
}

func (r *taxonomyConceptSchemeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TaxonomyConceptSchemeResourceSchema(ctx)
}

func (r *taxonomyConceptSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *taxonomyConceptSchemeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
		"concept_scheme_id": identityschema.StringAttribute{RequiredForImport: true},
	}}
}

func (r *taxonomyConceptSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{path.Root("organization_id"), path.Root("concept_scheme_id")}, req, resp)
}

func (r *taxonomyConceptSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config, plan TaxonomyConceptSchemeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiags := plan.Timeouts.Create(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prepared, prepareDiags := prepareTaxonomyConceptSchemeMutation(config, plan)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := prepared.planRequest(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutTaxonomyConceptSchemeParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptSchemeID: plan.ConceptSchemeID.ValueString()}
	response, err := r.providerData.client.PutTaxonomyConceptScheme(ctx, &request, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.create", map[string]any{"params": params, "response": response, "err": err})

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		resp.Diagnostics.AddError("Failed to create taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, responseDiags, consistencyDiags := prepared.ProjectResponse(ctx, *scheme)
	resp.Diagnostics.Append(responseDiags...)

	statePublished := false

	if !resp.Diagnostics.HasError() {
		var identity TaxonomyConceptSchemeIdentityModel
		resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)
			statePublished = !resp.Diagnostics.HasError()
		}
	}

	if statePublished {
		resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", scheme.Sys.Version)...)
	}

	// A successful remote mutation can disagree with the plan. Publish its
	// recovery state before returning the consistency diagnostics.
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *taxonomyConceptSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TaxonomyConceptSchemeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiags := state.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, max(timeout, minimumStoredResourceOperationTimeout))
	defer cancel()

	params := cm.GetTaxonomyConceptSchemeParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptSchemeID: state.ConceptSchemeID.ValueString()}
	response, err := r.providerData.client.GetTaxonomyConceptScheme(ctx, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.read", map[string]any{"params": params, "response": response, "err": err})

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, modelDiags := newTaxonomyConceptSchemeRefreshState(ctx, state, *scheme)
	resp.Diagnostics.Append(modelDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var identity TaxonomyConceptSchemeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", scheme.Sys.Version)...)
}

func (r *taxonomyConceptSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state TaxonomyConceptSchemeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiags := plan.Timeouts.Update(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prepared, prepareDiags := prepareTaxonomyConceptSchemeMutation(config, plan)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	patch, patchDiags := prepared.PatchFromState(ctx, state)
	resp.Diagnostics.Append(patchDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(patch) == 0 {
		data := prepared.NoopState(state)

		var identity TaxonomyConceptSchemeIdentityModel
		resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)
		}

		return
	}

	priorStateVersion, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PatchTaxonomyConceptSchemeParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptSchemeID: plan.ConceptSchemeID.ValueString(), XContentfulVersion: priorStateVersion}
	response, err := r.providerData.client.PatchTaxonomyConceptScheme(ctx, patch, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.update", map[string]any{"params": params, "response": response, "err": err})

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		resp.Diagnostics.AddError("Failed to update taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, responseDiags, consistencyDiags := prepared.ProjectResponse(ctx, *scheme)
	resp.Diagnostics.Append(responseDiags...)

	statePublished := false

	if !resp.Diagnostics.HasError() {
		var identity TaxonomyConceptSchemeIdentityModel
		resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)
			statePublished = !resp.Diagnostics.HasError()
		}
	}

	if statePublished {
		resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", scheme.Sys.Version)...)
	}

	// A successful remote mutation can disagree with the plan. Publish its
	// recovery state before returning the consistency diagnostics.
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *taxonomyConceptSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TaxonomyConceptSchemeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiags := state.Timeouts.Delete(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, max(timeout, minimumStoredResourceOperationTimeout))
	defer cancel()

	organizationID, schemeID := state.OrganizationID.ValueString(), state.ConceptSchemeID.ValueString()
	priorStateVersion, versionFound, versionDiags := optionalPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !versionFound {
		params := cm.GetTaxonomyConceptSchemeParams{OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID}
		response, err := r.providerData.client.GetTaxonomyConceptScheme(ctx, params)
		tflog.Info(ctx, "taxonomy_concept_scheme.delete.read", map[string]any{"params": params, "response": response, "err": err})

		scheme, ok := response.(*cm.TaxonomyConceptScheme)
		if !ok {
			if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
				return
			}

			resp.Diagnostics.AddError("Failed to read taxonomy concept scheme before deletion", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		priorStateVersion = scheme.Sys.Version
	}

	params := cm.DeleteTaxonomyConceptSchemeParams{OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID, XContentfulVersion: priorStateVersion}
	response, err := r.providerData.client.DeleteTaxonomyConceptScheme(ctx, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.delete", map[string]any{"params": params, "response": response, "err": err})

	if _, ok := response.(*cm.NoContent); ok {
		return
	}

	if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
		return
	}

	resp.Diagnostics.AddError("Failed to delete taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))
}
