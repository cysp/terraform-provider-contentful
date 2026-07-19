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
	var plan TaxonomyConceptModel
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

	request, requestDiags := plan.ToRequest(ctx)
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

	validationErr := validateTaxonomyConceptResponse(request, *concept)
	if validationErr != nil {
		resp.Diagnostics.AddError("Contentful normalized taxonomy concept configuration", validationErr.Error())

		return
	}

	r.setCreateState(ctx, plan, *concept, resp)
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

	data, modelDiags := NewTaxonomyConceptModelFromResponse(ctx, *concept)
	resp.Diagnostics.Append(modelDiags...)
	preserveConfiguredLabelMapShape(&data, state)

	data.Timeouts = state.Timeouts

	var identity TaxonomyConceptIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *taxonomyConceptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TaxonomyConceptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

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

	getParams := cm.GetTaxonomyConceptParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptID: plan.ConceptID.ValueString()}
	currentResponse, err := r.providerData.client.GetTaxonomyConcept(ctx, getParams)

	current, ok := currentResponse.(*cm.TaxonomyConcept)
	if !ok {
		resp.Diagnostics.AddError("Failed to refresh taxonomy concept before update", util.ErrorDetailFromContentfulManagementResponse(currentResponse, err))

		return
	}

	request, requestDiags := plan.ToRequest(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	patch, patchErr := taxonomyPatch(taxonomyConceptRequestFromResponse(*current), request)
	if patchErr != nil {
		resp.Diagnostics.AddError("Failed to build taxonomy concept update", patchErr.Error())

		return
	}

	if len(patch) == 0 {
		data, modelDiags := NewTaxonomyConceptModelFromResponse(ctx, *current)
		resp.Diagnostics.Append(modelDiags...)
		preserveConfiguredLabelMapShape(&data, plan)

		data.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	params := cm.PatchTaxonomyConceptParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptID: plan.ConceptID.ValueString(), XContentfulVersion: current.Sys.Version}
	response, err := r.providerData.client.PatchTaxonomyConcept(ctx, patch, params)
	tflog.Info(ctx, "taxonomy_concept.update", map[string]any{"params": params, "response": response, "err": err})

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		resp.Diagnostics.AddError("Failed to update taxonomy concept", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	validationErr := validateTaxonomyConceptResponse(request, *concept)
	if validationErr != nil {
		resp.Diagnostics.AddError("Contentful normalized taxonomy concept configuration", validationErr.Error())

		return
	}

	data, modelDiags := NewTaxonomyConceptModelFromResponse(ctx, *concept)
	resp.Diagnostics.Append(modelDiags...)
	preserveConfiguredLabelMapShape(&data, plan)

	data.Timeouts = plan.Timeouts

	var identity TaxonomyConceptIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	getParams := cm.GetTaxonomyConceptParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptID: state.ConceptID.ValueString()}
	currentResponse, err := r.providerData.client.GetTaxonomyConcept(ctx, getParams)

	current, ok := currentResponse.(*cm.TaxonomyConcept)
	if !ok {
		if status, ok := currentResponse.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
			return
		}

		resp.Diagnostics.AddError("Failed to refresh taxonomy concept before deletion", util.ErrorDetailFromContentfulManagementResponse(currentResponse, err))

		return
	}

	params := cm.DeleteTaxonomyConceptParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptID: state.ConceptID.ValueString(), XContentfulVersion: current.Sys.Version}
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

func (r *taxonomyConceptResource) setCreateState(ctx context.Context, prior TaxonomyConceptModel, concept cm.TaxonomyConcept, resp *resource.CreateResponse) {
	data, modelDiags := NewTaxonomyConceptModelFromResponse(ctx, concept)
	resp.Diagnostics.Append(modelDiags...)
	preserveConfiguredLabelMapShape(&data, prior)

	data.Timeouts = prior.Timeouts

	var identity TaxonomyConceptIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
