package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: webhookfilter.WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: webhookfilter.WebhookFilterValue{}.CustomType(ctx),
		},
		Optional: optional,
	}
}

func ToWebhookDefinitionFilter(ctx context.Context, m webhookfilter.WebhookFilterValue) (contentfulManagement.WebhookDefinitionFilter, diag.Diagnostics) {
	// en := jx.Encoder{}
	// return en.Encode(m)

	// b := []byte(`{"foo":"bar"}`)

	encoder := jx.Encoder{}

	encodeWebhookFilterValue(ctx, &encoder, m)

	return contentfulManagement.WebhookDefinitionFilter(encoder.Bytes()), nil
}

func encodeWebhookFilterValue(ctx context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		if !m.Not.IsNull() && !m.Not.IsUnknown() {
			encoder.Field("not", func(encoder *jx.Encoder) {
				encodeWebhookFilterNotValue(ctx, encoder, m.Not)
			})
		}
		if !m.Equals.IsNull() && !m.Equals.IsUnknown() {
			encoder.Field("equals", func(encoder *jx.Encoder) {
				encodeWebhookFilterEqualsValue(ctx, encoder, m.Equals)
			})
		}
		if !m.In.IsNull() && !m.In.IsUnknown() {
			encoder.Field("in", func(encoder *jx.Encoder) {
				encodeWebhookFilterInValue(ctx, encoder, m.In)
			})
		}
		if !m.Regexp.IsNull() && !m.Regexp.IsUnknown() {
			encoder.Field("regexp", func(encoder *jx.Encoder) {
				encodeWebhookFilterRegexpValue(ctx, encoder, m.Regexp)
			})
		}
	})
}

func encodeWebhookFilterNotValue(ctx context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterNotValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("equals", func(encoder *jx.Encoder) {
			encodeWebhookFilterEqualsValue(ctx, encoder, m.Equals)
		})
	})
}

func encodeWebhookFilterEqualsValue(ctx context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterEqualsValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(m.Doc.ValueString()) })
		encoder.Field("value", func(encoder *jx.Encoder) { encoder.Str(m.Value.ValueString()) })
	})
}

func encodeWebhookFilterInValue(ctx context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterInValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(m.Doc.ValueString()) })
		encoder.Field("values", func(encoder *jx.Encoder) {
			encoder.Arr(func(e *jx.Encoder) {
				values := make([]string, len(m.Values.Elements()))
				m.Values.ElementsAs(ctx, &values, false)
				for _, v := range values {
					e.Str(v)
				}
			})
		})
	})
}

func encodeWebhookFilterRegexpValue(ctx context.Context, encoder *jx.Encoder, m webhookfilter.WebhookFilterRegexpValue) {
	encoder.Obj(func(encoder *jx.Encoder) {
		encoder.Field("doc", func(encoder *jx.Encoder) { encoder.Str(m.Doc.ValueString()) })
		encoder.Field("pattern", func(encoder *jx.Encoder) { encoder.Str(m.Pattern.ValueString()) })
	})
}
