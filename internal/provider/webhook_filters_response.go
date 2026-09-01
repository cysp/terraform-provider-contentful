package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ReadWebhookFiltersListValueFromResponse(ctx context.Context, path path.Path, optNilFilters cm.OptNilWebhookDefinitionFilterArray) (TypedList[TypedObject[WebhookFilterValue]], diag.Diagnostics) {
	filters, filtersOk := optNilFilters.Get()
	if !filtersOk {
		return NewTypedListNull[TypedObject[WebhookFilterValue]](), nil
	}

	diags := diag.Diagnostics{}

	filtersElements := make([]TypedObject[WebhookFilterValue], len(filters))

	for index, filter := range filters {
		filtersElement, filtersElementDiags := ReadWebhookFilterValueFromResponse(ctx, path.AtListIndex(index), filter)
		diags.Append(filtersElementDiags...)

		filtersElements[index] = filtersElement
	}

	filtersList := NewTypedList(filtersElements)

	return filtersList, diags
}

func ReadWebhookFilterValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilter) (TypedObject[WebhookFilterValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}
	addUnsupportedWebhookFilterPropertiesWarning(&diags, path, input.AdditionalProps)

	value := WebhookFilterValue{
		Not:    NewTypedObjectNull[WebhookFilterNotValue](),
		Equals: NewTypedObjectNull[WebhookFilterEqualsValue](),
		In:     NewTypedObjectNull[WebhookFilterInValue](),
		Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
	}

	if filterNot, ok := input.Not.Get(); ok {
		filterNotValue, filterNotValueDiags := ReadWebhookFilterNotValueFromResponse(ctx, path.AtName("not"), filterNot)
		diags.Append(filterNotValueDiags...)

		value.Not = filterNotValue
	}

	if input.Equals != nil {
		filterEqualsValue, filterEqualsValueDiags := ReadWebhookFilterEqualsValueFromResponse(ctx, path.AtName("equals"), input.Equals)
		diags.Append(filterEqualsValueDiags...)

		value.Equals = filterEqualsValue
	}

	if input.In != nil {
		filterInValue, filterInValueDiags := ReadWebhookFilterInValueFromResponse(ctx, path.AtName("in"), input.In)
		diags.Append(filterInValueDiags...)

		value.In = filterInValue
	}

	if input.Regexp != nil {
		filterRegexpValue, filterRegexpValueDiags := ReadWebhookFilterRegexpValueFromResponse(ctx, path.AtName("regexp"), input.Regexp)
		diags.Append(filterRegexpValueDiags...)

		value.Regexp = filterRegexpValue
	}

	return NewTypedObject(value), diags
}

func ReadWebhookFilterNotValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterNot) (TypedObject[WebhookFilterNotValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}
	addUnsupportedWebhookFilterPropertiesWarning(&diags, path, input.AdditionalProps)

	value := WebhookFilterNotValue{
		Equals: NewTypedObjectNull[WebhookFilterEqualsValue](),
		In:     NewTypedObjectNull[WebhookFilterInValue](),
		Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
	}

	if input.Equals != nil {
		filterEqualsValue, filterEqualsValueDiags := ReadWebhookFilterEqualsValueFromResponse(ctx, path.AtName("equals"), input.Equals)
		diags.Append(filterEqualsValueDiags...)

		value.Equals = filterEqualsValue
	}

	if input.In != nil {
		filterInValue, filterInValueDiags := ReadWebhookFilterInValueFromResponse(ctx, path.AtName("in"), input.In)
		diags.Append(filterInValueDiags...)

		value.In = filterInValue
	}

	if input.Regexp != nil {
		filterRegexpValue, filterRegexpValueDiags := ReadWebhookFilterRegexpValueFromResponse(ctx, path.AtName("regexp"), input.Regexp)
		diags.Append(filterRegexpValueDiags...)

		value.Regexp = filterRegexpValue
	}

	return NewTypedObject(value), diags
}

