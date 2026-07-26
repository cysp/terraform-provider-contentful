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

const (
	webhookFilterNotIndex = iota
	webhookFilterEqualsIndex
	webhookFilterInIndex
	webhookFilterRegexpIndex
)

const (
	webhookNotFilterEqualsIndex = iota
	webhookNotFilterInIndex
	webhookNotFilterRegexpIndex
)

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

	selected, diags := ExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative{Name: "not", Path: notPath, Value: value.Not},
		KnownUnionAlternative{Name: "equals", Path: equalsPath, Value: value.Equals},
		KnownUnionAlternative{Name: "in", Path: inPath, Value: value.In},
		KnownUnionAlternative{Name: "regexp", Path: regexpPath, Value: value.Regexp},
	)
	if diags.HasError() {
		return cm.WebhookDefinitionFilter{}, diags
	}

	filter := cm.WebhookDefinitionFilter{}

	switch selected {
	case webhookFilterNotIndex:
		notValue, _ := value.Not.GetValue()
		filter.Not, diags = ToWebhookDefinitionFilterNot(ctx, notPath, notValue)
	case webhookFilterEqualsIndex:
		equalsValue, _ := value.Equals.GetValue()
		filter.Equals, diags = ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)
	case webhookFilterInIndex:
		inValue, _ := value.In.GetValue()
		filter.In, diags = ToWebhookDefinitionFilterIn(ctx, inPath, inValue)
	case webhookFilterRegexpIndex:
		regexpValue, _ := value.Regexp.GetValue()
		filter.Regexp, diags = ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)
	}

	if diags.HasError() {
		return cm.WebhookDefinitionFilter{}, diags
	}

	return filter, diags
}

func ToWebhookDefinitionFilterNot(ctx context.Context, valuePath path.Path, value WebhookFilterNotValue) (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
	equalsPath := valuePath.AtName("equals")
	inPath := valuePath.AtName("in")
	regexpPath := valuePath.AtName("regexp")

	selected, diags := ExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative{Name: "equals", Path: equalsPath, Value: value.Equals},
		KnownUnionAlternative{Name: "in", Path: inPath, Value: value.In},
		KnownUnionAlternative{Name: "regexp", Path: regexpPath, Value: value.Regexp},
	)
	if diags.HasError() {
		return cm.OptWebhookDefinitionFilterNot{}, diags
	}

	filterNot := cm.WebhookDefinitionFilterNot{}

	switch selected {
	case webhookNotFilterEqualsIndex:
		equalsValue, _ := value.Equals.GetValue()
		filterNot.Equals, diags = ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)
	case webhookNotFilterInIndex:
		inValue, _ := value.In.GetValue()
		filterNot.In, diags = ToWebhookDefinitionFilterIn(ctx, inPath, inValue)
	case webhookNotFilterRegexpIndex:
		regexpValue, _ := value.Regexp.GetValue()
		filterNot.Regexp, diags = ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)
	}

	if diags.HasError() {
		return cm.OptWebhookDefinitionFilterNot{}, diags
	}

	return cm.NewOptWebhookDefinitionFilterNot(filterNot), diags
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
