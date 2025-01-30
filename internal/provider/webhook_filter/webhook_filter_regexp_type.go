package webhookfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterRegexpType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterRegexpType{}

func (m WebhookFilterRegexpType) Equal(other attr.Type) bool {
	//xxx
	return true
}

func (m WebhookFilterRegexpType) ValueType(ctx context.Context) attr.Value {
	return WebhookFilterRegexpValue{}
}

func (m WebhookFilterRegexpType) String() string {
	return "WebhookFilterRegexpType"
}

func (m WebhookFilterRegexpType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: m.TerraformAttributeTypes(ctx),
	}
}

func (m WebhookFilterRegexpType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":     tftypes.String,
		"pattern": tftypes.String,
	}
}

func (m WebhookFilterRegexpType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.Type().Equal(m.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", m.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterRegexpValueUnknown(), nil
	}

	v := NewWebhookFilterRegexpValueKnown()

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

func (m WebhookFilterRegexpType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterRegexpValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterRegexpValueUnknown(), diags
	}

	v := NewWebhookFilterRegexpValueKnown()

	attributes := value.Attributes()

	doc, diags := types.StringType.ValueFromString(ctx, attributes["doc"].(types.String))
	if diags.HasError() {
		return v, diags
	}
	v.Doc, diags = doc.ToStringValue(ctx)
	if diags.HasError() {
		return v, diags
	}

	valuePattern, diags := types.StringType.ValueFromString(ctx, attributes["pattern"].(types.String))
	if diags.HasError() {
		return v, diags
	}
	v.Pattern, diags = valuePattern.ToStringValue(ctx)
	if diags.HasError() {
		return v, diags
	}

	return v, diags
}