func ReadWebhookFilterEqualsValueFromResponse(ctx context.Context, valuePath path.Path, input cm.WebhookDefinitionFilterEquals) (TypedObject[WebhookFilterEqualsValue], diag.Diagnostics) {
	return readWebhookBinaryFilterValue(ctx, valuePath, input, func(ctx context.Context, filterPath path.Path, doc, term jx.Raw) (WebhookFilterEqualsValue, diag.Diagnostics) {
		diags := diag.Diagnostics{}

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, filterPath.AtName("doc"), "doc", doc)
		diags.Append(valueDocDiags...)

		value, valueDiags := ReadWebhookDefinitionFilterTermString(ctx, filterPath.AtName("value"), term)
		diags.Append(valueDiags...)

		return WebhookFilterEqualsValue{Doc: valueDoc, Value: value}, diags
	})
}

func ReadWebhookFilterInValueFromResponse(ctx context.Context, valuePath path.Path, input cm.WebhookDefinitionFilterIn) (TypedObject[WebhookFilterInValue], diag.Diagnostics) {
	return readWebhookBinaryFilterValue(ctx, valuePath, input, func(ctx context.Context, filterPath path.Path, doc, term jx.Raw) (WebhookFilterInValue, diag.Diagnostics) {
		diags := diag.Diagnostics{}

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, filterPath.AtName("doc"), "doc", doc)
		diags.Append(valueDocDiags...)

		values, valuesDiags := ReadWebhookDefinitionFilterTermStringArray(ctx, filterPath.AtName("values"), term)
		diags.Append(valuesDiags...)

		return WebhookFilterInValue{Doc: valueDoc, Values: values}, diags
	})
}

func ReadWebhookFilterRegexpValueFromResponse(ctx context.Context, valuePath path.Path, input cm.WebhookDefinitionFilterRegexp) (TypedObject[WebhookFilterRegexpValue], diag.Diagnostics) {
	return readWebhookBinaryFilterValue(ctx, valuePath, input, func(ctx context.Context, filterPath path.Path, doc, term jx.Raw) (WebhookFilterRegexpValue, diag.Diagnostics) {
		diags := diag.Diagnostics{}

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, filterPath.AtName("doc"), "doc", doc)
		diags.Append(valueDocDiags...)

		pattern, patternDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, filterPath.AtName("pattern"), "pattern", term)
		diags.Append(patternDiags...)

		return WebhookFilterRegexpValue{Doc: valueDoc, Pattern: pattern}, diags
	})
}

func readWebhookBinaryFilterValue[T any](
	ctx context.Context,
	path path.Path,
	input []jx.Raw,
	decode func(context.Context, path.Path, jx.Raw, jx.Raw) (T, diag.Diagnostics),
) (TypedObject[T], diag.Diagnostics) {
	if input == nil {
		return NewTypedObjectNull[T](), nil
	}

	//nolint:mnd
	if len(input) != 2 {
		diags := diag.Diagnostics{}
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", fmt.Sprintf("Contentful returned an array of length %d; this filter requires exactly two terms. Terraform state retains a null alternative; a later request conversion will reject it until configured.", len(input)))

		return NewTypedObjectNull[T](), diags
	}

	value, diags := decode(ctx, path, input[0], input[1])

	if diags.HasError() {
		return NewTypedObjectNull[T](), diags
	}

	return NewTypedObject(value), diags
}

func ReadWebhookDefinitionFilterTermString(_ context.Context, path path.Path, input jx.Raw) (types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !json.Valid(input) {
		diags.AddAttributeError(path, "Failed to decode webhook filter value", "Contentful returned invalid JSON.")

		return types.StringNull(), diags
	}

	var value *string

	err := json.Unmarshal(input, &value)
	if err != nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", "Contentful returned a JSON value where this filter requires a string. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return types.StringNull(), diags
	}

	if value == nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", "Contentful returned JSON null where this filter requires a string. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return types.StringNull(), diags
	}

	return types.StringValue(*value), diags
}

