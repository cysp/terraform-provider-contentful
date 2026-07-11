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
	var plan TaxonomyConceptSchemeModel
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

	params := cm.PutTaxonomyConceptSchemeParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptSchemeID: plan.ConceptSchemeID.ValueString()}
	response, err := r.providerData.client.PutTaxonomyConceptScheme(ctx, &request, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.create", map[string]any{"params": params, "response": response, "err": err})

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		resp.Diagnostics.AddError("Failed to create taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	validationErr := validateTaxonomyConceptSchemeResponse(request, *scheme)
	if validationErr != nil {
		resp.Diagnostics.AddError("Contentful normalized taxonomy concept scheme configuration", validationErr.Error())

		return
	}

	data, modelDiags := NewTaxonomyConceptSchemeModelFromResponse(ctx, *scheme)
	resp.Diagnostics.Append(modelDiags...)

	data.Timeouts = plan.Timeouts

	var identity TaxonomyConceptSchemeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	data, modelDiags := NewTaxonomyConceptSchemeModelFromResponse(ctx, *scheme)
	resp.Diagnostics.Append(modelDiags...)

	data.Timeouts = state.Timeouts

	var identity TaxonomyConceptSchemeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *taxonomyConceptSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TaxonomyConceptSchemeModel
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

	getParams := cm.GetTaxonomyConceptSchemeParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptSchemeID: plan.ConceptSchemeID.ValueString()}
	currentResponse, err := r.providerData.client.GetTaxonomyConceptScheme(ctx, getParams)

	current, ok := currentResponse.(*cm.TaxonomyConceptScheme)
	if !ok {
		resp.Diagnostics.AddError("Failed to refresh taxonomy concept scheme before update", util.ErrorDetailFromContentfulManagementResponse(currentResponse, err))

		return
	}

	request, requestDiags := plan.ToRequest(ctx)
	resp.Diagnostics.Append(requestDiags...)

	patch, patchErr := taxonomyPatch(taxonomyConceptSchemeRequestFromResponse(*current), request)
	if patchErr != nil {
		resp.Diagnostics.AddError("Failed to build taxonomy concept scheme update", patchErr.Error())

		return
	}

	if len(patch) == 0 {
		data, modelDiags := NewTaxonomyConceptSchemeModelFromResponse(ctx, *current)
		resp.Diagnostics.Append(modelDiags...)

		data.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	params := cm.PatchTaxonomyConceptSchemeParams{OrganizationID: plan.OrganizationID.ValueString(), TaxonomyConceptSchemeID: plan.ConceptSchemeID.ValueString(), XContentfulVersion: current.Sys.Version}
	response, err := r.providerData.client.PatchTaxonomyConceptScheme(ctx, patch, params)
	tflog.Info(ctx, "taxonomy_concept_scheme.update", map[string]any{"params": params, "response": response, "err": err})

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		resp.Diagnostics.AddError("Failed to update taxonomy concept scheme", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	validationErr := validateTaxonomyConceptSchemeResponse(request, *scheme)
	if validationErr != nil {
		resp.Diagnostics.AddError("Contentful normalized taxonomy concept scheme configuration", validationErr.Error())

		return
	}

	data, modelDiags := NewTaxonomyConceptSchemeModelFromResponse(ctx, *scheme)
	resp.Diagnostics.Append(modelDiags...)

	data.Timeouts = plan.Timeouts

	var identity TaxonomyConceptSchemeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identity, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	getParams := cm.GetTaxonomyConceptSchemeParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptSchemeID: state.ConceptSchemeID.ValueString()}
	currentResponse, err := r.providerData.client.GetTaxonomyConceptScheme(ctx, getParams)

	current, ok := currentResponse.(*cm.TaxonomyConceptScheme)
	if !ok {
		if status, ok := currentResponse.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
			return
		}

		resp.Diagnostics.AddError("Failed to refresh taxonomy concept scheme before deletion", util.ErrorDetailFromContentfulManagementResponse(currentResponse, err))

		return
	}

	params := cm.DeleteTaxonomyConceptSchemeParams{OrganizationID: state.OrganizationID.ValueString(), TaxonomyConceptSchemeID: state.ConceptSchemeID.ValueString(), XContentfulVersion: current.Sys.Version}
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
