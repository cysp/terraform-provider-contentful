package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToOptNilWebhookDefinitionFilterArray(ctx context.Context, path path.Path, filterValuesList types.List) (contentfulManagement.OptNilWebhookDefinitionFilterArray, diag.Diagnostics) {
	if filterValuesList.IsNull() || filterValuesList.IsUnknown() {
		return contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(), nil
	}

	diags := diag.Diagnostics{}

	filterValues := make([]WebhookFilterValue, len(filterValuesList.Elements()))
	diags.Append(filterValuesList.ElementsAs(ctx, &filterValues, false)...)

	filters := make([]contentfulManagement.WebhookDefinitionFilter, len(filterValues))

	for index, filterValue := range filterValues {
		path := path.AtListIndex(index)

		filter, filterDiags := ToWebhookDefinitionFilter(ctx, path, filterValue)
		diags.Append(filterDiags...)

		filters[index] = filter
	}

	return contentfulManagement.NewOptNilWebhookDefinitionFilterArray(filters), diags
}

func ToWebhookDefinitionFilter(ctx context.Context, path path.Path, value WebhookFilterValue) (contentfulManagement.WebhookDefinitionFilter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	filter := contentfulManagement.WebhookDefinitionFilter{}

	if !value.Not.IsNull() && !value.Not.IsUnknown() {
		path := path.AtName("not")

		filterNot, filterNotDiags := ToWebhookDefinitionFilterNot(ctx, path, value.Not)
		diags.Append(filterNotDiags...)

		filter.Not = filterNot
	}

	if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
		path := path.AtName("equals")

		filterEquals, filterEqualsDiags := ToWebhookDefinitionFilterEquals(ctx, path, value.Equals)
		diags.Append(filterEqualsDiags...)

		filter.Equals = filterEquals
	}

	if !value.In.IsNull() && !value.In.IsUnknown() {
		path := path.AtName("in")

		filterIn, filterInDiags := ToWebhookDefinitionFilterIn(ctx, path, value.In)
		diags.Append(filterInDiags...)

		filter.In = filterIn
	}

	if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
		path := path.AtName("regexp")

		filterRegexp, filterRegexpDiags := ToWebhookDefinitionFilterRegexp(ctx, path, value.Regexp)
		diags.Append(filterRegexpDiags...)

		filter.Regexp = filterRegexp
	}

	return filter, diags
}

func ToWebhookDefinitionFilterNot(ctx context.Context, path path.Path, value WebhookFilterNotValue) (contentfulManagement.OptWebhookDefinitionFilterNot, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	optFilterNot := contentfulManagement.OptWebhookDefinitionFilterNot{}

	if !value.IsNull() && !value.IsUnknown() {
		filterNot := contentfulManagement.WebhookDefinitionFilterNot{}

		if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
			path := path.AtName("equals")

			equals, equalsDiags := ToWebhookDefinitionFilterEquals(ctx, path, value.Equals)
			diags.Append(equalsDiags...)

			filterNot.Equals = equals
		}

		if !value.In.IsNull() && !value.In.IsUnknown() {
			path := path.AtName("in")

			in, inDiags := ToWebhookDefinitionFilterIn(ctx, path, value.In)
			diags.Append(inDiags...)

			filterNot.In = in
		}

		if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
			path := path.AtName("regexp")

			regexp, regexpDiags := ToWebhookDefinitionFilterRegexp(ctx, path, value.Regexp)
			diags.Append(regexpDiags...)

			filterNot.Regexp = regexp
		}

		optFilterNot.SetTo(filterNot)
	}

	return optFilterNot, diags
}

func ToWebhookDefinitionFilterEquals(ctx context.Context, path path.Path, value WebhookFilterEqualsValue) (contentfulManagement.WebhookDefinitionFilterEquals, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	filter := contentfulManagement.WebhookDefinitionFilterEquals{}

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValue, filterTermValueDiags := toWebhookDefinitionFilterTermString(ctx, path.AtName("value"), value.Value)
	diags.Append(filterTermValueDiags...)

	filter = append(filter, filterTermValue)

	return filter, diags
}

func ToWebhookDefinitionFilterIn(ctx context.Context, path path.Path, value WebhookFilterInValue) (contentfulManagement.WebhookDefinitionFilterIn, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	filter := contentfulManagement.WebhookDefinitionFilterIn{}

	filterTermDoc, filterTermDocDiags := toWebhookDefinitionFilterTermStringObject(ctx, path.AtName("doc"), "doc", value.Doc)
	diags.Append(filterTermDocDiags...)

	filter = append(filter, filterTermDoc)

	filterTermValues, filterTermValuesDiags := toWebhookDefinitionFilterTermStringArray(ctx, path.AtName("values"), value.Values)
	diags.Append(filterTermValuesDiags...)

	filter = append(filter, filterTermValues)

	return filter, diags
}

func ToWebhookDefinitionFilterRegexp(ctx context.Context, path path.Path, value WebhookFilterRegexpValue) (contentfulManagement.WebhookDefinitionFilterRegexp, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	filter := contentfulManagement.WebhookDefinitionFilterRegexp{}

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

func toWebhookDefinitionFilterTermStringArray(ctx context.Context, path path.Path, value types.List) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoder := jx.Encoder{}

	if encoder.Arr(func(encoder *jx.Encoder) {
		values := make([]string, len(value.Elements()))
		diags.Append(value.ElementsAs(ctx, &values, false)...)

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

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field(name, func(encoder *jx.Encoder) { encoder.Str(value.ValueString()) }) {
			diags.AddAttributeError(path, "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return encoder.Bytes(), diags
}
