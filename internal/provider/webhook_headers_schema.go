package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.MapNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookHeaderValue{}.SchemaAttributes(ctx),
			CustomType: WebhookHeaderValue{}.CustomType(ctx),
		},
		CustomType: TypedMap[WebhookHeaderValue]{}.CustomType(ctx),
		Optional:   optional,
		Computed:   true,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}
