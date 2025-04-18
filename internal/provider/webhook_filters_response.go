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

func ReadWebhookFiltersListValueFromResponse(ctx context.Context, path path.Path, optNilFilters cm.OptNilWebhookDefinitionFilterArray) (TypedList[WebhookFilterValue], diag.Diagnostics) {
	filters, filtersOk := optNilFilters.Get()
	if !filtersOk {
		return NewTypedListNull[WebhookFilterValue](ctx), nil
	}

	diags := diag.Diagnostics{}

	filtersElements := make([]WebhookFilterValue, len(filters))

	for index, filter := range filters {
		filtersElement, filtersElementDiags := ReadWebhookFilterValueFromResponse(ctx, path.AtListIndex(index), filter)
		diags.Append(filtersElementDiags...)

		filtersElements[index] = filtersElement
	}

	filtersList, filtersListDiags := NewTypedList(ctx, filtersElements)
	diags.Append(filtersListDiags...)

	return filtersList, diags
}

func ReadWebhookFilterValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilter) (WebhookFilterValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := NewWebhookFilterValueKnown()

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

	return value, diags
}

func ReadWebhookFilterNotValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterNot) (WebhookFilterNotValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := NewWebhookFilterNotValueKnown()

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

	return value, diags
}

func ReadWebhookFilterEqualsValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterEquals) (WebhookFilterEqualsValue, diag.Diagnostics) {
	if input == nil {
		return NewWebhookFilterEqualsValueNull(), nil
	}

	diags := diag.Diagnostics{}

	value := WebhookFilterEqualsValue{}

	//nolint:mnd
	if len(input) == 2 {
		value = NewWebhookFilterEqualsValueKnown()

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", input[0])
		diags.Append(valueDocDiags...)

		value.Doc = valueDoc

		valueValue, valueValueDiags := ReadWebhookDefinitionFilterTermString(ctx, path.AtName("value"), input[1])
		diags.Append(valueValueDiags...)

		value.Value = valueValue
	} else {
		diags.AddAttributeError(path, "failed to decode value", fmt.Sprintf("expected array of length 2, received array of length %d", len(input)))
	}

	return value, diags
}

func ReadWebhookFilterInValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterIn) (WebhookFilterInValue, diag.Diagnostics) {
	if input == nil {
		return NewWebhookFilterInValueNull(), nil
	}

	diags := diag.Diagnostics{}

	value := WebhookFilterInValue{}

	//nolint:mnd
	if len(input) == 2 {
		value = NewWebhookFilterInValueKnown(ctx)

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", input[0])
		diags.Append(valueDocDiags...)

		value.Doc = valueDoc

		valueValues, valueValuesDiags := ReadWebhookDefinitionFilterTermStringArray(ctx, path.AtName("values"), input[1])
		diags.Append(valueValuesDiags...)

		value.Values = valueValues
	} else {
		diags.AddAttributeError(path, "failed to decode value", fmt.Sprintf("expected array of length 2, received array of length %d", len(input)))
	}

	return value, diags
}

func ReadWebhookFilterRegexpValueFromResponse(ctx context.Context, path path.Path, input cm.WebhookDefinitionFilterRegexp) (WebhookFilterRegexpValue, diag.Diagnostics) {
	if input == nil {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	diags := diag.Diagnostics{}

	value := WebhookFilterRegexpValue{}

	//nolint:mnd
	if len(input) == 2 {
		value = NewWebhookFilterRegexpValueKnown()

		valueDoc, valueDocDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", input[0])
		diags.Append(valueDocDiags...)

		value.Doc = valueDoc

		valuePattern, valuePatternDiags := ReadWebhookDefinitionFilterTermStringObject(ctx, path.AtName("pattern"), "pattern", input[1])
		diags.Append(valuePatternDiags...)

		value.Pattern = valuePattern
	} else {
		diags.AddAttributeError(path, "failed to decode value", fmt.Sprintf("expected array of length 2, received array of length %d", len(input)))
	}

	return value, diags
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

func ReadWebhookDefinitionFilterTermStringArray(ctx context.Context, path path.Path, input jx.Raw) (TypedList[types.String], diag.Diagnostics) {
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

	valueValuesList, valueValuesListDiags := NewTypedList(ctx, valueElements)
	diags.Append(valueValuesListDiags...)

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
