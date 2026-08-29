package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var errTaxonomyRequestFieldMissing = errors.New("taxonomy request field is missing")

type taxonomyConceptMutationOwnership struct {
	altLabels         bool
	hiddenLabels      bool
	notations         bool
	broaderConceptIDs bool
	relatedConceptIDs bool
}

type preparedTaxonomyConceptMutation struct {
	plan      TaxonomyConceptModel
	ownership taxonomyConceptMutationOwnership
}

// prepareTaxonomyConceptMutation captures configuration ownership before any
// request is built. A null Optional+Computed collection is response-owned;
// every non-null value is configuration-owned, and an unknown configuration
// value is rejected at its attribute path before network I/O.
func prepareTaxonomyConceptMutation(config, plan TaxonomyConceptModel) (preparedTaxonomyConceptMutation, diag.Diagnostics) {
	prepared := preparedTaxonomyConceptMutation{plan: plan}
	diags := diag.Diagnostics{}
	prepared.ownership.altLabels = taxonomyCollectionOwnership(config.AltLabels, plan.AltLabels, path.Root("alt_labels"), &diags)
	prepared.ownership.hiddenLabels = taxonomyCollectionOwnership(config.HiddenLabels, plan.HiddenLabels, path.Root("hidden_labels"), &diags)
	prepared.ownership.notations = taxonomyCollectionOwnership(config.Notations, plan.Notations, path.Root("notations"), &diags)
	prepared.ownership.broaderConceptIDs = taxonomyCollectionOwnership(config.BroaderConceptIDs, plan.BroaderConceptIDs, path.Root("broader_concept_ids"), &diags)
	prepared.ownership.relatedConceptIDs = taxonomyCollectionOwnership(config.RelatedConceptIDs, plan.RelatedConceptIDs, path.Root("related_concept_ids"), &diags)

	return prepared, diags
}

func taxonomyCollectionOwnership[T attr.Value](config, plan T, valuePath path.Path, diags *diag.Diagnostics) bool {
	if config.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unknown configuration value", "Taxonomy collection ownership must be known before Contentful can be changed.")

		return false
	}

	if config.IsNull() {
		return false
	}

	diags.Append(rejectUnknownConfigurationOwnedRequestValue(plan, config, valuePath)...)
	taxonomyRejectNullConfigurationOwnedCollection(plan, valuePath, diags)

	return true
}

func taxonomyRejectNullConfigurationOwnedCollection(value interface {
	IsNull() bool
	IsUnknown() bool
}, valuePath path.Path, diags *diag.Diagnostics,
) {
	if !value.IsNull() {
		return
	}

	diags.AddAttributeError(valuePath, "Unavailable planned value", "A configuration-owned taxonomy collection must be non-null before Contentful can be changed.")
}

func taxonomyKnownPlanValue(value interface {
	IsNull() bool
	IsUnknown() bool
},
) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func (prepared preparedTaxonomyConceptMutation) PatchFromState(ctx context.Context, state TaxonomyConceptModel) (cm.TaxonomyPatch, diag.Diagnostics) {
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
		diags.Append(taxonomyPatchErrorDiagnostics("taxonomy concept", err)...)

		return nil, diags
	}
	// Wire conversion adds preferred locales to owned label maps. Reconcile the
	// Terraform representation with the canonical wire maps: an equivalent
	// logical value may suppress a patch only when the wire maps also match.
	for _, field := range []struct {
		owned     bool
		path      string
		valuePath path.Path
		state     types.Map
		plan      types.Map
	}{
		{prepared.ownership.altLabels, "/altLabels", path.Root("alt_labels"), state.AltLabels, prepared.plan.AltLabels},
		{prepared.ownership.hiddenLabels, "/hiddenLabels", path.Root("hidden_labels"), state.HiddenLabels, prepared.plan.HiddenLabels},
	} {
		if !field.owned {
			continue
		}

		var reconcileErr error

		patch, reconcileErr = reconcileOwnedTaxonomyLabelPatch(patch, &current, &desired, field.path, field.valuePath, field.state, field.plan, prepared.plan.PrefLabel)
		if reconcileErr != nil {
			diags.Append(taxonomyPatchErrorDiagnostics("taxonomy concept", reconcileErr)...)

			return nil, diags
		}
	}

	sort.Slice(patch, func(i, j int) bool { return patch[i].Path < patch[j].Path })

	return patch, diags
}

