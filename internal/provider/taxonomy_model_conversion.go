package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var errTaxonomyLocaleNotPreserved = errors.New("contentful did not preserve configured taxonomy locale")

func nullableLocalizedString(ctx context.Context, value types.Map, valuePath path.Path, diags *diag.Diagnostics) cm.OptNilNullableLocalizedString {
	if value.IsNull() {
		var result cm.OptNilNullableLocalizedString
		result.SetToNull()

		return result
	}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A nullable localized string must be known before it can be sent to Contentful.")

		return cm.OptNilNullableLocalizedString{}
	}

	values, valueDiags := knownStringMap(ctx, value, valuePath)
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

func (model TaxonomyConceptModel) ToRequest(ctx context.Context) (cm.TaxonomyConceptRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	prefLabel, valueDiags := knownStringMap(ctx, model.PrefLabel, path.Root("pref_label"))
	diags.Append(valueDiags...)
	altLabels, valueDiags := optionalComputedStringMap(ctx, model.AltLabels)
	diags.Append(valueDiags...)
	hiddenLabels, valueDiags := optionalComputedStringMap(ctx, model.HiddenLabels)
	diags.Append(valueDiags...)

	for locale := range prefLabel {
		if _, ok := altLabels[locale]; !ok {
			altLabels[locale] = []string{}
		}

		if _, ok := hiddenLabels[locale]; !ok {
			hiddenLabels[locale] = []string{}
		}
	}

	notations, valueDiags := optionalComputedStringListValue(ctx, model.Notations)
	diags.Append(valueDiags...)
	broader, valueDiags := optionalComputedStringListValue(ctx, model.BroaderConceptIDs)
	diags.Append(valueDiags...)
	related, valueDiags := optionalComputedStringListValue(ctx, model.RelatedConceptIDs)
	diags.Append(valueDiags...)

	uri, uriDiags := optionalKnownStringPointer(model.URI, path.Root("uri"))
	diags.Append(uriDiags...)

	request := cm.TaxonomyConceptRequest{
		URI:           cm.NewOptNilPointerString(uri),
		PrefLabel:     cm.LocalizedString(prefLabel),
		AltLabels:     cm.NewOptLocalizedStringList(cm.LocalizedStringList(altLabels)),
		HiddenLabels:  cm.NewOptLocalizedStringList(cm.LocalizedStringList(hiddenLabels)),
		Notations:     notations,
		Note:          nullableLocalizedString(ctx, model.Note, path.Root("note"), &diags),
		ChangeNote:    nullableLocalizedString(ctx, model.ChangeNote, path.Root("change_note"), &diags),
		Definition:    nullableLocalizedString(ctx, model.Definition, path.Root("definition"), &diags),
		EditorialNote: nullableLocalizedString(ctx, model.EditorialNote, path.Root("editorial_note"), &diags),
		Example:       nullableLocalizedString(ctx, model.Example, path.Root("example"), &diags),
		HistoryNote:   nullableLocalizedString(ctx, model.HistoryNote, path.Root("history_note"), &diags),
		ScopeNote:     nullableLocalizedString(ctx, model.ScopeNote, path.Root("scope_note"), &diags),
		Broader:       conceptLinks(broader),
		Related:       conceptLinks(related),
	}

	if diags.HasError() {
		return cm.TaxonomyConceptRequest{}, diags
	}

	return request, diags
}

func NewTaxonomyConceptModelFromResponse(ctx context.Context, response cm.TaxonomyConcept) (TaxonomyConceptModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	organizationID := response.Sys.Organization.Sys.ID
	conceptID := response.Sys.ID
	prefLabel, valueDiags := types.MapValueFrom(ctx, types.StringType, map[string]string(response.PrefLabel))
	diags.Append(valueDiags...)

	altLabels, _ := response.AltLabels.Get()
	altLabelsValue, valueDiags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, map[string][]string(altLabels))
	diags.Append(valueDiags...)

	hiddenLabels, _ := response.HiddenLabels.Get()
	hiddenLabelsValue, valueDiags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, map[string][]string(hiddenLabels))
	diags.Append(valueDiags...)
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

func preserveConfiguredLabelMapShape(model *TaxonomyConceptModel, configured TaxonomyConceptModel) {
	model.AltLabels = labelMapWithConfiguredKeys(configured.AltLabels, model.AltLabels)
	model.HiddenLabels = labelMapWithConfiguredKeys(configured.HiddenLabels, model.HiddenLabels)
}

func labelMapWithConfiguredKeys(configured, returned types.Map) types.Map {
	if configured.IsNull() || configured.IsUnknown() || returned.IsNull() || returned.IsUnknown() {
		return returned
	}

	returnedElements := returned.Elements()
	values := make(map[string]attr.Value, len(configured.Elements()))

	for locale := range configured.Elements() {
		if value, ok := returnedElements[locale]; ok {
			values[locale] = value
		}
	}

	return types.MapValueMust(types.ListType{ElemType: types.StringType}, values)
}

