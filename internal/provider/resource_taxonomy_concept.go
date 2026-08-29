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
	_ resource.Resource                = (*taxonomyConceptResource)(nil)
	_ resource.ResourceWithConfigure   = (*taxonomyConceptResource)(nil)
	_ resource.ResourceWithIdentity    = (*taxonomyConceptResource)(nil)
	_ resource.ResourceWithImportState = (*taxonomyConceptResource)(nil)
)

//nolint:ireturn
func NewTaxonomyConceptResource() resource.Resource { return &taxonomyConceptResource{} }

type taxonomyConceptResource struct{ providerData ContentfulProviderData }

func (r *taxonomyConceptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_taxonomy_concept"
}

func (r *taxonomyConceptResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TaxonomyConceptResourceSchema(ctx)
}

func (r *taxonomyConceptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *taxonomyConceptResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"organization_id": identityschema.StringAttribute{RequiredForImport: true},
		"concept_id":      identityschema.StringAttribute{RequiredForImport: true},
	}}
}

func (r *taxonomyConceptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{path.Root("organization_id"), path.Root("concept_id")}, req, resp)
}

func (r *taxonomyConceptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config, plan TaxonomyConceptModel
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

	prepared, prepareDiags := prepareTaxonomyConceptMutation(config, plan)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := prepared.planRequest(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutTaxonomyConceptParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptID: plan.ConceptID.ValueString()}
	response, err := r.providerData.client.PutTaxonomyConcept(ctx, &request, params)
	tflog.Info(ctx, "taxonomy_concept.create", map[string]any{"params": params, "response": response, "err": err})

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		resp.Diagnostics.AddError("Failed to create taxonomy concept", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, responseDiags, consistencyDiags := prepared.ProjectResponse(ctx, *concept)
	resp.Diagnostics.Append(responseDiags...)

	statePublished := false

	if !resp.Diagnostics.HasError() {
		var identity TaxonomyConceptIdentityModel
		resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)
			statePublished = !resp.Diagnostics.HasError()
		}
	}

	if statePublished {
		resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", concept.Sys.Version)...)
	}

	// A successful remote mutation can disagree with the plan. Publish its
	// recovery state before returning the consistency diagnostics.
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *taxonomyConceptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TaxonomyConceptModel
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

	params := cm.GetTaxonomyConceptParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptID: state.ConceptID.ValueString()}
	response, err := r.providerData.client.GetTaxonomyConcept(ctx, params)
	tflog.Info(ctx, "taxonomy_concept.read", map[string]any{"params": params, "response": response, "err": err})

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read taxonomy concept", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, modelDiags := newTaxonomyConceptRefreshState(ctx, state, *concept)
	resp.Diagnostics.Append(modelDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var identity TaxonomyConceptIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", concept.Sys.Version)...)
}

func (r *taxonomyConceptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state TaxonomyConceptModel
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

	prepared, prepareDiags := prepareTaxonomyConceptMutation(config, plan)
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

		var identity TaxonomyConceptIdentityModel
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

	params := cm.PatchTaxonomyConceptParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptID: plan.ConceptID.ValueString(), XContentfulVersion: priorStateVersion}
	response, err := r.providerData.client.PatchTaxonomyConcept(ctx, patch, params)
	tflog.Info(ctx, "taxonomy_concept.update", map[string]any{"params": params, "response": response, "err": err})

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		resp.Diagnostics.AddError("Failed to update taxonomy concept", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data, responseDiags, consistencyDiags := prepared.ProjectResponse(ctx, *concept)
	resp.Diagnostics.Append(responseDiags...)

	statePublished := false

	if !resp.Diagnostics.HasError() {
		var identity TaxonomyConceptIdentityModel
		resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identity, &data)...)
			statePublished = !resp.Diagnostics.HasError()
		}
	}

	if statePublished {
		resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", concept.Sys.Version)...)
	}

	// A successful remote mutation can disagree with the plan. Publish its
	// recovery state before returning the consistency diagnostics.
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *taxonomyConceptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TaxonomyConceptModel
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

	organizationID, conceptID := state.OrganizationID.ValueString(), state.ConceptID.ValueString()
	priorStateVersion, versionFound, versionDiags := optionalPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !versionFound {
		params := cm.GetTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: conceptID}
		response, err := r.providerData.client.GetTaxonomyConcept(ctx, params)
		tflog.Info(ctx, "taxonomy_concept.delete.read", map[string]any{"params": params, "response": response, "err": err})

		concept, ok := response.(*cm.TaxonomyConcept)
		if !ok {
			if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
				return
			}

			resp.Diagnostics.AddError("Failed to read taxonomy concept before deletion", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		priorStateVersion = concept.Sys.Version
	}

	params := cm.DeleteTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: conceptID, XContentfulVersion: priorStateVersion}
	response, err := r.providerData.client.DeleteTaxonomyConcept(ctx, params)
	tflog.Info(ctx, "taxonomy_concept.delete", map[string]any{"params": params, "response": response, "err": err})

	if _, ok := response.(*cm.NoContent); ok {
		return
	}

	if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
		return
	}

	resp.Diagnostics.AddError("Failed to delete taxonomy concept", util.ErrorDetailFromContentfulManagementResponse(response, err))
}
