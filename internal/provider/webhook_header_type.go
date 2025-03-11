package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookHeaderType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookHeaderType{}

func (t WebhookHeaderType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookHeaderType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookHeaderType) ValueType(_ context.Context) attr.Value {
	return WebhookHeaderValue{}
}

func (t WebhookHeaderType) String() string {
	return "WebhookHeaderType"
}

//nolint:ireturn
func (t WebhookHeaderType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, WebhookHeaderValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t WebhookHeaderType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookHeaderValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookHeaderValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookHeaderValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookHeaderValue from Terraform: %w", err)
	}

	v, diags := NewWebhookHeaderValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookHeaderType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookHeaderValueNull(), nil
	case value.IsUnknown():
		return NewWebhookHeaderValueUnknown(), nil
	}

	return NewWebhookHeaderValueKnownFromAttributes(ctx, value.Attributes())
}
