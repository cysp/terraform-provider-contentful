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

type WebhookFilterInType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterInType{}

func (t WebhookFilterInType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterInType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterInType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterInValue{}
}

func (t WebhookFilterInType) String() string {
	return "WebhookFilterInType"
}

//nolint:ireturn
func (t WebhookFilterInType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, WebhookFilterInValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t WebhookFilterInType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterInValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterNotValue from Terraform: %w", err)
	}

	v, diags := NewWebhookFilterInValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterInType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookFilterInValueNull(), nil
	case value.IsUnknown():
		return NewWebhookFilterInValueUnknown(), nil
	}

	return NewWebhookFilterInValueKnownFromAttributes(ctx, value.Attributes())
}
