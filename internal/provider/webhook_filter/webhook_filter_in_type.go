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

type WebhookFilterInType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterInType{}

func (m WebhookFilterInType) Equal(other attr.Type) bool {
	//xxx
	return true
}

func (m WebhookFilterInType) ValueType(ctx context.Context) attr.Value {
	return WebhookFilterInValue{}
}

func (m WebhookFilterInType) String() string {
	return "WebhookFilterInType"
}

func (m WebhookFilterInType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: m.TerraformAttributeTypes(ctx),
	}
}

func (m WebhookFilterInType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":    tftypes.String,
		"values": tftypes.List{ElementType: tftypes.String},
	}
}

func (m WebhookFilterInType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.Type().Equal(m.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", m.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterInValueUnknown(), nil
	}

	v := NewWebhookFilterInValueKnown()

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

func (m WebhookFilterInType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterInValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterInValueUnknown(), diags
	}

	v := NewWebhookFilterInValueKnown()

	attributes := value.Attributes()

	doc, diags := types.StringType.ValueFromString(ctx, attributes["doc"].(types.String))
	if diags.HasError() {
		return v, diags
	}
	v.Doc, diags = doc.ToStringValue(ctx)
	if diags.HasError() {
		return v, diags
	}

	valueValues, diags := types.ListType{ElemType: types.StringType}.ValueFromList(ctx, attributes["values"].(types.List))
	if diags.HasError() {
		return v, diags
	}
	v.Values, diags = valueValues.ToListValue(ctx)
	if diags.HasError() {
		return v, diags
	}

	return v, diags
}
