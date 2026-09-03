package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func nullableLocalizedString(value types.Map, valuePath path.Path, diags *diag.Diagnostics) cm.OptNilNullableLocalizedString {
	if value.IsNull() {
		var result cm.OptNilNullableLocalizedString
		result.SetToNull()

		return result
	}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A nullable localized string must be known before it can be sent to Contentful.")

		return cm.OptNilNullableLocalizedString{}
	}

	values, valueDiags := RequireKnownStringMap(value, valuePath)
	*diags = append(*diags, valueDiags...)

	return cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString(values))
}

func localizedStringValue(ctx context.Context, value cm.OptNilNullableLocalizedString, diags *diag.Diagnostics) types.Map {
	if values, ok := value.Get(); ok {
		result, valueDiags := types.MapValueFrom(ctx, types.StringType, map[string]string(values))
		*diags = append(*diags, valueDiags...)

		return result
	}

	return types.MapNull(types.StringType)
}

func conceptLinks(ids []string) []cm.TaxonomyConceptLink {
	result := make([]cm.TaxonomyConceptLink, 0, len(ids))
	for _, id := range ids {
		result = append(result, cm.NewTaxonomyConceptLink(id))
	}

	return result
}

func conceptLinkIDs(links []cm.TaxonomyConceptLink) []string {
	result := make([]string, 0, len(links))
	for _, link := range links {
		result = append(result, link.Sys.ID)
	}

	return result
}

// optionalComputedStringListValue converts null to an empty intermediate slice
// and rejects unknown values. Prepared mutations replace response-owned
// collections with their omitted wire representation.
func optionalComputedStringListValue(value types.List, valuePath path.Path) ([]string, diag.Diagnostics) {
	result := []string{}
	if value.IsNull() {
		return result, nil
	}

	if value.IsUnknown() {
		return result, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected unknown value", "The taxonomy collection must be known before it can be sent to Contentful.")}
	}

	return RequireKnownStringList(value, valuePath)
}

func (model TaxonomyConceptModel) ToRequest(_ context.Context) (cm.TaxonomyConceptRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	prefLabel, valueDiags := RequireKnownStringMap(model.PrefLabel, path.Root("pref_label"))
	diags.Append(valueDiags...)
	altLabels := taxonomyLabelMapRequest(prefLabel, model.AltLabels, path.Root("alt_labels"), &diags)
	hiddenLabels := taxonomyLabelMapRequest(prefLabel, model.HiddenLabels, path.Root("hidden_labels"), &diags)

	notations, valueDiags := optionalComputedStringListValue(model.Notations, path.Root("notations"))
	diags.Append(valueDiags...)
	broader, valueDiags := optionalComputedStringListValue(model.BroaderConceptIDs, path.Root("broader_concept_ids"))
	diags.Append(valueDiags...)
	related, valueDiags := optionalComputedStringListValue(model.RelatedConceptIDs, path.Root("related_concept_ids"))
	diags.Append(valueDiags...)

	uri, uriDiags := requestNullableString(model.URI, path.Root("uri"))
	diags.Append(uriDiags...)

	request := cm.TaxonomyConceptRequest{
		URI:           uri,
		PrefLabel:     cm.LocalizedString(prefLabel),
		AltLabels:     altLabels,
		HiddenLabels:  hiddenLabels,
		Notations:     notations,
		Note:          nullableLocalizedString(model.Note, path.Root("note"), &diags),
		ChangeNote:    nullableLocalizedString(model.ChangeNote, path.Root("change_note"), &diags),
		Definition:    nullableLocalizedString(model.Definition, path.Root("definition"), &diags),
		EditorialNote: nullableLocalizedString(model.EditorialNote, path.Root("editorial_note"), &diags),
		Example:       nullableLocalizedString(model.Example, path.Root("example"), &diags),
		HistoryNote:   nullableLocalizedString(model.HistoryNote, path.Root("history_note"), &diags),
		ScopeNote:     nullableLocalizedString(model.ScopeNote, path.Root("scope_note"), &diags),
		Broader:       conceptLinks(broader),
		Related:       conceptLinks(related),
	}

	if diags.HasError() {
		return cm.TaxonomyConceptRequest{}, diags
	}

	return request, diags
}

func taxonomyLabelMapRequest(prefLabel map[string]string, value types.Map, valuePath path.Path, diags *diag.Diagnostics) cm.OptLocalizedStringList {
	if value.IsNull() {
		return cm.OptLocalizedStringList{}
	}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "The taxonomy label map must be known before it can be sent to Contentful.")

		return cm.OptLocalizedStringList{}
	}

	labels, valueDiags := RequireKnownStringListMap(value, valuePath)
	*diags = append(*diags, valueDiags...)

	if valueDiags.HasError() {
		return cm.OptLocalizedStringList{}
	}

	for locale := range prefLabel {
		if _, ok := labels[locale]; !ok {
			labels[locale] = []string{}
		}
	}

	return cm.NewOptLocalizedStringList(cm.LocalizedStringList(labels))
}

