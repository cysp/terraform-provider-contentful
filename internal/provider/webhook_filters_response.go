package provider

import (
	"context"
	"fmt"

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

	if diags.HasError() {
		return NewTypedListNull[TypedObject[WebhookFilterValue]](), diags
	}

	filtersList := NewTypedList(filtersElements)

	return filtersList, diags
}

func ReadWebhookFilterValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilter) (TypedObject[WebhookFilterValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

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

	if diags.HasError() {
		return NewTypedObjectNull[WebhookFilterValue](), diags
	}

	return NewTypedObject(value), diags
}

func ReadWebhookFilterNotValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterNot) (TypedObject[WebhookFilterNotValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

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

	if diags.HasError() {
		return NewTypedObjectNull[WebhookFilterNotValue](), diags
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
		diags.AddAttributeError(path, "failed to decode value", fmt.Sprintf("expected array of length 2, received array of length %d", len(input)))

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
	decoder := jx.DecodeBytes(input)

	valueValue, valueValueErr := decoder.Str()
	if valueValueErr != nil {
		diags.AddAttributeError(path, "failed to decode value", valueValueErr.Error())
	}

	decoder.Next()

	return types.StringValue(valueValue), diags
}

func ReadWebhookDefinitionFilterTermStringArray(_ context.Context, path path.Path, input jx.Raw) (TypedList[types.String], diag.Diagnostics) {
	diags := diag.Diagnostics{}
	decoder := jx.DecodeBytes(input)

	valueElements := make([]types.String, 0)

	arrDecodeErr := decoder.Arr(func(decoder *jx.Decoder) error {
		valueValue, valueValueErr := decoder.Str()
		if valueValueErr != nil {
			//nolint:wrapcheck
			return valueValueErr
		}

		valueElements = append(valueElements, types.StringValue(valueValue))

		return nil
	})
	if arrDecodeErr != nil {
		diags.AddAttributeError(path, "failed to decode value", "")
	}

	valueValuesList := NewTypedList(valueElements)

	return valueValuesList, diags
}

func ReadWebhookDefinitionFilterTermStringObject(_ context.Context, path path.Path, name string, input jx.Raw) (types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	decoder := jx.DecodeBytes(input)

	value := types.StringNull()

	objDecodeErr := decoder.Obj(func(decoder *jx.Decoder, key string) error {
		if key != name {
			return decoder.Skip()
		}

		valuePattern, valuePatternErr := decoder.Str()
		if valuePatternErr != nil {
			//nolint:wrapcheck
			return valuePatternErr
		}

		value = types.StringValue(valuePattern)

		return nil
	})
	if objDecodeErr != nil {
		diags.AddAttributeError(path, "failed to decode value", objDecodeErr.Error())
	}

	return value, diags
}