func ReadWebhookDefinitionFilterTermStringArray(_ context.Context, path path.Path, input jx.Raw) (TypedList[types.String], diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !json.Valid(input) {
		diags.AddAttributeError(path, "Failed to decode webhook filter values", "Contentful returned invalid JSON.")

		return NewTypedListNull[types.String](), diags
	}

	var rawElements []json.RawMessage

	err := json.Unmarshal(input, &rawElements)
	if err != nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter values", "Contentful returned a JSON value where this filter requires an array of strings. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return NewTypedListNull[types.String](), diags
	}

	if rawElements == nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter values", "Contentful returned JSON null where this filter requires an array of strings. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return NewTypedListNull[types.String](), diags
	}

	valueElements := make([]types.String, len(rawElements))
	for index, rawElement := range rawElements {
		var value *string

		err := json.Unmarshal(rawElement, &value)
		if err != nil {
			diags.AddAttributeWarning(path.AtListIndex(index), "Unsupported webhook filter value", "Contentful returned a non-string array element. Terraform state retains a null element; a later request conversion will reject it until configured.")
			valueElements[index] = types.StringNull()

			continue
		}

		if value == nil {
			diags.AddAttributeWarning(path.AtListIndex(index), "Unsupported webhook filter value", "Contentful returned a null array element where this filter requires a string. Terraform state retains a null element; a later request conversion will reject it until configured.")
			valueElements[index] = types.StringNull()

			continue
		}

		valueElements[index] = types.StringValue(*value)
	}

	return NewTypedList(valueElements), diags
}

func ReadWebhookDefinitionFilterTermStringObject(_ context.Context, path path.Path, name string, input jx.Raw) (types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !json.Valid(input) {
		diags.AddAttributeError(path, "Failed to decode webhook filter value", "Contentful returned invalid JSON.")

		return types.StringNull(), diags
	}

	var object map[string]json.RawMessage

	err := json.Unmarshal(input, &object)
	if err != nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", "Contentful returned a JSON value where this filter requires an object. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return types.StringNull(), diags
	}

	if object == nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", "Contentful returned JSON null where this filter requires an object. Terraform state retains a null value; a later request conversion will reject it until configured.")

		return types.StringNull(), diags
	}

	rawValue, ok := object[name]
	if !ok {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", fmt.Sprintf("Contentful returned an object without the required %q property. Terraform state retains a null value; a later request conversion will reject it until configured.", name))

		return types.StringNull(), diags
	}

	if len(object) > 1 {
		unsupportedProperties := make([]string, 0, len(object)-1)
		for propertyName := range object {
			if propertyName != name {
				unsupportedProperties = append(unsupportedProperties, propertyName)
			}
		}

		slices.Sort(unsupportedProperties)
		diags.AddAttributeWarning(path, "Unsupported webhook filter term response properties", fmt.Sprintf("Contentful returned properties %q alongside the required %q property. The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.", unsupportedProperties, name))
	}

	var value *string

	err = json.Unmarshal(rawValue, &value)
	if err != nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", fmt.Sprintf("Contentful returned a non-string %q property. Terraform state retains a null value; a later request conversion will reject it until configured.", name))

		return types.StringNull(), diags
	}

	if value == nil {
		diags.AddAttributeWarning(path, "Unsupported webhook filter value", fmt.Sprintf("Contentful returned a null %q property where this filter requires a string. Terraform state retains a null value; a later request conversion will reject it until configured.", name))

		return types.StringNull(), diags
	}

	return types.StringValue(*value), diags
}

func addUnsupportedWebhookFilterPropertiesWarning(diags *diag.Diagnostics, valuePath path.Path, properties map[string]jx.Raw) {
	if len(properties) == 0 {
		return
	}

	propertyNames := make([]string, 0, len(properties))
	for propertyName := range properties {
		propertyNames = append(propertyNames, propertyName)
	}

	slices.Sort(propertyNames)
	diags.AddAttributeWarning(valuePath, "Unsupported webhook filter response properties", fmt.Sprintf("Contentful returned webhook filter properties %q that this provider cannot represent. The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.", propertyNames))
}
