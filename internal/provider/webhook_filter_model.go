package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: webhookfilter.WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: webhookfilter.WebhookFilterValue{}.CustomType(ctx),
		},
		Optional: optional,
	}
}

func ToWebhookDefinitionFilter(ctx context.Context, value webhookfilter.WebhookFilterValue) (contentfulManagement.WebhookDefinitionFilter, diag.Diagnostics) {
	encoder := jx.Encoder{}

	encodeWebhookFilterValue(ctx, &encoder, value)

	return contentfulManagement.WebhookDefinitionFilter(encoder.Bytes()), nil
}

func encodeWebhookFilterValue(ctx context.Context, encoder *jx.Encoder, value webhookfilter.WebhookFilterValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		if !value.Not.IsNull() && !value.Not.IsUnknown() {
			encoder.Field("not", func(encoder *jx.Encoder) {
				encodeWebhookFilterNotValue(ctx, encoder, value.Not)
			})
		}

		if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
			encoder.Field("equals", func(encoder *jx.Encoder) {
				encodeWebhookFilterEqualsValue(ctx, encoder, value.Equals)
			})
		}

		if !value.In.IsNull() && !value.In.IsUnknown() {
			encoder.Field("in", func(encoder *jx.Encoder) {
				encodeWebhookFilterInValue(ctx, encoder, value.In)
			})
		}

		if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
			encoder.Field("regexp", func(encoder *jx.Encoder) {
				encodeWebhookFilterRegexpValue(ctx, encoder, value.Regexp)
			})
		}
	})
}

func encodeWebhookFilterNotValue(ctx context.Context, encoder *jx.Encoder, value webhookfilter.WebhookFilterNotValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		if !value.Equals.IsNull() && !value.Equals.IsUnknown() {
			encoder.Field("equals", func(encoder *jx.Encoder) {
				encodeWebhookFilterEqualsValue(ctx, encoder, value.Equals)
			})
		}

		if !value.In.IsNull() && !value.In.IsUnknown() {
			encoder.Field("in", func(encoder *jx.Encoder) {
				encodeWebhookFilterInValue(ctx, encoder, value.In)
			})
		}

		if !value.Regexp.IsNull() && !value.Regexp.IsUnknown() {
			encoder.Field("regexp", func(encoder *jx.Encoder) {
				encodeWebhookFilterRegexpValue(ctx, encoder, value.Regexp)
			})
		}
	})
}

func encodeWebhookFilterEqualsValue(_ context.Context, encoder *jx.Encoder, value webhookfilter.WebhookFilterEqualsValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(value.Doc.ValueString()) })
		encoder.Field("value", func(encoder *jx.Encoder) { encoder.Str(value.Value.ValueString()) })
	})
}

func encodeWebhookFilterInValue(ctx context.Context, encoder *jx.Encoder, value webhookfilter.WebhookFilterInValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(value.Doc.ValueString()) })
		encoder.Field("values", func(encoder *jx.Encoder) {
			encoder.Arr(func(e *jx.Encoder) {
				values := make([]string, len(value.Values.Elements()))
				value.Values.ElementsAs(ctx, &values, false)

				for _, v := range values {
					e.Str(v)
				}
			})
		})
	})
}

func encodeWebhookFilterRegexpValue(_ context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterRegexpValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(m.Doc.ValueString()) })
		encoder.Field("pattern", func(encoder *jx.Encoder) { encoder.Str(m.Pattern.ValueString()) })
	})
}
