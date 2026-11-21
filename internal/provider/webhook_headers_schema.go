package provider

import (
	"context"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.MapNestedAttribute{
		Description: "HTTP headers sent by the webhook, keyed by header name. When omitted on Create, Terraform adopts the headers returned by Contentful. When omitted on Update, Terraform preserves known prior headers. Set to `{}` to remove all headers.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookHeaderValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookHeaderValue]().CustomType(ctx),
		},
		CustomType: TypedMap[TypedObject[WebhookHeaderValue]]{}.CustomType(ctx),
		Optional:   optional,
		Computed:   true,
		Validators: []validator.Map{
			mapvalidator.NoNullValues(),
		},
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
			validateWebhookHeaderValues{},
		},
	}
}

// validateWebhookHeaderValues checks the effective map after UseStateForUnknown.
type validateWebhookHeaderValues struct{}

func (validateWebhookHeaderValues) Description(context.Context) string {
	return "A header value may be omitted only to retain an existing secret with the exact same name."
}

func (m validateWebhookHeaderValues) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (validateWebhookHeaderValues) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// An unchanged map can contain values inherited through omission or ignore_changes.
	// Request conversion checks those values if an update actually sends a PUT.
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.Equal(req.StateValue) {
		return
	}

	var planned, prior map[string]TypedObject[WebhookHeaderValue]
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planned, false)...)
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &prior, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, key := range slices.Sorted(maps.Keys(planned)) {
		if header, known := planned[key].GetValue(); known {
			resp.Diagnostics.Append(header.validateValueOmission(req.Path.AtMapKey(key), prior[key])...)
		}
	}
}
