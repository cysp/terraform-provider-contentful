package webhookfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterType{}

func (t WebhookFilterType) Equal(other attr.Type) bool {
	return true
}

func (t WebhookFilterType) ValueType(ctx context.Context) attr.Value {
	return WebhookFilterValue{}
}

func (t WebhookFilterType) String() string {
	return "WebhookFilterType"
}

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

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	v, diags := NewWebhookFilterValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to create WebhookFilterValue from attributes: %s", diags)
	}

	return v, nil
}

func (t WebhookFilterType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewWebhookFilterValueNull(), diags
	}

	if in.IsUnknown() {
		return NewWebhookFilterValueUnknown(), diags
	}

	value := NewWebhookFilterValueKnown()

	attributes := in.Attributes()

	valueNot, diags := WebhookFilterNotType{}.ValueFromObject(ctx, attributes["not"].(basetypes.ObjectValue))
	if diags.HasError() {
		return value, diags
	}
	value.Not = valueNot.(WebhookFilterNotValue)

	valueEquals, diags := WebhookFilterEqualsType{}.ValueFromObject(ctx, attributes["equals"].(basetypes.ObjectValue))
	if diags.HasError() {
		return value, diags
	}
	value.Equals = valueEquals.(WebhookFilterEqualsValue)

	valueIn, diags := WebhookFilterInType{}.ValueFromObject(ctx, attributes["in"].(basetypes.ObjectValue))
	if diags.HasError() {
		return value, diags
	}
	value.In = valueIn.(WebhookFilterInValue)

	valueRegexp, diags := WebhookFilterRegexpType{}.ValueFromObject(ctx, attributes["regexp"].(basetypes.ObjectValue))
	if diags.HasError() {
		return value, diags
	}
	value.Regexp = valueRegexp.(WebhookFilterRegexpValue)

	return value, diags
}
