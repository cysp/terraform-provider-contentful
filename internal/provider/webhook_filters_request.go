package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToOptNilWebhookDefinitionFilterArray(ctx context.Context, path path.Path, filterValuesList TypedList[TypedObject[WebhookFilterValue]]) (cm.OptNilWebhookDefinitionFilterArray, diag.Diagnostics) {
	if filterValuesList.IsNull() || filterValuesList.IsUnknown() {
		return cm.NewOptNilWebhookDefinitionFilterArrayNull(), nil
	}

	diags := diag.Diagnostics{}

	filterValues := filterValuesList.Elements()

	filters := make([]cm.WebhookDefinitionFilter, len(filterValues))

	for index, filterValue := range filterValues {
		filterPath := path.AtListIndex(index)

		filter, filterDiags := ToWebhookDefinitionFilter(ctx, filterPath, filterValue.Value())
		diags.Append(filterDiags...)

		filters[index] = filter
	}

	return cm.NewOptNilWebhookDefinitionFilterArray(filters), diags
}

func ToWebhookDefinitionFilter(ctx context.Context, path path.Path, value WebhookFilterValue) (cm.WebhookDefinitionFilter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := cm.WebhookDefinitionFilter{}

	notValue, notValueOk := value.Not.GetValue()
	if notValueOk {
		notPath := path.AtName("not")

		filterNot, filterNotDiags := ToWebhookDefinitionFilterNot(ctx, notPath, notValue)
		diags.Append(filterNotDiags...)

		filter.Not = filterNot
	}

	equalsValue, equalsValueOk := value.Equals.GetValue()
	if equalsValueOk {
		equalsPath := path.AtName("equals")

		filterEquals, filterEqualsDiags := ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)
		diags.Append(filterEqualsDiags...)

		filter.Equals = filterEquals
	}

	inValue, inValueOk := value.In.GetValue()
	if inValueOk {
		inPath := path.AtName("in")

		filterIn, filterInDiags := ToWebhookDefinitionFilterIn(ctx, inPath, inValue)
		diags.Append(filterInDiags...)

		filter.In = filterIn
	}

	regexpValue, regexpValueOk := value.Regexp.GetValue()
	if regexpValueOk {
		regexpPath := path.AtName("regexp")

		filterRegexp, filterRegexpDiags := ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)
		diags.Append(filterRegexpDiags...)

		filter.Regexp = filterRegexp
	}

	return filter, diags
}

func ToWebhookDefinitionFilterNot(ctx context.Context, path path.Path, value WebhookFilterNotValue) (cm.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filterNot := cm.WebhookDefinitionFilterNot{}

	if equalsValue, equalsValueOk := value.Equals.GetValue(); equalsValueOk {
		equalsPath := path.AtName("equals")

		equals, equalsDiags := ToWebhookDefinitionFilterEquals(ctx, equalsPath, equalsValue)
		diags.Append(equalsDiags...)

		filterNot.Equals = equals
	}

	if inValue, inValueOk := value.In.GetValue(); inValueOk {
		inPath := path.AtName("in")

		in, inDiags := ToWebhookDefinitionFilterIn(ctx, inPath, inValue)
		diags.Append(inDiags...)

		filterNot.In = in
	}

	if regexpValue, regexpValueOk := value.Regexp.GetValue(); regexpValueOk {
		regexpPath := path.AtName("regexp")

		regexp, regexpDiags := ToWebhookDefinitionFilterRegexp(ctx, regexpPath, regexpValue)
		diags.Append(regexpDiags...)

		filterNot.Regexp = regexp
	}

	return cm.NewOptWebhookDefinitionFilterNot(filterNot), diags
}

func ToWebhookDefinitionFilterEquals(ctx context.Context, path path.Path, value WebhookFilterEqualsValue) (cm.WebhookDefinitionFilterEquals, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := cm.WebhookDefinitionFilterEquals{}

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValue, filterTermValueDiags := toWebhookDefinitionFilterTermString(ctx, path.AtName("value"), value.Value)
	diags.Append(filterTermValueDiags...)

	filter = append(filter, filterTermValue)

	return filter, diags
}

func ToWebhookDefinitionFilterIn(ctx context.Context, path path.Path, value WebhookFilterInValue) (cm.WebhookDefinitionFilterIn, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := cm.WebhookDefinitionFilterIn{}

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValues, filterTermValuesDiags := toWebhookDefinitionFilterTermStringArray(ctx, path.AtName("values"), value.Values)
	diags.Append(filterTermValuesDiags...)

	filter = append(filter, filterTermValues)

	return filter, diags
}

func ToWebhookDefinitionFilterRegexp(ctx context.Context, path path.Path, value WebhookFilterRegexpValue) (cm.WebhookDefinitionFilterRegexp, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := cm.WebhookDefinitionFilterRegexp{}

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermPattern, filterTermPatternDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("pattern"), "pattern", value.Pattern)
	diags.Append(filterTermPatternDiags...)

	filter = append(filter, filterTermPattern)

	return filter, diags
}

func toWebhookDefinitionFilterTermString(_ context.Context, path path.Path, value types.String) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	if encoder.Str(value.ValueString()) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}

func toWebhookDefinitionFilterTermStringArray(ctx context.Context, path path.Path, value TypedList[types.String]) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	if encoder.Arr(func(encoder *jx.Encoder) {
		values := make([]string, len(value.Elements()))
		diags.Append(tfsdk.ValueAs(ctx, value, &values)...)

		for index, v := range values {
			indexPath := path.AtListIndex(index)
			if encoder.Str(v) {
				diags.AddAttributeError(indexPath, "failed to encode value", "")
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

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field(name, func(encoder *jx.Encoder) { encoder.Str(value.ValueString()) }) {
			diags.AddAttributeError(path, "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}
