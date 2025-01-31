package webhookfilter

import (
	"context"
	"fmt"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:revive
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

func (t WebhookFilterEqualsType) TerraformAttributeTypes(context.Context) map[string]tftypes.Type {
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
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterEqualsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterEqualsValueUnknown(), nil
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

	v, diags := NewWebhookFilterEqualsValueKnownFromAttributes(ctx, attributes)

	return v, util.ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterEqualsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterEqualsValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterEqualsValueUnknown(), diags
	}

	attributes := value.Attributes()

	return NewWebhookFilterEqualsValueKnownFromAttributes(ctx, attributes)
}
