package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToWebhookDefinitionFilter(ctx context.Context, path path.Path, value webhookfilter.WebhookFilterValue) (contentfulManagement.WebhookDefinitionFilter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	encoder := jx.Encoder{}

	diags.Append(encodeWebhookFilterValue(ctx, path, &encoder, value)...)

	return contentfulManagement.WebhookDefinitionFilter(encoder.Bytes()), diags
}

//nolint:cyclop
func encodeWebhookFilterValue(ctx context.Context, path path.Path, encoder *jx.Encoder, value webhookfilter.WebhookFilterValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if !value.Not.IsNull() && !value.Not.IsUnknown() {
			path := path.AtName("not")
			if encoder.Field("not", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterNotValue(ctx, path, encoder, value.Not)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}

		if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
			path := path.AtName("equals")
			if encoder.Field("equals", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterEqualsValue(ctx, path, encoder, value.Equals)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}

		if !value.In.IsNull() && !value.In.IsUnknown() {
			path := path.AtName("in")
			if encoder.Field("in", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterInValue(ctx, path, encoder, value.In)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}

		if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
			path := path.AtName("regexp")
			if encoder.Field("regexp", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterRegexpValue(ctx, path, encoder, value.Regexp)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return diags
}

//nolint:cyclop
func encodeWebhookFilterNotValue(ctx context.Context, path path.Path, encoder *jx.Encoder, value webhookfilter.WebhookFilterNotValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
			path := path.AtName("equals")
			if encoder.Field("equals", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterEqualsValue(ctx, path, encoder, value.Equals)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}

		if !value.In.IsNull() && !value.In.IsUnknown() {
			path := path.AtName("in")
			if encoder.Field("in", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterInValue(ctx, path, encoder, value.In)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}

		if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
			path := path.AtName("regexp")
			if encoder.Field("regexp", func(encoder *jx.Encoder) {
				diags.Append(encodeWebhookFilterRegexpValue(ctx, path, encoder, value.Regexp)...)
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return diags
}

func encodeWebhookFilterEqualsValue(_ context.Context, path path.Path, encoder *jx.Encoder, value webhookfilter.WebhookFilterEqualsValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(value.Doc.ValueString()) }) {
			diags.AddAttributeError(path.AtName("doc"), "failed to encode value", "")
		}

		if encoder.Field("value", func(encoder *jx.Encoder) { encoder.Str(value.Value.ValueString()) }) {
			diags.AddAttributeError(path.AtName("value"), "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return diags
}

func encodeWebhookFilterInValue(ctx context.Context, path path.Path, encoder *jx.Encoder, value webhookfilter.WebhookFilterInValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(value.Doc.ValueString()) }) {
			diags.AddAttributeError(path.AtName("doc"), "failed to encode value", "")
		}

		if encoder.Field("values", func(encoder *jx.Encoder) {
			path := path.AtName("values")
			if encoder.Arr(func(e *jx.Encoder) {
				values := make([]string, len(value.Values.Elements()))
				diags.Append(value.Values.ElementsAs(ctx, &values, false)...)

				for index, v := range values {
					path := path.AtListIndex(index)
					if e.Str(v) {
						diags.AddAttributeError(path, "failed to encode value", "")
					}
				}
			}) {
				diags.AddAttributeError(path, "failed to encode value", "")
			}
		}) {
			diags.AddAttributeError(path.AtName("values"), "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return diags
}

func encodeWebhookFilterRegexpValue(_ context.Context, path path.Path, encoder *jx.Encoder, m webhookfilter.WebhookFilterRegexpValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if encoder.Obj(func(encoder *jx.Encoder) {
		if encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(m.Doc.ValueString()) }) {
			diags.AddAttributeError(path.AtName("doc"), "failed to encode value", "")
		}

		if encoder.Field("pattern", func(encoder *jx.Encoder) { encoder.Str(m.Pattern.ValueString()) }) {
			diags.AddAttributeError(path.AtName("pattern"), "failed to encode value", "")
		}
	}) {
		diags.AddAttributeError(path, "failed to encode value", "")
	}

	return diags
}
