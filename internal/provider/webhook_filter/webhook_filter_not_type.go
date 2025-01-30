package webhookfilter

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

func (m WebhookFilterNotType) Equal(other attr.Type) bool {
	//xxx
	return true
}

func (m WebhookFilterNotType) ValueType(ctx context.Context) attr.Value {
	return WebhookFilterNotValue{}
}

func (m WebhookFilterNotType) String() string {
	return "WebhookFilterNotType"
}

func (m WebhookFilterNotType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: m.TerraformAttributeTypes(ctx),
	}
}

func (m WebhookFilterNotType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"equals": WebhookFilterEqualsType{}.TerraformType(ctx),
		"in":     WebhookFilterInType{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpType{}.TerraformType(ctx),
	}
}

func (m WebhookFilterNotType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.Type().Equal(m.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", m.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterNotValueUnknown(), nil
	}

	v := NewWebhookFilterNotValueKnown()

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := value.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := m.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return v, nil
}

func (m WebhookFilterNotType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewWebhookFilterNotValueNull(), diags
	}

	if in.IsUnknown() {
		return NewWebhookFilterNotValueUnknown(), diags
	}

	value := NewWebhookFilterNotValueKnown()

	attributes := in.Attributes()

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
