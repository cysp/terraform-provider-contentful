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

func (t WebhookFilterInType) TerraformAttributeTypes(_ context.Context) map[string]tftypes.Type {
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

	attributes, err := util.AttributesFromTerraform(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterNotValue from Terraform: %w", err)
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

	return NewWebhookFilterInValueKnownFromAttributes(ctx, value.Attributes())
}
