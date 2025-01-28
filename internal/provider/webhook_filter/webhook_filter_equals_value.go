package webhookfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterEqualsValue struct {
	Doc   string `tfsdk:"doc"`
	Value string `tfsdk:"value"`
}

func (m WebhookFilterEqualsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterEqualsType{
		ObjectType: m.ObjectType(ctx),
	}
}

var _ basetypes.ObjectTypable = WebhookFilterEqualsType{}

var _ basetypes.ObjectValuable = WebhookFilterEqualsValue{}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (v WebhookFilterEqualsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: v.TerraformAttributeTypes(ctx)}

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = basetypes.NewStringValue(v.Doc).ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["doc"] = val

	val, err = basetypes.NewStringValue(v.Value).ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["value"] = val

	if err := tftypes.ValidateValue(objectType, vals); err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(objectType, vals), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterEqualsValue) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterEqualsValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
			Required: true,
		},
	}
}

func (m WebhookFilterEqualsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterEqualsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":   basetypes.StringType{},
		"value": basetypes.StringType{},
	}
}

func (m WebhookFilterEqualsValue) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterEqualsValue) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":   tftypes.String,
		"value": tftypes.String,
	}
}
