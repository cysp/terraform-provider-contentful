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

type WebhookFilterRegexpType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterRegexpType{}

func (t WebhookFilterRegexpType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterRegexpType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterRegexpType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterRegexpValue{}
}

func (t WebhookFilterRegexpType) String() string {
	return "WebhookFilterRegexpType"
}

//nolint:ireturn
func (t WebhookFilterRegexpType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, WebhookFilterRegexpValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t WebhookFilterRegexpType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterRegexpValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterNotValue from Terraform: %w", err)
	}

	v, diags := NewWebhookFilterRegexpValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterRegexpType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookFilterRegexpValueNull(), nil
	case value.IsUnknown():
		return NewWebhookFilterRegexpValueUnknown(), nil
	}

	return NewWebhookFilterRegexpValueKnownFromAttributes(ctx, value.Attributes())
}
