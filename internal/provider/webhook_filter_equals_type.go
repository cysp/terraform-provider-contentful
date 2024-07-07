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

type WebhookFilterEqualsType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterEqualsType{}

func (t WebhookFilterEqualsType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterEqualsType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterEqualsType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterEqualsValue{}
}

func (t WebhookFilterEqualsType) String() string {
	return "WebhookFilterEqualsType"
}

//nolint:ireturn
func (t WebhookFilterEqualsType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookFilterEqualsType) TerraformAttributeTypes(_ context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":   tftypes.String,
		"value": tftypes.String,
	}
}

//nolint:ireturn
func (t WebhookFilterEqualsType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterEqualsValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookFilterEqualsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterEqualsValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterNotValue from Terraform: %w", err)
	}

	v, diags := NewWebhookFilterEqualsValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterEqualsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookFilterEqualsValueNull(), nil
	case value.IsUnknown():
		return NewWebhookFilterEqualsValueUnknown(), nil
	}

	return NewWebhookFilterEqualsValueKnownFromAttributes(ctx, value.Attributes())
}
