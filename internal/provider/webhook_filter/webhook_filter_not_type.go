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

//nolint:ireturn
func (t WebhookFilterNotType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterNotValue{}
}

func (t WebhookFilterNotType) String() string {
	return "WebhookFilterNotType"
}

//nolint:ireturn
func (t WebhookFilterNotType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookFilterNotType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"equals": WebhookFilterEqualsType{}.TerraformType(ctx),
		"in":     WebhookFilterInType{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpType{}.TerraformType(ctx),
	}
}

//nolint:ireturn
func (t WebhookFilterNotType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterNotValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterNotValueUnknown(), nil
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

	v, diags := NewWebhookFilterNotValueKnownFromAttributes(ctx, attributes)

	return v, util.ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterNotType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterNotValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterNotValueUnknown(), diags
	}

	return NewWebhookFilterNotValueKnownFromAttributes(ctx, value.Attributes())

}
