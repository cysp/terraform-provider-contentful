package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const webhookDefinitionBinaryFilterTermCount = 2

func ToOptNilWebhookDefinitionFilterArray(ctx context.Context, path path.Path, filterValuesList TypedList[TypedObject[WebhookFilterValue]]) (cm.OptNilWebhookDefinitionFilterArray, diag.Diagnostics) {
	if filterValuesList.IsNull() {
		return cm.NewOptNilWebhookDefinitionFilterArrayNull(), nil
	}

	diags := diag.Diagnostics{}

	if filterValuesList.IsUnknown() {
		diags.AddAttributeError(path, "Unexpected unknown filters", "Webhook filters must be known before they can be sent to Contentful.")

		return cm.OptNilWebhookDefinitionFilterArray{}, diags
	}

	filters, filterDiags := ConvertKnownObjectListElements(ctx, path, filterValuesList.Elements(), ToWebhookDefinitionFilter)
	diags.Append(filterDiags...)

	if filterDiags.HasError() {
		return cm.OptNilWebhookDefinitionFilterArray{}, diags
	}

	return cm.NewOptNilWebhookDefinitionFilterArray(filters), diags
}

func ToWebhookDefinitionFilter(ctx context.Context, valuePath path.Path, value WebhookFilterValue) (cm.WebhookDefinitionFilter, diag.Diagnostics) {
	notPath := valuePath.AtName("not")
	equalsPath := valuePath.AtName("equals")
	inPath := valuePath.AtName("in")
	regexpPath := valuePath.AtName("regexp")

	return ConvertExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name: "not", Path: notPath, Value: value.Not,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				notValue, _ := value.Not.GetValue()
				not, diags := ToWebhookDefinitionFilterNot(ctx, notPath, notValue)

				return cm.WebhookDefinitionFilter{Not: not}, diags
			},
		},
		KnownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name: "equals", Path: equalsPath, Value: value.Equals,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				equalsValue, _ := value.Equals.GetValue()
				equals, diags := ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)

				return cm.WebhookDefinitionFilter{Equals: equals}, diags
			},
		},
		KnownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name: "in", Path: inPath, Value: value.In,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				inValue, _ := value.In.GetValue()
				in, diags := ToWebhookDefinitionFilterIn(ctx, inPath, inValue)

				return cm.WebhookDefinitionFilter{In: in}, diags
			},
		},
		KnownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name: "regexp", Path: regexpPath, Value: value.Regexp,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				regexpValue, _ := value.Regexp.GetValue()
				regexp, diags := ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)

				return cm.WebhookDefinitionFilter{Regexp: regexp}, diags
			},
		},
	)
}

func ToWebhookDefinitionFilterNot(ctx context.Context, valuePath path.Path, value WebhookFilterNotValue) (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
	equalsPath := valuePath.AtName("equals")
	inPath := valuePath.AtName("in")
	regexpPath := valuePath.AtName("regexp")

	return ConvertExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative[cm.OptWebhookDefinitionFilterNot]{
			Name: "equals", Path: equalsPath, Value: value.Equals,
			Convert: func() (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
				equalsValue, _ := value.Equals.GetValue()
				equals, diags := ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)

				return cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{Equals: equals}), diags
			},
		},
		KnownUnionAlternative[cm.OptWebhookDefinitionFilterNot]{
			Name: "in", Path: inPath, Value: value.In,
			Convert: func() (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
				inValue, _ := value.In.GetValue()
				in, diags := ToWebhookDefinitionFilterIn(ctx, inPath, inValue)

				return cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{In: in}), diags
			},
		},
		KnownUnionAlternative[cm.OptWebhookDefinitionFilterNot]{
			Name: "regexp", Path: regexpPath, Value: value.Regexp,
			Convert: func() (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
				regexpValue, _ := value.Regexp.GetValue()
				regexp, diags := ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)

				return cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{Regexp: regexp}), diags
			},
		},
	)
}

func ToWebhookDefinitionFilterEquals(ctx context.Context, path path.Path, value WebhookFilterEqualsValue) (cm.WebhookDefinitionFilterEquals, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := make(cm.WebhookDefinitionFilterEquals, 0, webhookDefinitionBinaryFilterTermCount)

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValue, filterTermValueDiags := toWebhookDefinitionFilterTermString(ctx, path.AtName("value"), value.Value)
	diags.Append(filterTermValueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	filter = append(filter, filterTermValue)

	return filter, diags
}

func ToWebhookDefinitionFilterIn(ctx context.Context, path path.Path, value WebhookFilterInValue) (cm.WebhookDefinitionFilterIn, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := make(cm.WebhookDefinitionFilterIn, 0, webhookDefinitionBinaryFilterTermCount)

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValues, filterTermValuesDiags := toWebhookDefinitionFilterTermStringArray(ctx, path.AtName("values"), value.Values)
	diags.Append(filterTermValuesDiags...)

	if diags.HasError() {
		return nil, diags
	}

	filter = append(filter, filterTermValues)

	return filter, diags
}

func ToWebhookDefinitionFilterRegexp(ctx context.Context, path path.Path, value WebhookFilterRegexpValue) (cm.WebhookDefinitionFilterRegexp, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := make(cm.WebhookDefinitionFilterRegexp, 0, webhookDefinitionBinaryFilterTermCount)

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermPattern, filterTermPatternDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("pattern"), "pattern", value.Pattern)
	diags.Append(filterTermPatternDiags...)

	if diags.HasError() {
		return nil, diags
	}

	filter = append(filter, filterTermPattern)

	return filter, diags
}

func toWebhookDefinitionFilterTermString(_ context.Context, path path.Path, value types.String) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	stringValue, valueDiags := KnownStringValue(value, path)
	diags.Append(valueDiags...)

	if valueDiags.HasError() {
		return nil, diags
	}

	if encoder.Str(stringValue) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}

func toWebhookDefinitionFilterTermStringArray(_ context.Context, path path.Path, value TypedList[types.String]) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	if value.IsNull() || value.IsUnknown() {
		if value.IsUnknown() {
			diags.AddAttributeError(path, "Unexpected unknown filter values", "Webhook filter values must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path, "Unexpected null filter values", "Webhook filter values cannot be null.")
		}

		return nil, diags
	}

	values, valueDiags := KnownStringValues(value.Elements(), path)
	diags.Append(valueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if encoder.Arr(func(encoder *jx.Encoder) {
		for index, v := range values {
			path := path.AtListIndex(index)
			if encoder.Str(v) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}

func toWebhookDefinitionFilterTermStringObject(_ context.Context, path path.Path, name string, value types.String) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	stringValue, valueDiags := KnownStringValue(value, path)
	diags.Append(valueDiags...)

	if valueDiags.HasError() {
		return nil, diags
	}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field(name, func(encoder *jx.Encoder) { encoder.Str(stringValue) }) {
			diags.AddAttributeError(path, "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}