func NewTaxonomyConceptModelFromResponse(ctx context.Context, response cm.TaxonomyConcept) (TaxonomyConceptModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	organizationID := response.Sys.Organization.Sys.ID
	conceptID := response.Sys.ID

	prefLabel, valueDiags := types.MapValueFrom(ctx, types.StringType, map[string]string(response.PrefLabel))
	diags.Append(valueDiags...)

	labelListType := types.ListType{ElemType: types.StringType}
	altLabels, altLabelsSet := response.AltLabels.Get()

	altLabelsValue := types.MapNull(labelListType)
	if altLabelsSet {
		altLabelsValue, valueDiags = types.MapValueFrom(ctx, labelListType, map[string][]string(altLabels))
		diags.Append(valueDiags...)
	}

	hiddenLabels, hiddenLabelsSet := response.HiddenLabels.Get()

	hiddenLabelsValue := types.MapNull(labelListType)
	if hiddenLabelsSet {
		hiddenLabelsValue, valueDiags = types.MapValueFrom(ctx, labelListType, map[string][]string(hiddenLabels))
		diags.Append(valueDiags...)
	}

	notations, valueDiags := types.ListValueFrom(ctx, types.StringType, response.Notations)
	diags.Append(valueDiags...)
	broader, valueDiags := types.ListValueFrom(ctx, types.StringType, conceptLinkIDs(response.Broader))
	diags.Append(valueDiags...)
	related, valueDiags := types.ListValueFrom(ctx, types.StringType, conceptLinkIDs(response.Related))
	diags.Append(valueDiags...)

	schemeIDs := make([]string, 0, len(response.ConceptSchemes))
	for _, link := range response.ConceptSchemes {
		schemeIDs = append(schemeIDs, link.Sys.ID)
	}

	conceptSchemes, valueDiags := types.SetValueFrom(ctx, types.StringType, schemeIDs)
	diags.Append(valueDiags...)

	return TaxonomyConceptModel{
		IDIdentityModel:              NewIDIdentityModelFromMultipartID(organizationID, conceptID),
		TaxonomyConceptIdentityModel: TaxonomyConceptIdentityModel{OrganizationID: types.StringValue(organizationID), ConceptID: types.StringValue(conceptID)},
		URI:                          types.StringPointerValue(response.URI.ValueStringPointer()), PrefLabel: prefLabel, AltLabels: altLabelsValue, HiddenLabels: hiddenLabelsValue,
		Notations: notations, Note: localizedStringValue(ctx, response.Note, &diags), ChangeNote: localizedStringValue(ctx, response.ChangeNote, &diags),
		Definition: localizedStringValue(ctx, response.Definition, &diags), EditorialNote: localizedStringValue(ctx, response.EditorialNote, &diags),
		Example: localizedStringValue(ctx, response.Example, &diags), HistoryNote: localizedStringValue(ctx, response.HistoryNote, &diags), ScopeNote: localizedStringValue(ctx, response.ScopeNote, &diags),
		BroaderConceptIDs: broader, RelatedConceptIDs: related, ConceptSchemeIDs: conceptSchemes,
	}, diags
}

func (model TaxonomyConceptSchemeModel) ToRequest(_ context.Context) (cm.TaxonomyConceptSchemeRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	prefLabel, valueDiags := RequireKnownStringMap(model.PrefLabel, path.Root("pref_label"))
	diags.Append(valueDiags...)
	topIDs, valueDiags := optionalComputedStringListValue(model.TopConceptIDs, path.Root("top_concept_ids"))
	diags.Append(valueDiags...)
	ids, valueDiags := optionalComputedStringListValue(model.ConceptIDs, path.Root("concept_ids"))
	diags.Append(valueDiags...)

	uri, uriDiags := requestNullableString(model.URI, path.Root("uri"))
	diags.Append(uriDiags...)

	request := cm.TaxonomyConceptSchemeRequest{
		URI: uri, PrefLabel: cm.LocalizedString(prefLabel),
		Definition: nullableLocalizedString(model.Definition, path.Root("definition"), &diags), TopConcepts: conceptLinks(topIDs), Concepts: conceptLinks(ids),
	}

	if diags.HasError() {
		return cm.TaxonomyConceptSchemeRequest{}, diags
	}

	return request, diags
}

func NewTaxonomyConceptSchemeModelFromResponse(ctx context.Context, response cm.TaxonomyConceptScheme) (TaxonomyConceptSchemeModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	organizationID, schemeID := response.Sys.Organization.Sys.ID, response.Sys.ID

	prefLabel, valueDiags := types.MapValueFrom(ctx, types.StringType, map[string]string(response.PrefLabel))
	diags.Append(valueDiags...)
	topIDs, valueDiags := types.ListValueFrom(ctx, types.StringType, conceptLinkIDs(response.TopConcepts))
	diags.Append(valueDiags...)
	ids, valueDiags := types.ListValueFrom(ctx, types.StringType, conceptLinkIDs(response.Concepts))
	diags.Append(valueDiags...)

	return TaxonomyConceptSchemeModel{
		IDIdentityModel:                    NewIDIdentityModelFromMultipartID(organizationID, schemeID),
		TaxonomyConceptSchemeIdentityModel: TaxonomyConceptSchemeIdentityModel{OrganizationID: types.StringValue(organizationID), ConceptSchemeID: types.StringValue(schemeID)},
		URI:                                types.StringPointerValue(response.URI.ValueStringPointer()), PrefLabel: prefLabel, Definition: localizedStringValue(ctx, response.Definition, &diags),
		TopConceptIDs: topIDs, ConceptIDs: ids, TotalConcepts: types.Int64Value(int64(response.TotalConcepts)),
	}, diags
}
