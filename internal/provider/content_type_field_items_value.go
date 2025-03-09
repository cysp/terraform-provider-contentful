package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	diags := diag.Diagnostics{}

	itemsValue, itemsOk := attributes["items"].(basetypes.StringValue)
	if !itemsOk {
		diags.AddAttributeError(path.Root("items"), "invalid data", fmt.Sprintf("expected object of type types.Object, got %T", attributes["items"]))
	}

	linkTypeValue, linkTypeOk := attributes["link_type"].(basetypes.StringValue)
	if !linkTypeOk {
		diags.AddAttributeError(path.Root("link_type"), "invalid data", fmt.Sprintf("expected object of type types.Object, got %T", attributes["link_type"]))
	}

	validationsValue, validationsOk := attributes["validations"].(basetypes.ListValue)
	if !validationsOk {
		diags.AddAttributeError(path.Root("validations"), "invalid data", fmt.Sprintf("expected object of type types.Object, got %T", attributes["validations"]))
	}

	return ItemsValue{
		ItemsType:   itemsValue,
		LinkType:    linkTypeValue,
		Validations: validationsValue,
		state:       attr.ValueStateKnown,
	}, diags
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

func (v ItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ItemsValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return v.ItemsType.Equal(other.ItemsType)
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
	tft := FieldsType{}.TerraformType(ctx)

	switch v.state {
	case attr.ValueStateKnown:
		break
	case attr.ValueStateNull:
		return tftypes.NewValue(tft, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}

	//nolint:gomnd,mnd
	val := make(map[string]tftypes.Value, 3)

	var linkTypeErr error
	val["link_type"], linkTypeErr = v.LinkType.ToTerraformValue(ctx)

	var itemsErr error
	val["items"], itemsErr = v.ItemsType.ToTerraformValue(ctx)

	var validationsErr error
	val["validations"], validationsErr = v.Validations.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(itemsErr, linkTypeErr, validationsErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"link_type":   v.LinkType,
		"type":        v.ItemsType,
		"validations": v.Validations,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
