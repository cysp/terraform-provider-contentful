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

	model.Metadata = NewEntryMetadataFromResponse(entry.Metadata)
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
		localizedValues, localizedValuesDiags := entryLocalizedFieldFromJSON(path.AtMapKey(fieldID), fieldValue)
		diags.Append(localizedValuesDiags...)

		if localizedValuesDiags.HasError() {
			continue
		}

		elements[fieldID] = localizedValues
	}

	return NewTypedMap(elements), diags
}

func entryLocalizedFieldFromJSON(path path.Path, raw []byte) (TypedMap[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var localizedValues map[string]json.RawMessage

	err := json.Unmarshal(raw, &localizedValues)
	if err != nil {
		diags.AddAttributeError(path, "Invalid Entry Field Value", fmt.Sprintf("Expected a JSON object keyed by locale: %s", err))

		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	if localizedValues == nil {
		diags.AddAttributeError(path, "Unrepresentable Entry field", "The localized Entry fields schema cannot represent a whole-field JSON null. A Terraform-null field omits the field from the request. A JSON-null translation belongs under its locale code.")

		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	elements := make(map[string]jsontypes.Normalized, len(localizedValues))
	for locale, value := range localizedValues {
		elements[locale] = NewNormalizedJSONValue(value)
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

	// CMA has been observed to reorder unchanged metadata links.
	return unorderedStringListsEquivalent(left.Value().Concepts, right.Value().Concepts) &&
		unorderedStringListsEquivalent(left.Value().Tags, right.Value().Tags)
}

func entryFieldsEquivalent(ctx context.Context, left, right TypedMap[TypedMap[jsontypes.Normalized]]) (bool, diag.Diagnostics) {
	return entryMapValuesEquivalent(entryFieldsRequestProjection(left), entryFieldsRequestProjection(right), func(left, right TypedMap[jsontypes.Normalized]) (bool, diag.Diagnostics) {
		return entryMapValuesEquivalent(left, right, func(left, right jsontypes.Normalized) (bool, diag.Diagnostics) {
			return normalizedJSONEquivalent(ctx, left, right)
		})
	})
}

func entryMapValuesEquivalent[T attr.Value](left, right TypedMap[T], equal func(T, T) (bool, diag.Diagnostics)) (bool, diag.Diagnostics) {
	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return left.Equal(right), nil
	}

	if len(left.Elements()) != len(right.Elements()) {
		return false, nil
	}

	var diagnostics diag.Diagnostics

	for key, value := range left.Elements() {
		other, ok := right.Elements()[key]
		if !ok {
			return false, diagnostics
		}

		equivalent, diags := equal(value, other)
		diagnostics.Append(diags...)

		if !equivalent || diagnostics.HasError() {
			return false, diagnostics
		}
	}

	return true, diagnostics
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
		if returned, exists := responseElements[fieldID]; exists {
			if merged, restored := restoreOmittedEntryLocales(returned, value); restored {
				responseElements[fieldID] = merged
				changed = true
			}

			continue
		}

		if !value.IsNull() && !entryLocalizedFieldHasOnlyEmptyArrays(value) {
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

func restoreOmittedEntryLocales(response, prior TypedMap[jsontypes.Normalized]) (TypedMap[jsontypes.Normalized], bool) {
	if response.IsNull() || response.IsUnknown() {
		return response, false
	}

	var restored map[string]jsontypes.Normalized

	for code, value := range prior.Elements() {
		if _, present := response.Elements()[code]; present || !value.IsNull() {
			continue
		}

		if restored == nil {
			restored = make(map[string]jsontypes.Normalized, len(response.Elements())+1)
			maps.Copy(restored, response.Elements())
		}

		restored[code] = value
	}

	if restored == nil {
		return response, false
	}

	return NewTypedMap(restored), true
}

func entryLocalizedFieldHasOnlyEmptyArrays(value TypedMap[jsontypes.Normalized]) bool {
	sent := false

	for _, locale := range value.Elements() {
		if locale.IsNull() {
			continue
		}

		if locale.IsUnknown() {
			return false
		}

		var array []json.RawMessage

		err := json.Unmarshal([]byte(locale.ValueString()), &array)
		if err != nil || array == nil || len(array) != 0 {
			return false
		}

		sent = true
	}

	return sent
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
) (EntryModel, diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	response.Timeouts = plan.Timeouts
	response.Fields = mergeEntryResponseFieldsWithOmissionFallback(response.Fields, plan.Fields)

	if !response.ContentTypeID.Equal(plan.ContentTypeID) {
		diags.AddError("Unexpected entry content type", fmt.Sprintf("Contentful returned content type %q after Terraform sent %q.", response.ContentTypeID.ValueString(), plan.ContentTypeID.ValueString()))
	}

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

	identityConfirmed := response.SpaceID.Equal(spaceID) && response.EnvironmentID.Equal(environmentID) && response.EntryID.Equal(entryID)
	if !identityConfirmed {
		diags.AddError("Unexpected entry identity", "Contentful returned an entry identity that differs from the mutation endpoint.")
	}

	if !diags.HasError() {
		response.Fields = plan.Fields
		response.Metadata = plan.Metadata
	}

	response.EntryIdentityModel = EntryIdentityModel{SpaceID: spaceID, EnvironmentID: environmentID, EntryID: entryID}
	response.IDIdentityModel = NewIDIdentityModelFromMultipartID(spaceID.ValueString(), environmentID.ValueString(), entryID.ValueString())

	return response, diags, identityConfirmed
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

func NewEntryMetadataFromResponse(metadata cm.OptEntryMetadata) TypedObject[EntryMetadataValue] {
	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue]()
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

	return NewTypedObject(EntryMetadataValue{
		Concepts: conceptsValue,
		Tags:     tagsValue,
	})
}