func (model TaxonomyConceptSchemeModel) ToRequest(ctx context.Context) (cm.TaxonomyConceptSchemeRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	prefLabel, valueDiags := knownStringMap(ctx, model.PrefLabel, path.Root("pref_label"))
	diags.Append(valueDiags...)
	topIDs, valueDiags := optionalComputedStringListValue(ctx, model.TopConceptIDs)
	diags.Append(valueDiags...)
	ids, valueDiags := optionalComputedStringListValue(ctx, model.ConceptIDs)
	diags.Append(valueDiags...)

	uri, uriDiags := optionalKnownStringPointer(model.URI, path.Root("uri"))
	diags.Append(uriDiags...)

	request := cm.TaxonomyConceptSchemeRequest{
		URI: cm.NewOptNilPointerString(uri), PrefLabel: cm.LocalizedString(prefLabel),
		Definition: nullableLocalizedString(ctx, model.Definition, path.Root("definition"), &diags), TopConcepts: conceptLinks(topIDs), Concepts: conceptLinks(ids),
	}

	if diags.HasError() {
		return cm.TaxonomyConceptSchemeRequest{}, diags
	}

	return request, diags
}

func optionalKnownStringPointer(value types.String, valuePath path.Path) (*string, diag.Diagnostics) {
	if value.IsNull() {
		return nil, nil
	}

	if value.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected unknown value", "The optional string must be known before it can be sent to Contentful.")}
	}

	result := value.ValueString()

	return &result, nil
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

func taxonomyPatch(current, desired any) (cm.TaxonomyPatch, error) {
	currentEncoded, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("encode current taxonomy request: %w", err)
	}

	currentFields := map[string]json.RawMessage{}

	err = json.Unmarshal(currentEncoded, &currentFields)
	if err != nil {
		return nil, fmt.Errorf("decode current taxonomy request: %w", err)
	}

	desiredEncoded, err := json.Marshal(desired)
	if err != nil {
		return nil, fmt.Errorf("encode desired taxonomy request: %w", err)
	}

	desiredFields := map[string]json.RawMessage{}

	err = json.Unmarshal(desiredEncoded, &desiredFields)
	if err != nil {
		return nil, fmt.Errorf("decode desired taxonomy request: %w", err)
	}

	keys := make([]string, 0, len(desiredFields))
	for key := range desiredFields {
		if bytes.Equal(currentFields[key], desiredFields[key]) {
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	result := make(cm.TaxonomyPatch, 0, len(keys))
	for _, key := range keys {
		result = append(result, cm.TaxonomyPatchItem{Op: cm.TaxonomyPatchItemOpAdd, Path: "/" + key, Value: jx.Raw(desiredFields[key])})
	}

	return result, nil
}

func taxonomyConceptRequestFromResponse(concept cm.TaxonomyConcept) cm.TaxonomyConceptRequest {
	return cm.TaxonomyConceptRequest{
		URI: concept.URI, PrefLabel: concept.PrefLabel, AltLabels: concept.AltLabels, HiddenLabels: concept.HiddenLabels,
		Notations: concept.Notations, Note: concept.Note, ChangeNote: concept.ChangeNote, Definition: concept.Definition,
		EditorialNote: concept.EditorialNote, Example: concept.Example, HistoryNote: concept.HistoryNote,
		ScopeNote: concept.ScopeNote, Broader: concept.Broader, Related: concept.Related,
	}
}

func taxonomyConceptSchemeRequestFromResponse(scheme cm.TaxonomyConceptScheme) cm.TaxonomyConceptSchemeRequest {
	return cm.TaxonomyConceptSchemeRequest{
		URI: scheme.URI, PrefLabel: scheme.PrefLabel, Definition: scheme.Definition,
		TopConcepts: scheme.TopConcepts, Concepts: scheme.Concepts,
	}
}

func validateLocalizedStrings(field string, configured, returned map[string]string) error {
	for locale, configuredValue := range configured {
		if returnedValue, exists := returned[locale]; !exists || returnedValue != configuredValue {
			return fmt.Errorf("%w: field %s, locale %q", errTaxonomyLocaleNotPreserved, field, locale)
		}
	}

	return nil
}

func validateLocalizedStringLists(field string, configured, returned map[string][]string) error {
	for locale, configuredValue := range configured {
		if returnedValue, exists := returned[locale]; !exists || !reflect.DeepEqual(returnedValue, configuredValue) {
			return fmt.Errorf("%w: field %s, locale %q", errTaxonomyLocaleNotPreserved, field, locale)
		}
	}

	return nil
}

func validateTaxonomyConceptResponse(request cm.TaxonomyConceptRequest, response cm.TaxonomyConcept) error {
	err := validateLocalizedStrings("pref_label", map[string]string(request.PrefLabel), map[string]string(response.PrefLabel))
	if err != nil {
		return err
	}

	configuredAlt, _ := request.AltLabels.Get()
	returnedAlt, _ := response.AltLabels.Get()

	err = validateLocalizedStringLists("alt_labels", map[string][]string(configuredAlt), map[string][]string(returnedAlt))
	if err != nil {
		return err
	}

	configuredHidden, _ := request.HiddenLabels.Get()
	returnedHidden, _ := response.HiddenLabels.Get()

	return validateLocalizedStringLists("hidden_labels", map[string][]string(configuredHidden), map[string][]string(returnedHidden))
}

func validateTaxonomyConceptSchemeResponse(request cm.TaxonomyConceptSchemeRequest, response cm.TaxonomyConceptScheme) error {
	return validateLocalizedStrings("pref_label", map[string]string(request.PrefLabel), map[string]string(response.PrefLabel))
}