func reconcileOwnedTaxonomyLabelPatch(patch cm.TaxonomyPatch, current, desired *cm.TaxonomyConceptRequest, fieldPath string, valuePath path.Path, state, plan, prefLabel types.Map) (cm.TaxonomyPatch, error) {
	equal, _ := taxonomyLabelMapsEquivalentAt(plan, state, prefLabel, valuePath)
	if !equal {
		if taxonomyPatchHasPath(patch, fieldPath) {
			return patch, nil
		}

		value, err := taxonomyRequestField(desired, fieldPath[1:])
		if err != nil {
			return nil, err
		}

		return append(patch, cm.TaxonomyPatchItem{Op: cm.TaxonomyPatchItemOpAdd, Path: fieldPath, Value: value}), nil
	}

	if !taxonomyPatchHasPath(patch, fieldPath) {
		return patch, nil
	}

	currentValue, err := taxonomyRequestField(current, fieldPath[1:])
	if err != nil {
		return nil, err
	}

	desiredValue, err := taxonomyRequestField(desired, fieldPath[1:])
	if err != nil {
		return nil, err
	}

	wireEqual, err := taxonomyJSONEqual(currentValue, desiredValue)
	if err != nil {
		return nil, err
	}

	if wireEqual {
		return taxonomyPatchWithoutPath(patch, fieldPath), nil
	}

	return patch, nil
}

func taxonomyRequestField(request *cm.TaxonomyConceptRequest, key string) (jx.Raw, error) {
	fields, err := taxonomyRequestFields(request, "provided")
	if err != nil {
		return nil, err
	}

	value, ok := fields[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", errTaxonomyRequestFieldMissing, key)
	}

	return jx.Raw(value), nil
}

func taxonomyPatchHasPath(patch cm.TaxonomyPatch, target string) bool {
	for _, item := range patch {
		if item.Path == target {
			return true
		}
	}

	return false
}

func taxonomyPatchWithoutPath(patch cm.TaxonomyPatch, target string) cm.TaxonomyPatch {
	result := patch[:0]
	for _, item := range patch {
		if item.Path != target {
			result = append(result, item)
		}
	}

	return result
}

func (prepared preparedTaxonomyConceptMutation) NoopState(state TaxonomyConceptModel) TaxonomyConceptModel {
	data := state
	data.IDIdentityModel = prepared.plan.IDIdentityModel
	data.TaxonomyConceptIdentityModel = prepared.plan.TaxonomyConceptIdentityModel
	data.URI = prepared.plan.URI
	data.PrefLabel = prepared.plan.PrefLabel
	data.Note = prepared.plan.Note
	data.ChangeNote = prepared.plan.ChangeNote
	data.Definition = prepared.plan.Definition
	data.EditorialNote = prepared.plan.EditorialNote
	data.Example = prepared.plan.Example
	data.HistoryNote = prepared.plan.HistoryNote

	data.ScopeNote = prepared.plan.ScopeNote
	if taxonomyKnownPlanValue(prepared.plan.AltLabels) {
		data.AltLabels = prepared.plan.AltLabels
	}

	if taxonomyKnownPlanValue(prepared.plan.HiddenLabels) {
		data.HiddenLabels = prepared.plan.HiddenLabels
	}

	if taxonomyKnownPlanValue(prepared.plan.Notations) {
		data.Notations = prepared.plan.Notations
	}

	if taxonomyKnownPlanValue(prepared.plan.BroaderConceptIDs) {
		data.BroaderConceptIDs = prepared.plan.BroaderConceptIDs
	}

	if taxonomyKnownPlanValue(prepared.plan.RelatedConceptIDs) {
		data.RelatedConceptIDs = prepared.plan.RelatedConceptIDs
	}

	data.Timeouts = prepared.plan.Timeouts

	return data
}

