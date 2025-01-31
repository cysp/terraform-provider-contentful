package webhookfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:revive
type WebhookFilterType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterType{}

func (t WebhookFilterType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterValue{}
}

func (t WebhookFilterType) String() string {
	return "WebhookFilterType"
}

//nolint:ireturn
func (t WebhookFilterType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookFilterType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"not":    WebhookFilterNotType{}.TerraformType(ctx),
		"equals": WebhookFilterEqualsType{}.TerraformType(ctx),
		"in":     WebhookFilterInType{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpType{}.TerraformType(ctx),
	}
}

//nolint:ireturn
func (t WebhookFilterType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterValueUnknown(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := value.As(&val)
	if err != nil {
		return nil, err
	}

	for key, tfattrval := range val {
		attrval, err := t.AttrTypes[key].ValueFromTerraform(ctx, tfattrval)
		if err != nil {
			return nil, err
		}

		attributes[key] = attrval
	}

	v, diags := NewWebhookFilterValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to create WebhookFilterValue from attributes: %s", diags)
	}

	return v, nil
}

//nolint:ireturn
func (t WebhookFilterType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterValueUnknown(), diags
	}

	return NewWebhookFilterValueKnownFromAttributes(ctx, value.Attributes())
}
