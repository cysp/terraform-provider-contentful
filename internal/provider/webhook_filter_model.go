package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookFilterModel{}.AttributesSchema(ctx),
		},
		Optional: optional,
	}
}

type WebhookFilterType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterType{}

type WebhookFilterModel struct {
	Not    WebhookFilterInteriorModel           `tfsdk:"not"`
	Equals WebhookFilterEqualityConstraintModel `tfsdk:"equals"`
	In     WebhookFilterInConstraintModel       `tfsdk:"in"`
	Regexp WebhookFilterRegexpConstraintModel   `tfsdk:"regexp"`
}

var _ basetypes.ObjectValuable = WebhookFilterModel{}

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

func (m WebhookFilterModel) IsNull() bool {
	return false
}

func (m WebhookFilterModel) IsUnknown() bool {
	return false
}

func (m WebhookFilterModel) String() string {
	return ""
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

func (m WebhookFilterModel) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

func (m WebhookFilterModel) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := m.AttributeTypes(ctx)

	if m.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if m.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"not":    m.Not,
			"equals": m.Equals,
			"in":     m.In,
			"regexp": m.Regexp,
		})

	return objVal, diags
}

func (m WebhookFilterModel) Equal(other attr.Value) bool {
	return false
}

type WebhookFilterInteriorModel struct {
	Equals WebhookFilterEqualityConstraintModel `tfsdk:"equals"`
	In     WebhookFilterInConstraintModel       `tfsdk:"in"`
	Regexp WebhookFilterRegexpConstraintModel   `tfsdk:"regexp"`
}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

var _ basetypes.ObjectValuable = WebhookFilterInteriorModel{}

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
	Doc   string `tfsdk:"doc"`
	Value string `tfsdk:"value"`
}

var _ basetypes.ObjectValuable = WebhookFilterEqualityConstraintModel{}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterEqualityConstraintModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterEqualityConstraintModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
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
		"doc":   basetypes.StringType{},
		"value": basetypes.StringType{},
	}
}

type WebhookFilterInConstraintModel struct {
	Doc    string   `tfsdk:"doc"`
	Values []string `tfsdk:"values"`
}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

var _ basetypes.ObjectValuable = WebhookFilterInConstraintModel{}

func (m WebhookFilterInConstraintModel) AttributesSchema(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"values": schema.ListAttribute{
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
		"values": basetypes.ListType{
			ElemType: basetypes.StringType{},
		},
	}
}

type WebhookFilterRegexpConstraintModel struct {
	Doc     string `tfsdk:"doc"`
	Pattern string `tfsdk:"pattern"`
}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

var _ basetypes.ObjectValuable = WebhookFilterRegexpConstraintModel{}

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