func (prepared preparedTaxonomyConceptMutation) ProjectResponse(ctx context.Context, response cm.TaxonomyConcept) (TaxonomyConceptModel, diag.Diagnostics, diag.Diagnostics) {
	data, responseDiags := NewTaxonomyConceptModelFromResponse(ctx, response)

	data.Timeouts = prepared.plan.Timeouts
	if responseDiags.HasError() {
		return data, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}
	mismatch := taxonomyMutationIdentityConsistency("taxonomy concept", path.Root("organization_id"), prepared.plan.OrganizationID, data.OrganizationID, &consistencyDiags)
	mismatch = taxonomyMutationIdentityConsistency("taxonomy concept", path.Root("concept_id"), prepared.plan.ConceptID, data.ConceptID, &consistencyDiags) || mismatch
	data.OrganizationID, data.ConceptID = prepared.plan.OrganizationID, prepared.plan.ConceptID
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(prepared.plan.OrganizationID.ValueString(), prepared.plan.ConceptID.ValueString())

	var projects []func()

	compare := func(name string, valuePath path.Path, planned, remote attr.Value, project func()) {
		if compareTaxonomyMutationValue("taxonomy concept", name, valuePath, planned, remote, &consistencyDiags) {
			projects = append(projects, project)
		} else {
			mismatch = true
		}
	}
	compareNullableLocalizedString := func(name string, valuePath path.Path, planned, remote types.Map, project func()) {
		if compareTaxonomyNullableLocalizedStringMutationValue("taxonomy concept", name, valuePath, planned, remote, &consistencyDiags) {
			projects = append(projects, project)
		} else {
			mismatch = true
		}
	}

	compare("uri", path.Root("uri"), prepared.plan.URI, data.URI, func() { data.URI = prepared.plan.URI })
	compare("pref_label", path.Root("pref_label"), prepared.plan.PrefLabel, data.PrefLabel, func() { data.PrefLabel = prepared.plan.PrefLabel })

	if taxonomyKnownPlanValue(prepared.plan.AltLabels) {
		if compareTaxonomyLabelMutationValue("taxonomy concept", "alt_labels", path.Root("alt_labels"), prepared.plan.AltLabels, data.AltLabels, prepared.plan.PrefLabel, &consistencyDiags) {
			projects = append(projects, func() { data.AltLabels = prepared.plan.AltLabels })
		} else {
			mismatch = true
		}
	}

	if taxonomyKnownPlanValue(prepared.plan.HiddenLabels) {
		if compareTaxonomyLabelMutationValue("taxonomy concept", "hidden_labels", path.Root("hidden_labels"), prepared.plan.HiddenLabels, data.HiddenLabels, prepared.plan.PrefLabel, &consistencyDiags) {
			projects = append(projects, func() { data.HiddenLabels = prepared.plan.HiddenLabels })
		} else {
			mismatch = true
		}
	}

	if taxonomyKnownPlanValue(prepared.plan.Notations) {
		compare("notations", path.Root("notations"), prepared.plan.Notations, data.Notations, func() { data.Notations = prepared.plan.Notations })
	}

	compareNullableLocalizedString("note", path.Root("note"), prepared.plan.Note, data.Note, func() { data.Note = prepared.plan.Note })
	compareNullableLocalizedString("change_note", path.Root("change_note"), prepared.plan.ChangeNote, data.ChangeNote, func() { data.ChangeNote = prepared.plan.ChangeNote })
	compareNullableLocalizedString("definition", path.Root("definition"), prepared.plan.Definition, data.Definition, func() { data.Definition = prepared.plan.Definition })
	compareNullableLocalizedString("editorial_note", path.Root("editorial_note"), prepared.plan.EditorialNote, data.EditorialNote, func() { data.EditorialNote = prepared.plan.EditorialNote })
	compareNullableLocalizedString("example", path.Root("example"), prepared.plan.Example, data.Example, func() { data.Example = prepared.plan.Example })
	compareNullableLocalizedString("history_note", path.Root("history_note"), prepared.plan.HistoryNote, data.HistoryNote, func() { data.HistoryNote = prepared.plan.HistoryNote })
	compareNullableLocalizedString("scope_note", path.Root("scope_note"), prepared.plan.ScopeNote, data.ScopeNote, func() { data.ScopeNote = prepared.plan.ScopeNote })

	if taxonomyKnownPlanValue(prepared.plan.BroaderConceptIDs) {
		compare("broader_concept_ids", path.Root("broader_concept_ids"), prepared.plan.BroaderConceptIDs, data.BroaderConceptIDs, func() { data.BroaderConceptIDs = prepared.plan.BroaderConceptIDs })
	}

	if taxonomyKnownPlanValue(prepared.plan.RelatedConceptIDs) {
		compare("related_concept_ids", path.Root("related_concept_ids"), prepared.plan.RelatedConceptIDs, data.RelatedConceptIDs, func() { data.RelatedConceptIDs = prepared.plan.RelatedConceptIDs })
	}

	if !mismatch {
		for _, project := range projects {
			project()
		}
	}

	return data, responseDiags, consistencyDiags
}

