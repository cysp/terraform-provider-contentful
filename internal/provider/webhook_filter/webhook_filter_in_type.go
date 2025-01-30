package webhookfilter

import (
	"context"
	"fmt"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:revive
type WebhookFilterInType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterInType{}

func (t WebhookFilterInType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterInType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterInType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterInValue{}
}

func (t WebhookFilterInType) String() string {
	return "WebhookFilterInType"
}

//nolint:ireturn
func (t WebhookFilterInType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookFilterInType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":    tftypes.String,
		"values": tftypes.List{ElementType: tftypes.String},
	}
}

//nolint:ireturn
func (t WebhookFilterInType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterInValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterInValueUnknown(), nil
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

	v, diags := NewWebhookFilterInValueKnownFromAttributes(ctx, attributes)

	return v, util.ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterInType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
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
