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

func (m WebhookFilterType) Equal(other attr.Type) bool {
	panic("unimplemented")
}

func (m WebhookFilterType) ValueType(ctx context.Context) attr.Value {
	return WebhookFilterValue{}
}

func (m WebhookFilterType) String() string {
	return "WebhookFilterType"
}

func (m WebhookFilterType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: m.TerraformAttributeTypes(ctx),
	}
}

func (m WebhookFilterType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"not":    WebhookFilterNotType{}.TerraformType(ctx),
		"equals": WebhookFilterEqualsType{}.TerraformType(ctx),
		"in":     WebhookFilterInType{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpType{}.TerraformType(ctx),
	}
}

func (m WebhookFilterType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterValueNull(), nil
	}

	if !value.Type().Equal(m.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", m.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterValueUnknown(), nil
	}

	v := NewWebhookFilterValueKnown()

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

	return v, nil
}

func (m WebhookFilterType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterValueUnknown(), diags
	}

	v := NewWebhookFilterValueKnown()

	attributes := value.Attributes()

	v.Not, diags = WebhookFilterNotValue{}.ValueFromObject(ctx, attributes["not"])
	if diags.HasError() {
		return v, diags
	}

	v.Equals, diags = WebhookFilterEqualsValue{}.ValueFromObject(ctx, attributes["equals"])
	if diags.HasError() {
		return v, diags
	}

	v.In, diags = WebhookFilterInValue{}.ValueFromObject(ctx, attributes["in"])
	if diags.HasError() {
		return v, diags
	}

	v.Regexp, diags = WebhookFilterRegexpValue{}.ValueFromObject(ctx, attributes["regexp"])
	if diags.HasError() {
		return v, diags
	}

	return v, diags
}
