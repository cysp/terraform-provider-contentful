package terraformpluginframeworkreflection_test

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testInvalidType struct {
	basetypes.ObjectType
}
type testInvalidValue struct {
	A     string `tfsdk:"a"`
	state attr.ValueState
}

var (
	_ attr.Type               = testInvalidType{}
	_ basetypes.ObjectTypable = testInvalidType{}
	_ attr.Value              = testInvalidValue{}
)

//nolint:ireturn
func (t testInvalidType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": types.StringType.TerraformType(ctx),
		},
	}
}

func (v testInvalidValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v testInvalidValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v testInvalidValue) String() string {
	return "TestInvalidValue"
}

func (v testInvalidValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

//nolint:ireturn
func (v testInvalidValue) Type(_ context.Context) attr.Type {
	return testInvalidType{}
}

func (v testInvalidValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[testInvalidValue](v, o)
}
