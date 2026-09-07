package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type taxonomyConceptSchemeMutationOwnership struct {
	topConceptIDs bool
	conceptIDs    bool
}

type preparedTaxonomyConceptSchemeMutation struct {
	plan      TaxonomyConceptSchemeModel
	ownership taxonomyConceptSchemeMutationOwnership
}

func prepareTaxonomyConceptSchemeMutation(config, plan TaxonomyConceptSchemeModel) (preparedTaxonomyConceptSchemeMutation, diag.Diagnostics) {
	prepared := preparedTaxonomyConceptSchemeMutation{plan: plan}
	diags := diag.Diagnostics{}
	prepared.ownership.topConceptIDs = taxonomyCollectionOwnership(config.TopConceptIDs, plan.TopConceptIDs, path.Root("top_concept_ids"), &diags)
	prepared.ownership.conceptIDs = taxonomyCollectionOwnership(config.ConceptIDs, plan.ConceptIDs, path.Root("concept_ids"), &diags)

	return prepared, diags
}

func (prepared preparedTaxonomyConceptSchemeMutation) PatchFromState(ctx context.Context, state TaxonomyConceptSchemeModel) (cm.TaxonomyPatch, diag.Diagnostics) {
	desired, desiredDiags := prepared.planRequest(ctx)
	diags := diag.Diagnostics{}
	diags.Append(desiredDiags...)

	if diags.HasError() {
		return nil, diags
	}

	current, currentDiags := prepared.stateRequest(ctx, state)
	diags.Append(currentDiags...)

	if diags.HasError() {
		return nil, diags
	}

	patch, err := taxonomyPatch(&current, &desired)
	if err != nil {
		diags.Append(taxonomyPatchErrorDiagnostics("taxonomy concept scheme", err)...)
	}

	return patch, diags
}

func (prepared preparedTaxonomyConceptSchemeMutation) NoopState(state TaxonomyConceptSchemeModel) TaxonomyConceptSchemeModel {
	data := state
	data.IDIdentityModel = prepared.plan.IDIdentityModel
	data.TaxonomyConceptSchemeIdentityModel = prepared.plan.TaxonomyConceptSchemeIdentityModel
	data.URI = prepared.plan.URI
	data.PrefLabel = prepared.plan.PrefLabel

	data.Definition = prepared.plan.Definition
	if taxonomyKnownPlanValue(prepared.plan.TopConceptIDs) {
		data.TopConceptIDs = prepared.plan.TopConceptIDs
	}

	if taxonomyKnownPlanValue(prepared.plan.ConceptIDs) {
		data.ConceptIDs = prepared.plan.ConceptIDs
	}

	data.Timeouts = prepared.plan.Timeouts

	return data
}

