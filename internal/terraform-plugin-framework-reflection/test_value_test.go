package terraformpluginframeworkreflection_test

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testType struct {
	basetypes.ObjectType
}
type testValue struct {
	state attr.ValueState
}

var (
	_ attr.Type               = testType{}
	_ basetypes.ObjectTypable = testType{}
	_ attr.Value              = testValue{}
)

func (v testValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v testValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v testValue) String() string {
	return "TestValue"
}

func (v testValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

//nolint:ireturn
func (v testValue) Type(_ context.Context) attr.Type {
	return testType{}
}

func (v testValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[testValue](v, o)
}
