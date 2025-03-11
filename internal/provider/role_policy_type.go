//nolint:dupl
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type RolePolicyType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = RolePolicyType{}

func (t RolePolicyType) Equal(o attr.Type) bool {
	other, ok := o.(RolePolicyType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t RolePolicyType) ValueType(_ context.Context) attr.Value {
	return RolePolicyValue{}
}

func (t RolePolicyType) String() string {
	return "RolePolicyType"
}

//nolint:ireturn
func (t RolePolicyType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, RolePolicyValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t RolePolicyType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewRolePolicyValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewRolePolicyValueNull(), nil
	}

	if !value.IsKnown() {
		return NewRolePolicyValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create RolePolicyValue from Terraform: %w", err)
	}

	v, diags := NewRolePolicyValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t RolePolicyType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewRolePolicyValueNull(), nil
	case value.IsUnknown():
		return NewRolePolicyValueUnknown(), nil
	}

	return NewRolePolicyValueKnownFromAttributes(ctx, value.Attributes())
}
