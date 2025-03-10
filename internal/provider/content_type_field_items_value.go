package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldItemsValue struct {
	ItemsType   basetypes.StringValue `tfsdk:"type"`
	LinkType    basetypes.StringValue `tfsdk:"link_type"`
	Validations basetypes.ListValue   `tfsdk:"validations"`
	state       attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeFieldItemsValue{}

func NewContentTypeFieldItemsValueUnknown() ContentTypeFieldItemsValue {
	return ContentTypeFieldItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeFieldItemsValueNull() ContentTypeFieldItemsValue {
	return ContentTypeFieldItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldItemsValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) ContentTypeFieldItemsValue {
	value, diags := NewContentTypeFieldItemsValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func NewContentTypeFieldItemsValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldItemsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldItemsValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeFieldItemsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
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
func (v ContentTypeFieldItemsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldItemsType{
		v.ObjectType(ctx),
	}
}

//nolint:ireturn
func (v ContentTypeFieldItemsValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldItemsType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v ContentTypeFieldItemsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v ContentTypeFieldItemsValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"type":        basetypes.StringType{},
		"link_type":   basetypes.StringType{},
		"validations": basetypes.ListType{ElemType: jsontypes.NormalizedType{}},
	}
}

func (v ContentTypeFieldItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ContentTypeFieldItemsValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return compareTFSDKAttributesEqual(v, other)
	}

	return true
}

func (v ContentTypeFieldItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldItemsValue) String() string {
	return "ContentTypeFieldItemsValue"
}

//nolint:dupl
func (v ContentTypeFieldItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := ContentTypeFieldItemsType{}.TerraformType(ctx)

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

	var itemsErr error
	val["type"], itemsErr = v.ItemsType.ToTerraformValue(ctx)

	var linkTypeErr error
	val["link_type"], linkTypeErr = v.LinkType.ToTerraformValue(ctx)

	var validationsErr error
	val["validations"], validationsErr = v.Validations.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(itemsErr, linkTypeErr, validationsErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v ContentTypeFieldItemsValue) ToObjectValueMust(ctx context.Context) basetypes.ObjectValue {
	value, diags := v.ToObjectValue(ctx)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func (v ContentTypeFieldItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"type":        v.ItemsType,
		"link_type":   v.LinkType,
		"validations": v.Validations,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
