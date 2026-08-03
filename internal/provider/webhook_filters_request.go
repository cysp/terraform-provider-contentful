package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToOptNilWebhookDefinitionFilterArray(
	ctx context.Context,
	valuePath path.Path,
	filterValuesList TypedList[TypedObject[WebhookFilterValue]],
) (cm.OptNilWebhookDefinitionFilterArray, diag.Diagnostics) {
	if filterValuesList.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown webhook filters",
			"Webhook filters must be known before they can be sent to Contentful.",
		)

		return cm.OptNilWebhookDefinitionFilterArray{}, diags
	}

	if filterValuesList.IsNull() {
		return cm.NewOptNilWebhookDefinitionFilterArrayNull(), nil
	}

	filters, diags := convertKnownObjectListElements(
		ctx,
		valuePath,
		filterValuesList.Elements(),
		func(_ context.Context, filterPath path.Path, filterValue WebhookFilterValue) (cm.WebhookDefinitionFilter, diag.Diagnostics) {
			return ToWebhookDefinitionFilter(filterPath, filterValue)
		},
	)
	if diags.HasError() {
		return cm.OptNilWebhookDefinitionFilterArray{}, diags
	}

	return cm.NewOptNilWebhookDefinitionFilterArray(filters), diags
}

func ToWebhookDefinitionFilter(
	valuePath path.Path,
	value WebhookFilterValue,
) (cm.WebhookDefinitionFilter, diag.Diagnostics) {
	notPath := valuePath.AtName("not")
	equalsPath := valuePath.AtName("equals")
	inPath := valuePath.AtName("in")
	regexpPath := valuePath.AtName("regexp")

	return convertExactlyOneKnownAlternative(
		valuePath,
		knownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name:  "not",
			Path:  notPath,
			Value: value.Not,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				notValue, _ := value.Not.GetValue()
				not, diags := ToWebhookDefinitionFilterNot(notPath, notValue)

				return cm.WebhookDefinitionFilter{Not: not}, diags
			},
		},
		knownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name:  "equals",
			Path:  equalsPath,
			Value: value.Equals,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				equalsValue, _ := value.Equals.GetValue()
				equals, diags := ToWebhookDefinitionFilterEquals(equalsPath, equalsValue)

				return cm.WebhookDefinitionFilter{Equals: equals}, diags
			},
		},
		knownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name:  "in",
			Path:  inPath,
			Value: value.In,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				inValue, _ := value.In.GetValue()
				in, diags := ToWebhookDefinitionFilterIn(inPath, inValue)

				return cm.WebhookDefinitionFilter{In: in}, diags
			},
		},
		knownUnionAlternative[cm.WebhookDefinitionFilter]{
			Name:  "regexp",
			Path:  regexpPath,
			Value: value.Regexp,
			Convert: func() (cm.WebhookDefinitionFilter, diag.Diagnostics) {
				regexpValue, _ := value.Regexp.GetValue()
				regexp, diags := ToWebhookDefinitionFilterRegexp(regexpPath, regexpValue)

				return cm.WebhookDefinitionFilter{Regexp: regexp}, diags
			},
		},
	)
}

func ToWebhookDefinitionFilterNot(
	valuePath path.Path,
	value WebhookFilterNotValue,
) (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
	equalsPath := valuePath.AtName("equals")
	inPath := valuePath.AtName("in")
	regexpPath := valuePath.AtName("regexp")

	filterNot, diags := convertExactlyOneKnownAlternative(
		valuePath,
		knownUnionAlternative[cm.WebhookDefinitionFilterNot]{
			Name:  "equals",
			Path:  equalsPath,
			Value: value.Equals,
			Convert: func() (cm.WebhookDefinitionFilterNot, diag.Diagnostics) {
				equalsValue, _ := value.Equals.GetValue()
				equals, diags := ToWebhookDefinitionFilterEquals(equalsPath, equalsValue)

				return cm.WebhookDefinitionFilterNot{Equals: equals}, diags
			},
		},
		knownUnionAlternative[cm.WebhookDefinitionFilterNot]{
			Name:  "in",
			Path:  inPath,
			Value: value.In,
			Convert: func() (cm.WebhookDefinitionFilterNot, diag.Diagnostics) {
				inValue, _ := value.In.GetValue()
				in, diags := ToWebhookDefinitionFilterIn(inPath, inValue)

				return cm.WebhookDefinitionFilterNot{In: in}, diags
			},
		},
		knownUnionAlternative[cm.WebhookDefinitionFilterNot]{
			Name:  "regexp",
			Path:  regexpPath,
			Value: value.Regexp,
			Convert: func() (cm.WebhookDefinitionFilterNot, diag.Diagnostics) {
				regexpValue, _ := value.Regexp.GetValue()
				regexp, diags := ToWebhookDefinitionFilterRegexp(regexpPath, regexpValue)

				return cm.WebhookDefinitionFilterNot{Regexp: regexp}, diags
			},
		},
	)
	if diags.HasError() {
		return cm.OptWebhookDefinitionFilterNot{}, diags
	}

	return cm.NewOptWebhookDefinitionFilterNot(filterNot), diags
}

