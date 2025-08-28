package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterValue]().CustomType(ctx),
		},
		CustomType: NewTypedListNull[TypedObject[WebhookFilterValue]]().CustomType(ctx),
		Optional:   optional,
	}
}

func (v WebhookFilterEqualsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
			Required: true,
		},
	}
}

func (v WebhookFilterInValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"values": schema.ListAttribute{
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
		},
	}
}

func (v WebhookFilterNotValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:   true,
		},
	}
}

func (v WebhookFilterRegexpValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"pattern": schema.StringAttribute{
			Required: true,
		},
	}
}

func (v WebhookFilterValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"not": schema.SingleNestedAttribute{
			Attributes: WebhookFilterNotValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterNotValue]().CustomType(ctx),
			Optional:   true,
		},
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:   true,
		},
	}
}

func (v WebhookHeaderValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Required: true,
		},
		"secret": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				UseStateForUnknown(),
			},
		},
	}
}

func (v WebhookTransformationValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"method": schema.StringAttribute{
			Optional: true,
		},
		"content_type": schema.StringAttribute{
			Optional: true,
		},
		"include_content_length": schema.BoolAttribute{
			Optional: true,
		},
		"body": schema.StringAttribute{
			Optional:   true,
			CustomType: jsontypes.NormalizedType{},
		},
	}
}
