package webhookfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterRegexpValue struct {
	Doc     string `tfsdk:"doc"`
	Pattern string `tfsdk:"pattern"`
}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) IsNull() bool {
	panic("unimplemented")
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) IsUnknown() bool {
	panic("unimplemented")
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}

	return tftypes.NewValue(objectType, nil), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterRegexpValue) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterRegexpValue) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterRegexpValue) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":     tftypes.String,
		"pattern": tftypes.String,
	}
}

var _ basetypes.ObjectValuable = WebhookFilterRegexpValue{}

func (m WebhookFilterRegexpValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"pattern": schema.StringAttribute{
			Required: true,
		},
	}
}

func (m WebhookFilterRegexpValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterRegexpValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":     basetypes.StringType{},
		"pattern": basetypes.StringType{},
	}
}
