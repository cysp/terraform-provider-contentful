package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
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

	b := []byte(`{"foo":"bar"}`)

	return contentfulManagement.WebhookDefinitionFilter(b), nil
}
