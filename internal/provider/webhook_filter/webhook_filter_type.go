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
type WebhookFilterType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterType{}

func (t WebhookFilterType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookFilterType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookFilterType) ValueType(_ context.Context) attr.Value {
	return WebhookFilterValue{}
}

func (t WebhookFilterType) String() string {
	return "WebhookFilterType"
}

//nolint:ireturn
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

//nolint:ireturn
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

	attributes, err := util.AttributesFromTerraform(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookFilterValue from Terraform: %w", err)
	}

	v, diags := NewWebhookFilterValueKnownFromAttributes(ctx, attributes)
	return v, util.ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookFilterType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() {
		return NewWebhookFilterValueNull(), diags
	}

	if value.IsUnknown() {
		return NewWebhookFilterValueUnknown(), diags
	}

	return NewWebhookFilterValueKnownFromAttributes(ctx, value.Attributes())
}
