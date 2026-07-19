package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToWebhookDefinitionHeaders(ctx context.Context, valuePath path.Path, model TypedMap[TypedObject[WebhookHeaderValue]]) (cm.WebhookDefinitionHeaders, diag.Diagnostics) {
	// An omitted Optional+Computed map is unknown on initial create. Omit it
	// from the request; UseStateForUnknown supplies known prior state on update.
	if model.IsNull() || model.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}
	headers, headerDiags := ConvertKnownObjectMapElements(
		ctx,
		valuePath,
		model.Elements(),
		func(ctx context.Context, headerPath path.Path, key string, value WebhookHeaderValue) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
			return value.ToWebhookDefinitionHeader(ctx, headerPath, key)
		},
	)
	diags.Append(headerDiags...)

	return headers, diags
}
