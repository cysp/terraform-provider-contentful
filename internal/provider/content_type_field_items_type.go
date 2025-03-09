package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ItemsType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ItemsType{}

func (t ItemsType) Equal(o attr.Type) bool {
	other, ok := o.(ItemsType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t ItemsType) ValueType(_ context.Context) attr.Value {
	return ItemsValue{}
}

func (t ItemsType) String() string {
	return "ItemsType"
}

//nolint:ireturn
func (t ItemsType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t ItemsType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"type":        tftypes.String,
		"link_type":   tftypes.String,
		"validations": tftypes.List{ElementType: jsontypes.NormalizedType{}.TerraformType(ctx)},
	}
}

//nolint:ireturn
func (t ItemsType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewItemsValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewItemsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewItemsValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ItemsValue from Terraform: %w", err)
	}

	v, diags := NewItemsValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ItemsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewItemsValueNull(), nil
	case value.IsUnknown():
		return NewItemsValueUnknown(), nil
	}

	return NewItemsValueKnownFromAttributes(ctx, value.Attributes())
}
