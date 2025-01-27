package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookFilterModel{}.AttributesSchema(ctx),
		},
		Optional: optional,
	}
}

type WebhookFilterModel struct {
	Not    WebhookFilterInteriorModel           `tfsdk:"not"`
	Equals WebhookFilterEqualityConstraintModel `tfsdk:"equals"`
	In     WebhookFilterInConstraintModel       `tfsdk:"in"`
	Regexp WebhookFilterRegexpConstraintModel   `tfsdk:"regexp"`
}

func (m WebhookFilterModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"not": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInteriorModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterInteriorModel{}.ObjectType(ctx),
			Optional:   true,
		},
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualityConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterEqualityConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterInConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterRegexpConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
	}
}

func (m WebhookFilterModel) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterType{
		m.ObjectType(ctx),
	}
}

var _ basetypes.ObjectTypable = WebhookFilterType{}

type WebhookFilterType struct {
	basetypes.ObjectType
}

func (m WebhookFilterModel) Type(ctx context.Context) attr.Type {
	return WebhookFilterType{
		m.ObjectType(ctx),
	}
}

func (m WebhookFilterModel) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"not":    WebhookFilterInteriorModel{}.ObjectType(ctx),
		"equals": WebhookFilterEqualityConstraintModel{}.ObjectType(ctx),
		"in":     WebhookFilterInConstraintModel{}.ObjectType(ctx),
		"regexp": WebhookFilterRegexpConstraintModel{}.ObjectType(ctx),
	}
}

type WebhookFilterInteriorModel struct {
	Equals WebhookFilterEqualityConstraintModel `tfsdk:"equals"`
	In     WebhookFilterInConstraintModel       `tfsdk:"in"`
	Regexp WebhookFilterRegexpConstraintModel   `tfsdk:"regexp"`
}

func (m WebhookFilterInteriorModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualityConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterEqualityConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterInConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpConstraintModel{}.AttributesSchema(ctx),
			CustomType: WebhookFilterRegexpConstraintModel{}.ObjectType(ctx),
			Optional:   true,
		},
	}
}

func (m WebhookFilterInteriorModel) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterInteriorModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"equals": WebhookFilterEqualityConstraintModel{}.ObjectType(ctx),
		"in":     WebhookFilterInConstraintModel{}.ObjectType(ctx),
		"regexp": WebhookFilterRegexpConstraintModel{}.ObjectType(ctx),
	}
}

type WebhookFilterEqualityConstraintModel struct {
	Doc    string `tfsdk:"doc"`
	String string `tfsdk:"string"`
}

func (m WebhookFilterEqualityConstraintModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"string": schema.StringAttribute{
			Required: true,
		},
	}
}

func (m WebhookFilterEqualityConstraintModel) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterEqualityConstraintModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":    basetypes.StringType{},
		"string": basetypes.StringType{},
	}
}

type WebhookFilterInConstraintModel struct {
	Doc     string   `tfsdk:"doc"`
	Strings []string `tfsdk:"string"`
}

func (m WebhookFilterInConstraintModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"strings": schema.ListAttribute{
			ElementType: basetypes.StringType{},
			Required:    true,
		},
	}
}

func (m WebhookFilterInConstraintModel) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterInConstraintModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc": basetypes.StringType{},
		"strings": basetypes.ListType{
			ElemType: basetypes.StringType{},
		},
	}
}

type WebhookFilterRegexpConstraintModel struct {
	Doc     string `tfsdk:"doc"`
	Pattern string `tfsdk:"pattern"`
}

func (m WebhookFilterRegexpConstraintModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"pattern": schema.StringAttribute{
			Required: true,
		},
	}
}

func (m WebhookFilterRegexpConstraintModel) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterRegexpConstraintModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":     basetypes.StringType{},
		"pattern": basetypes.StringType{},
	}
}
