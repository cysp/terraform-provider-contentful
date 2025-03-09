package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ItemsValue struct {
	LinkType    basetypes.StringValue `tfsdk:"link_type"`
	ItemsType   basetypes.StringValue `tfsdk:"type"`
	Validations basetypes.ListValue   `tfsdk:"validations"`
	state       attr.ValueState
}

var _ basetypes.ObjectValuable = ItemsValue{}

func NewItemsValueUnknown() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewItemsValueNull() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewItemsValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (ItemsValue, diag.Diagnostics) {
}

func (v ItemsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required: true,
		},
		"link_type": schema.StringAttribute{
			Optional: true,
		},
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
		},
	}
}

//nolint:ireturn
func (v ItemsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ItemsType{
		v.ObjectType(ctx),
	}
}

//nolint:ireturn
func (v ItemsValue) Type(ctx context.Context) attr.Type {
	return ItemsType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v ItemsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v ItemsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"link_type":   basetypes.StringType{},
		"type":        basetypes.StringType{},
		"validations": basetypes.ListType{ElemType: types.StringType},
	}
}

func (v ItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ItemsValue) String() string {
	return "ItemsValue"
}

func (v ItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
}

func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var validationsVal basetypes.ListValue
	switch {
	case v.Validations.IsUnknown():
		validationsVal = types.ListUnknown(types.StringType)
	case v.Validations.IsNull():
		validationsVal = types.ListNull(types.StringType)
	default:
		var d diag.Diagnostics
		validationsVal, d = types.ListValue(types.StringType, v.Validations.Elements())
		diags.Append(d...)
	}

	if diags.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"link_type": basetypes.StringType{},
			"type":      basetypes.StringType{},
			"validations": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"link_type": basetypes.StringType{},
		"type":      basetypes.StringType{},
		"validations": basetypes.ListType{
			ElemType: types.StringType,
		},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"link_type":   v.LinkType,
			"type":        v.ItemsType,
			"validations": validationsVal,
		})

	return objVal, diags
}

func (v ItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ItemsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.LinkType.Equal(other.LinkType) {
		return false
	}

	if !v.ItemsType.Equal(other.ItemsType) {
		return false
	}

	if !v.Validations.Equal(other.Validations) {
		return false
	}

	return true
}
