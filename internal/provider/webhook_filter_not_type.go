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

type WebhookFilterNotType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterNotType{}

func (t WebhookFilterNotType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterNotType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t WebhookFilterNotType) String() string {
	return "WebhookFilterNotType"
}

//nolint:ireturn
func (t WebhookFilterNotType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, WebhookFilterNotValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t WebhookFilterNotType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterNotValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterNotValue from Terraform: %w", err)
	}

	v, diags := NewWebhookFilterNotValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterNotType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookFilterNotValueNull(), nil
	case value.IsUnknown():
		return NewWebhookFilterNotValueUnknown(), nil
	}

	return NewWebhookFilterNotValueKnownFromAttributes(ctx, value.Attributes())
}
