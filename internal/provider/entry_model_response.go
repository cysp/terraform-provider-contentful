package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEntryResourceModelFromResponse(ctx context.Context, entry cm.Entry) (EntryModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := entry.Sys.Space.Sys.ID
	environmentID := entry.Sys.Environment.Sys.ID
	contentTypeID := entry.Sys.ContentType.Sys.ID
	entryID := entry.Sys.ID

	model := EntryModel{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
		EntryIdentityModel: NewEntryIdentityModel(spaceID, environmentID, entryID),
		ContentTypeID:      types.StringValue(contentTypeID),
		PublishedVersion:   types.Int64Null(),
	}
	if publishedVersion, ok := entry.Sys.PublishedVersion.Get(); ok {
		model.PublishedVersion = types.Int64Value(int64(publishedVersion))
	}

	fields, fieldsDiags := NewEntryFieldsFromResponse(ctx, path.Root("fields"), entry.Fields)
	diags.Append(fieldsDiags...)

	model.Fields = fields

	metadata, metadataDiags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), entry.Metadata)
	diags.Append(metadataDiags...)

	model.Metadata = metadata
	model.Timeouts = TimeoutsNull()

	return model, diags
}

func NewEntryFieldsFromResponse(_ context.Context, path path.Path, fields cm.OptEntryFields) (TypedMap[TypedMap[jsontypes.Normalized]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !fields.IsSet() {
		return NewTypedMapNull[TypedMap[jsontypes.Normalized]](), diags
	}

	elements := map[string]TypedMap[jsontypes.Normalized]{}

	for fieldID, fieldValue := range fields.Value {
		localizedValues, localizedValuesDiags := NewEntryLocalizedFieldFromRaw(path.AtMapKey(fieldID), fieldValue)
		diags.Append(localizedValuesDiags...)

		if localizedValuesDiags.HasError() {
			continue
		}

		elements[fieldID] = localizedValues
	}

	return NewTypedMap(elements), diags
}

func NewEntryLocalizedFieldFromRaw(path path.Path, raw []byte) (TypedMap[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if isRawJSONNull(raw) {
		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	var localizedValues map[string]json.RawMessage

	err := json.Unmarshal(raw, &localizedValues)
	if err != nil {
		diags.AddAttributeError(path, "Invalid Entry Field Value", fmt.Sprintf("Expected a JSON object keyed by locale: %s", err))

		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	elements := make(map[string]jsontypes.Normalized, len(localizedValues))
	for locale, value := range localizedValues {
		elements[locale] = NewNormalizedJSONTypesNormalizedValue(value)
	}

	return NewTypedMap(elements), diags
}

func mergeEntryFieldsWithFallback(responseFields, fallbackFields TypedMap[TypedMap[jsontypes.Normalized]]) TypedMap[TypedMap[jsontypes.Normalized]] {
	elements := make(map[string]TypedMap[jsontypes.Normalized], len(responseFields.Elements())+len(fallbackFields.Elements()))

	maps.Copy(elements, responseFields.Elements())

	for key, value := range fallbackFields.Elements() {
		if _, ok := elements[key]; !ok {
			elements[key] = value
		}
	}

	return NewTypedMap(elements)
}

func entryDraftMutationRequired(ctx context.Context, plan, state EntryModel) (bool, diag.Diagnostics) {
	fieldsEquivalent, fieldsDiags := entryFieldsEquivalent(ctx, plan.Fields, state.Fields)

	return !fieldsEquivalent || !entryMetadataEquivalent(plan.Metadata, state.Metadata), fieldsDiags
}

func entryMetadataEquivalent(left, right TypedObject[EntryMetadataValue]) bool {
	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return (left.IsNull() && right.IsNull()) || (left.IsUnknown() && right.IsUnknown())
	}

	return entryStringListUnorderedEquivalent(left.Value().Concepts, right.Value().Concepts) &&
		entryStringListUnorderedEquivalent(left.Value().Tags, right.Value().Tags)
}

// entryStringListUnorderedEquivalent ignores metadata-link ordering while
// preserving duplicate multiplicity. CMA has been observed to return unchanged
// links in an order different from the request.
func entryStringListUnorderedEquivalent(left, right TypedList[types.String]) bool {
	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return left.Equal(right)
	}

	leftElements := left.Elements()

	rightElements := right.Elements()
	if len(leftElements) != len(rightElements) {
		return false
	}

	matched := make([]bool, len(rightElements))

	for _, leftElement := range leftElements {
		found := false

		for index, rightElement := range rightElements {
			if !matched[index] && leftElement.Equal(rightElement) {
				matched[index] = true
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func entryFieldsEquivalent(ctx context.Context, left, right TypedMap[TypedMap[jsontypes.Normalized]]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	left = entryFieldsRequestProjection(left)
	right = entryFieldsRequestProjection(right)

	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return left.Equal(right), diags
	}

	leftElements := left.Elements()

	rightElements := right.Elements()
	if len(leftElements) != len(rightElements) {
		return false, diags
	}

	for key, leftElement := range leftElements {
		rightElement, ok := rightElements[key]
		if !ok {
			return false, diags
		}

		if leftElement.IsNull() || leftElement.IsUnknown() || rightElement.IsNull() || rightElement.IsUnknown() {
			if !leftElement.Equal(rightElement) {
				return false, diags
			}

			continue
		}

		equivalent, elementDiags := entryLocalizedFieldEquivalent(ctx, leftElement, rightElement)
		diags.Append(elementDiags...)

		if diags.HasError() || !equivalent {
			return false, diags
		}
	}

	return true, diags
}

func entryLocalizedFieldEquivalent(ctx context.Context, left, right TypedMap[jsontypes.Normalized]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return left.Equal(right), diags
	}

	leftElements := left.Elements()

	rightElements := right.Elements()
	if len(leftElements) != len(rightElements) {
		return false, diags
	}

	for locale, leftValue := range leftElements {
		rightValue, ok := rightElements[locale]
		if !ok {
			return false, diags
		}

		if leftValue.IsNull() || leftValue.IsUnknown() || rightValue.IsNull() || rightValue.IsUnknown() {
			if !leftValue.Equal(rightValue) {
				return false, diags
			}

			continue
		}

		equivalent, valueDiags := leftValue.StringSemanticEquals(ctx, rightValue)
		diags.Append(valueDiags...)

		if diags.HasError() || !equivalent {
			return false, diags
		}
	}

	return true, diags
}

func mergeEntryResponseFieldsWithOmissionFallback(response, fallback TypedMap[TypedMap[jsontypes.Normalized]]) TypedMap[TypedMap[jsontypes.Normalized]] {
	if response.IsUnknown() || fallback.IsNull() || fallback.IsUnknown() {
		return response
	}

	if response.IsNull() && len(fallback.Elements()) == 0 {
		return fallback
	}

	responseElements := make(map[string]TypedMap[jsontypes.Normalized], len(response.Elements())+len(fallback.Elements()))
	maps.Copy(responseElements, response.Elements())

	changed := false

	for fieldID, value := range fallback.Elements() {
		if _, exists := responseElements[fieldID]; exists ||
			(!value.IsNull() && !entryLocalizedFieldHasOnlyEmptyArrays(value)) {
			continue
		}

		responseElements[fieldID] = value
		changed = true
	}

	if !changed {
		return response
	}

	return NewTypedMap(responseElements)
}

func entryLocalizedFieldHasOnlyEmptyArrays(value TypedMap[jsontypes.Normalized]) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}

	localized := value.Elements()
	if len(localized) == 0 {
		return false
	}

	for _, localeValue := range localized {
		if localeValue.IsNull() || localeValue.IsUnknown() {
			return false
		}

		var decoded any

		err := json.Unmarshal([]byte(localeValue.ValueString()), &decoded)
		if err != nil {
			return false
		}

		items, ok := decoded.([]any)
		if !ok || len(items) != 0 {
			return false
		}
	}

	return true
}

type entryResponseFieldPolicy uint8

const (
	entryResponseFieldsExact entryResponseFieldPolicy = iota
	entryResponseFieldsCreationDefaults
)

func projectEntryMutationResponse(
	ctx context.Context,
	plan, response EntryModel,
	fieldPolicy entryResponseFieldPolicy,
) (EntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	response.Timeouts = plan.Timeouts
	response.Fields = mergeEntryResponseFieldsWithOmissionFallback(response.Fields, plan.Fields)

	if !response.ContentTypeID.Equal(plan.ContentTypeID) {
		diags.AddError("Unexpected entry content type", fmt.Sprintf("Contentful returned content type %q after Terraform sent %q.", response.ContentTypeID.ValueString(), plan.ContentTypeID.ValueString()))
	}

	response.ContentTypeID = plan.ContentTypeID

	fieldsEquivalent, fieldsDiags := entryResponseFieldsConsistent(ctx, plan.Fields, response.Fields, fieldPolicy)
	diags.Append(fieldsDiags...)

	if !fieldsEquivalent {
		diags.AddError("Unexpected entry fields", "Contentful returned entry fields that differ from the effective Terraform plan.")
	}

	if !entryMetadataEquivalent(response.Metadata, plan.Metadata) {
		diags.AddError("Unexpected entry metadata", "Contentful returned entry metadata that differs from the effective Terraform plan.")
	}

	spaceID := plan.SpaceID
	environmentID := plan.EnvironmentID
	entryID := plan.EntryID

	if entryID.IsNull() || entryID.IsUnknown() {
		entryID = response.EntryID
	}

	if !response.SpaceID.Equal(spaceID) || !response.EnvironmentID.Equal(environmentID) || !response.EntryID.Equal(entryID) {
		diags.AddError("Unexpected entry identity", "Contentful returned an entry identity that differs from the mutation endpoint.")
	}

	if !diags.HasError() {
		response.Fields = plan.Fields
		response.Metadata = plan.Metadata
	}

	response.EntryIdentityModel = EntryIdentityModel{SpaceID: spaceID, EnvironmentID: environmentID, EntryID: entryID}
	response.IDIdentityModel = NewIDIdentityModelFromMultipartID(spaceID.ValueString(), environmentID.ValueString(), entryID.ValueString())

	return response, diags
}

func entryResponseFieldsConsistent(
	ctx context.Context,
	plan, response TypedMap[TypedMap[jsontypes.Normalized]],
	policy entryResponseFieldPolicy,
) (bool, diag.Diagnostics) {
	if policy != entryResponseFieldsCreationDefaults {
		return entryFieldsEquivalent(ctx, plan, response)
	}

	if plan.IsNull() || plan.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return plan.Equal(response), nil
	}

	plan = entryFieldsRequestProjection(plan)
	responsePlanFields := make(map[string]TypedMap[jsontypes.Normalized], len(plan.Elements()))

	responseElements := response.Elements()
	for fieldID := range plan.Elements() {
		if responseField, ok := responseElements[fieldID]; ok {
			responsePlanFields[fieldID] = responseField
		}
	}

	return entryFieldsEquivalent(ctx, plan, NewTypedMap(responsePlanFields))
}

func NewEntryMetadataFromResponse(ctx context.Context, _ path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue](), diags
	}

	conceptsValue := NewTypedListNull[types.String]()

	if metadata.Value.Concepts != nil {
		concepts := []types.String{}

		for _, concept := range metadata.Value.Concepts {
			concepts = append(concepts, types.StringValue(concept.Sys.ID))
		}

		conceptsValue = NewTypedList(concepts)
	}

	tagsValue := NewTypedListNull[types.String]()

	if metadata.Value.Tags != nil {
		tags := []types.String{}

		for _, tag := range metadata.Value.Tags {
			tags = append(tags, types.StringValue(tag.Sys.ID))
		}

		tagsValue = NewTypedList(tags)
	}

	obj, objDiags := NewTypedObjectFromAttributes[EntryMetadataValue](ctx, map[string]attr.Value{
		"concepts": conceptsValue,
		"tags":     tagsValue,
	})
	diags.Append(objDiags...)

	return obj, diags
}