func (prepared preparedTaxonomyConceptMutation) planRequest(ctx context.Context) (cm.TaxonomyConceptRequest, diag.Diagnostics) {
	model := prepared.plan
	if !prepared.ownership.altLabels && !taxonomyKnownPlanValue(model.AltLabels) {
		model.AltLabels = types.MapNull(types.ListType{ElemType: types.StringType})
	}

	if !prepared.ownership.hiddenLabels && !taxonomyKnownPlanValue(model.HiddenLabels) {
		model.HiddenLabels = types.MapNull(types.ListType{ElemType: types.StringType})
	}

	if !prepared.ownership.notations && !taxonomyKnownPlanValue(model.Notations) {
		model.Notations = types.ListNull(types.StringType)
	}

	if !prepared.ownership.broaderConceptIDs && !taxonomyKnownPlanValue(model.BroaderConceptIDs) {
		model.BroaderConceptIDs = types.ListNull(types.StringType)
	}

	if !prepared.ownership.relatedConceptIDs && !taxonomyKnownPlanValue(model.RelatedConceptIDs) {
		model.RelatedConceptIDs = types.ListNull(types.StringType)
	}

	return taxonomyConceptRequestFromModel(ctx, model)
}

func (prepared preparedTaxonomyConceptMutation) stateRequest(ctx context.Context, state TaxonomyConceptModel) (cm.TaxonomyConceptRequest, diag.Diagnostics) {
	if !prepared.ownership.altLabels {
		state.AltLabels = types.MapNull(types.ListType{ElemType: types.StringType})
	}

	if !prepared.ownership.hiddenLabels {
		state.HiddenLabels = types.MapNull(types.ListType{ElemType: types.StringType})
	}

	if !prepared.ownership.notations {
		state.Notations = types.ListNull(types.StringType)
	}

	if !prepared.ownership.broaderConceptIDs {
		state.BroaderConceptIDs = types.ListNull(types.StringType)
	}

	if !prepared.ownership.relatedConceptIDs {
		state.RelatedConceptIDs = types.ListNull(types.StringType)
	}

	return taxonomyConceptRequestFromModel(ctx, state)
}

func taxonomyConceptRequestFromModel(ctx context.Context, model TaxonomyConceptModel) (cm.TaxonomyConceptRequest, diag.Diagnostics) {
	request, diags := model.ToRequest(ctx)
	if diags.HasError() {
		return cm.TaxonomyConceptRequest{}, diags
	}

	if model.AltLabels.IsNull() {
		request.AltLabels.Reset()
	}

	if model.HiddenLabels.IsNull() {
		request.HiddenLabels.Reset()
	}

	if model.Notations.IsNull() {
		request.Notations = nil
	}

	if model.BroaderConceptIDs.IsNull() {
		request.Broader = nil
	}

	if model.RelatedConceptIDs.IsNull() {
		request.Related = nil
	}

	return request, nil
}

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

func compareTaxonomyMutationValue(resourceName, attributeName string, valuePath path.Path, planned, remote attr.Value, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A configuration-owned taxonomy value must be known before Contentful can be changed.")

		return false
	}

	if planned.Equal(remote) {
		return true
	}

	diags.AddAttributeError(taxonomyMutationDifferencePath(valuePath, planned, remote), "Unexpected Contentful "+resourceName+" response", fmt.Sprintf("The %s response differed meaningfully from the Terraform plan.", attributeName))

	return false
}

func compareTaxonomyNullableLocalizedStringMutationValue(resourceName, attributeName string, valuePath path.Path, planned, remote types.Map, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A configuration-owned taxonomy value must be known before Contentful can be changed.")

		return false
	}

	if taxonomyNullableLocalizedStringEquivalentAfterMutation(planned, remote) {
		return true
	}

	diags.AddAttributeError(taxonomyMutationDifferencePath(valuePath, planned, remote), "Unexpected Contentful "+resourceName+" response", fmt.Sprintf("The %s response differed meaningfully from the Terraform plan.", attributeName))

	return false
}

