package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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

func (m WebhookFilterModel) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: m.TerraformAttributeTypes(ctx),
	}
}

func (m WebhookFilterModel) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"not":    WebhookFilterInteriorModel{}.TerraformType(ctx),
		"equals": WebhookFilterEqualityConstraintModel{}.TerraformType(ctx),
		"in":     WebhookFilterInConstraintModel{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpConstraintModel{}.TerraformType(ctx),
	}
}

func (v WebhookFilterModel) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: v.TerraformAttributeTypes(ctx)}

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = v.Not.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["not"] = val

	val, err = v.Equals.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["equals"] = val

	val, err = v.In.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["in"] = val

	val, err = v.Regexp.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["regexp"] = val

	if err := tftypes.ValidateValue(objectType, vals); err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(objectType, vals), nil
	// case attr.ValueStateNull:
	// 	return tftypes.NewValue(objectType, nil), nil
	// case attr.ValueStateUnknown:
	// 	return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	// default:
	// 	panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	// }
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

func (m WebhookFilterModel) ToWebhookDefinitionFilter(context.Context) (contentfulManagement.WebhookDefinitionFilter, diag.Diagnostics) {
	// en := jx.Encoder{}
	// return en.Encode(m)

	b := []byte(`{"foo":"bar"}`)

	return contentfulManagement.WebhookDefinitionFilter(b), nil
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
func (v WebhookFilterInteriorModel) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: v.TerraformAttributeTypes(ctx)}

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = v.Equals.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["equals"] = val

	val, err = v.In.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["in"] = val

	val, err = v.Regexp.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["regexp"] = val

	if err := tftypes.ValidateValue(objectType, vals); err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(objectType, vals), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterInteriorModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterInteriorModel) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterInteriorModel) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"equals": WebhookFilterEqualityConstraintModel{}.TerraformType(ctx),
		"in":     WebhookFilterInConstraintModel{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpConstraintModel{}.TerraformType(ctx),
	}
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
func (v WebhookFilterEqualityConstraintModel) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: v.TerraformAttributeTypes(ctx)}

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = basetypes.NewStringValue(v.Doc).ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["doc"] = val

	val, err = basetypes.NewStringValue(v.Value).ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["value"] = val

	if err := tftypes.ValidateValue(objectType, vals); err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(objectType, vals), nil
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

func (m WebhookFilterEqualityConstraintModel) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterEqualityConstraintModel) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":   tftypes.String,
		"value": tftypes.String,
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
func (m WebhookFilterInConstraintModel) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}

	return tftypes.NewValue(objectType, nil), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterInConstraintModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterInConstraintModel) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterInConstraintModel) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":    tftypes.String,
		"values": tftypes.List{ElementType: tftypes.String},
	}
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
func (m WebhookFilterRegexpConstraintModel) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}

	return tftypes.NewValue(objectType, nil), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpConstraintModel) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterRegexpConstraintModel) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterRegexpConstraintModel) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":     tftypes.String,
		"pattern": tftypes.String,
	}
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