func ToWebhookDefinitionFilterEquals(
	valuePath path.Path,
	value WebhookFilterEqualsValue,
) (cm.WebhookDefinitionFilterEquals, diag.Diagnostics) {
	doc, docDiags := toWebhookDefinitionFilterTermStringObject(valuePath.AtName("doc"), "doc", value.Doc)
	filterValue, valueDiags := toWebhookDefinitionFilterTermString(valuePath.AtName("value"), value.Value)
	diags := diag.Diagnostics{}
	diags.Append(docDiags...)
	diags.Append(valueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return cm.WebhookDefinitionFilterEquals{doc, filterValue}, diags
}

func ToWebhookDefinitionFilterIn(
	valuePath path.Path,
	value WebhookFilterInValue,
) (cm.WebhookDefinitionFilterIn, diag.Diagnostics) {
	doc, docDiags := toWebhookDefinitionFilterTermStringObject(valuePath.AtName("doc"), "doc", value.Doc)
	values, valuesDiags := toWebhookDefinitionFilterTermStringArray(valuePath.AtName("values"), value.Values)
	diags := diag.Diagnostics{}
	diags.Append(docDiags...)
	diags.Append(valuesDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return cm.WebhookDefinitionFilterIn{doc, values}, diags
}

func ToWebhookDefinitionFilterRegexp(
	valuePath path.Path,
	value WebhookFilterRegexpValue,
) (cm.WebhookDefinitionFilterRegexp, diag.Diagnostics) {
	doc, docDiags := toWebhookDefinitionFilterTermStringObject(valuePath.AtName("doc"), "doc", value.Doc)
	pattern, patternDiags := toWebhookDefinitionFilterTermStringObject(valuePath.AtName("pattern"), "pattern", value.Pattern)
	diags := diag.Diagnostics{}
	diags.Append(docDiags...)
	diags.Append(patternDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return cm.WebhookDefinitionFilterRegexp{doc, pattern}, diags
}

func toWebhookDefinitionFilterTermString(valuePath path.Path, value types.String) (jx.Raw, diag.Diagnostics) {
	stringValue, diags := requestRequiredString(value, valuePath)
	if diags.HasError() {
		return nil, diags
	}

	encoder := jx.Encoder{}
	if encoder.Str(stringValue) {
		diags.AddAttributeError(valuePath, "failed to encode value", "")
	}

	if diags.HasError() {
		return nil, diags
	}

	return encoder.Bytes(), diags
}

func toWebhookDefinitionFilterTermStringArray(
	valuePath path.Path,
	value TypedList[types.String],
) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown webhook filter values",
			"Webhook filter values must be known before they can be sent to Contentful.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null webhook filter values",
			"Required webhook filter values cannot be null.",
		)
	}

	if diags.HasError() {
		return nil, diags
	}

	values, valueDiags := knownStringListElements(valuePath, value.Elements())
	diags.Append(valueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	encoder := jx.Encoder{}
	if encoder.Arr(func(encoder *jx.Encoder) {
		for index, stringValue := range values {
			elementPath := valuePath.AtListIndex(index)
			if encoder.Str(stringValue) {
				diags.AddAttributeError(elementPath, "failed to encode value", "")
			}
		}
	}) {
		diags.AddAttributeError(valuePath, "failed to encode value", "")
	}

	if diags.HasError() {
		return nil, diags
	}

	return encoder.Bytes(), diags
}

func toWebhookDefinitionFilterTermStringObject(
	valuePath path.Path,
	name string,
	value types.String,
) (jx.Raw, diag.Diagnostics) {
	stringValue, diags := requestRequiredString(value, valuePath)
	if diags.HasError() {
		return nil, diags
	}

	encoder := jx.Encoder{}
	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field(name, func(encoder *jx.Encoder) { encoder.Str(stringValue) }) {
			diags.AddAttributeError(valuePath, "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(valuePath, "failed to encode value", "")
	}

	if diags.HasError() {
		return nil, diags
	}

	return encoder.Bytes(), diags
}
