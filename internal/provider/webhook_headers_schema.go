package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.MapNestedAttribute{
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
			UseStateForUnknown(),
		},
	}
}
