package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		Validators: []validator.List{
			listvalidator.NoNullValues(),
		},
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
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}

func (v WebhookFilterNotValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
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
			Validators: exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:   true,
			Validators: exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
	}
}

func exactlyOneWebhookFilterOperator(names ...string) []validator.Object {
	expressions := make([]path.Expression, 0, len(names))
	for _, name := range names {
		expressions = append(expressions, path.MatchRelative().AtParent().AtName(name))
	}

	return []validator.Object{objectvalidator.ExactlyOneOf(expressions...)}
}

func (v WebhookHeaderValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Optional: true,
		},
		"value_wo": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			WriteOnly: true,
		},
		"secret": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
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