func taxonomyMutationIdentityConsistency(resourceName string, valuePath path.Path, planned, remote types.String, diags *diag.Diagnostics) bool {
	if !planned.IsNull() && !planned.IsUnknown() && !planned.Equal(remote) {
		diags.AddAttributeError(valuePath, "Unexpected Contentful "+resourceName+" response", "The response identity differed from the requested taxonomy endpoint.")

		return true
	}

	return false
}

func compareTaxonomyLabelMutationValue(resourceName, attributeName string, valuePath path.Path, planned, remote, prefLabel types.Map, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A configuration-owned taxonomy value must be known before Contentful can be changed.")

		return false
	}

	equal, differencePath := taxonomyLabelMapsEquivalentAt(planned, remote, prefLabel, valuePath)
	if equal {
		return true
	}

	diags.AddAttributeError(differencePath, "Unexpected Contentful "+resourceName+" response", fmt.Sprintf("The %s response differed meaningfully from the Terraform plan.", attributeName))

	return false
}

func taxonomyMutationDifferencePath(valuePath path.Path, planned, remote attr.Value) path.Path {
	plannedList, plannedIsList := planned.(types.List)

	remoteList, remoteIsList := remote.(types.List)
	if plannedIsList && remoteIsList && !plannedList.IsNull() && !plannedList.IsUnknown() && !remoteList.IsNull() && !remoteList.IsUnknown() {
		return taxonomyListDifferencePath(valuePath, plannedList, remoteList)
	}

	plannedMap, plannedIsMap := planned.(types.Map)

	remoteMap, remoteIsMap := remote.(types.Map)
	if plannedIsMap && remoteIsMap && !plannedMap.IsNull() && !plannedMap.IsUnknown() && !remoteMap.IsNull() && !remoteMap.IsUnknown() {
		keys := make([]string, 0, len(plannedMap.Elements())+len(remoteMap.Elements()))
		for key := range plannedMap.Elements() {
			keys = append(keys, key)
		}

		for key := range remoteMap.Elements() {
			if _, ok := plannedMap.Elements()[key]; !ok {
				keys = append(keys, key)
			}
		}

		sort.Strings(keys)

		for _, key := range keys {
			plannedValue, plannedExists := plannedMap.Elements()[key]

			remoteValue, remoteExists := remoteMap.Elements()[key]
			if !plannedExists || !remoteExists || !plannedValue.Equal(remoteValue) {
				return valuePath.AtMapKey(key)
			}
		}
	}

	return valuePath
}

func taxonomyListDifferencePath(valuePath path.Path, planned, remote types.List) path.Path {
	plannedElements, remoteElements := planned.Elements(), remote.Elements()

	limit := min(len(plannedElements), len(remoteElements))
	for index := range limit {
		if !plannedElements[index].Equal(remoteElements[index]) {
			return valuePath.AtListIndex(index)
		}
	}

	return valuePath.AtListIndex(limit)
}

func newTaxonomyConceptRefreshState(ctx context.Context, prior TaxonomyConceptModel, response cm.TaxonomyConcept) (TaxonomyConceptModel, diag.Diagnostics) {
	data, diags := NewTaxonomyConceptModelFromResponse(ctx, response)
	if diags.HasError() {
		return data, diags
	}

	data.AltLabels = taxonomyLabelMapAfterRefresh(prior.AltLabels, data.AltLabels, data.PrefLabel)
	data.HiddenLabels = taxonomyLabelMapAfterRefresh(prior.HiddenLabels, data.HiddenLabels, data.PrefLabel)
	data.Note = taxonomyNullableLocalizedStringAfterRefresh(prior.Note, data.Note)
	data.ChangeNote = taxonomyNullableLocalizedStringAfterRefresh(prior.ChangeNote, data.ChangeNote)
	data.Definition = taxonomyNullableLocalizedStringAfterRefresh(prior.Definition, data.Definition)
	data.EditorialNote = taxonomyNullableLocalizedStringAfterRefresh(prior.EditorialNote, data.EditorialNote)
	data.Example = taxonomyNullableLocalizedStringAfterRefresh(prior.Example, data.Example)
	data.HistoryNote = taxonomyNullableLocalizedStringAfterRefresh(prior.HistoryNote, data.HistoryNote)
	data.ScopeNote = taxonomyNullableLocalizedStringAfterRefresh(prior.ScopeNote, data.ScopeNote)
	data.Timeouts = prior.Timeouts

	return data, diags
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
