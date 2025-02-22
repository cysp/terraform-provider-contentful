package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookTransformationType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookTransformationType{}

func (t WebhookTransformationType) Equal(o attr.Type) bool {
	other, ok := o.(WebhookTransformationType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t WebhookTransformationType) ValueType(_ context.Context) attr.Value {
	return WebhookTransformationValue{}
}

func (t WebhookTransformationType) String() string {
	return "WebhookTransformationType"
}

//nolint:ireturn
func (t WebhookTransformationType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t WebhookTransformationType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"method":                 types.String{}.Type(ctx).TerraformType(ctx),
		"content_type":           types.String{}.Type(ctx).TerraformType(ctx),
		"include_content_length": types.Bool{}.Type(ctx).TerraformType(ctx),
		"body":                   jsontypes.Normalized{}.Type(ctx).TerraformType(ctx),
	}
}

//nolint:ireturn
func (t WebhookTransformationType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewWebhookTransformationValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewWebhookTransformationValueNull(), nil
	}

	if !value.IsKnown() {
		return NewWebhookTransformationValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookTransformationValue from Terraform: %w", err)
	}

	v, diags := NewWebhookTransformationValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t WebhookTransformationType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewWebhookTransformationValueNull(), nil
	case value.IsUnknown():
		return NewWebhookTransformationValueUnknown(), nil
	}

	return NewWebhookTransformationValueKnownFromAttributes(ctx, value.Attributes())
}
