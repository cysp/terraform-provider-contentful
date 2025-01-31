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
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookFilterRegexpType) TerraformAttributeTypes(_ context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":     tftypes.String,
		"pattern": tftypes.String,
	}
}

//nolint:ireturn
func (t WebhookFilterRegexpType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), value.Type())
	}

	if value.IsNull() {
		return NewWebhookFilterRegexpValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookFilterRegexpValueUnknown(), nil
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

	v, diags := NewWebhookFilterRegexpValueKnownFromAttributes(ctx, attributes)

	return v, util.ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterRegexpType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterRegexpValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterRegexpValueUnknown(), diags
	}

	return NewWebhookFilterRegexpValueKnownFromAttributes(ctx, value.Attributes())
}