func (prepared preparedTaxonomyConceptSchemeMutation) ProjectResponse(ctx context.Context, response cm.TaxonomyConceptScheme) (TaxonomyConceptSchemeModel, diag.Diagnostics, diag.Diagnostics) {
	data, responseDiags := NewTaxonomyConceptSchemeModelFromResponse(ctx, response)

	data.Timeouts = prepared.plan.Timeouts
	if responseDiags.HasError() {
		return data, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}
	mismatch := taxonomyMutationIdentityConsistency("taxonomy concept scheme", path.Root("organization_id"), prepared.plan.OrganizationID, data.OrganizationID, &consistencyDiags)
	mismatch = taxonomyMutationIdentityConsistency("taxonomy concept scheme", path.Root("concept_scheme_id"), prepared.plan.ConceptSchemeID, data.ConceptSchemeID, &consistencyDiags) || mismatch
	data.OrganizationID, data.ConceptSchemeID = prepared.plan.OrganizationID, prepared.plan.ConceptSchemeID
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(prepared.plan.OrganizationID.ValueString(), prepared.plan.ConceptSchemeID.ValueString())

	var projects []func()

	compare := func(name string, valuePath path.Path, planned, remote attr.Value, project func()) {
		if compareTaxonomyMutationValue("taxonomy concept scheme", name, valuePath, planned, remote, &consistencyDiags) {
			projects = append(projects, project)
		} else {
			mismatch = true
		}
	}
	compareNullableLocalizedString := func(name string, valuePath path.Path, planned, remote types.Map, project func()) {
		if compareTaxonomyNullableLocalizedStringMutationValue("taxonomy concept scheme", name, valuePath, planned, remote, &consistencyDiags) {
			projects = append(projects, project)
		} else {
			mismatch = true
		}
	}

	compare("uri", path.Root("uri"), prepared.plan.URI, data.URI, func() { data.URI = prepared.plan.URI })
	compare("pref_label", path.Root("pref_label"), prepared.plan.PrefLabel, data.PrefLabel, func() { data.PrefLabel = prepared.plan.PrefLabel })
	compareNullableLocalizedString("definition", path.Root("definition"), prepared.plan.Definition, data.Definition, func() { data.Definition = prepared.plan.Definition })

	if taxonomyKnownPlanValue(prepared.plan.TopConceptIDs) {
		compare("top_concept_ids", path.Root("top_concept_ids"), prepared.plan.TopConceptIDs, data.TopConceptIDs, func() { data.TopConceptIDs = prepared.plan.TopConceptIDs })
	}

	if taxonomyKnownPlanValue(prepared.plan.ConceptIDs) {
		compare("concept_ids", path.Root("concept_ids"), prepared.plan.ConceptIDs, data.ConceptIDs, func() { data.ConceptIDs = prepared.plan.ConceptIDs })
	}

	if !mismatch {
		for _, project := range projects {
			project()
		}
	}

	return data, responseDiags, consistencyDiags
}

func (prepared preparedTaxonomyConceptSchemeMutation) planRequest(ctx context.Context) (cm.TaxonomyConceptSchemeRequest, diag.Diagnostics) {
	model := prepared.plan
	if !prepared.ownership.topConceptIDs && !taxonomyKnownPlanValue(model.TopConceptIDs) {
		model.TopConceptIDs = types.ListNull(types.StringType)
	}

	if !prepared.ownership.conceptIDs && !taxonomyKnownPlanValue(model.ConceptIDs) {
		model.ConceptIDs = types.ListNull(types.StringType)
	}

	return taxonomyConceptSchemeRequestFromModel(ctx, model)
}

func (prepared preparedTaxonomyConceptSchemeMutation) stateRequest(ctx context.Context, state TaxonomyConceptSchemeModel) (cm.TaxonomyConceptSchemeRequest, diag.Diagnostics) {
	if !prepared.ownership.topConceptIDs {
		state.TopConceptIDs = types.ListNull(types.StringType)
	}

	if !prepared.ownership.conceptIDs {
		state.ConceptIDs = types.ListNull(types.StringType)
	}

	return taxonomyConceptSchemeRequestFromModel(ctx, state)
}

func taxonomyConceptSchemeRequestFromModel(ctx context.Context, model TaxonomyConceptSchemeModel) (cm.TaxonomyConceptSchemeRequest, diag.Diagnostics) {
	request, diags := model.ToRequest(ctx)
	if diags.HasError() {
		return cm.TaxonomyConceptSchemeRequest{}, diags
	}

	if model.TopConceptIDs.IsNull() {
		request.TopConcepts = nil
	}

	if model.ConceptIDs.IsNull() {
		request.Concepts = nil
	}

	return request, nil
}

func newTaxonomyConceptSchemeRefreshState(ctx context.Context, prior TaxonomyConceptSchemeModel, response cm.TaxonomyConceptScheme) (TaxonomyConceptSchemeModel, diag.Diagnostics) {
	data, diags := NewTaxonomyConceptSchemeModelFromResponse(ctx, response)
	if diags.HasError() {
		return data, diags
	}

	data.Definition = taxonomyNullableLocalizedStringAfterRefresh(prior.Definition, data.Definition)
	data.Timeouts = prior.Timeouts

	return data, diags
}
