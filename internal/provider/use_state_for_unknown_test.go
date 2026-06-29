package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestUseStateForUnknownPlanModifyList(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	stateValue := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("existing"),
	})
	resp := &planmodifier.ListResponse{}

	UseStateForUnknown().PlanModifyList(ctx, planmodifier.ListRequest{
		State:       priorNonNullState(),
		ConfigValue: types.ListNull(types.StringType),
		PlanValue:   types.ListUnknown(types.StringType),
		StateValue:  stateValue,
	}, resp)

	assert.True(t, resp.PlanValue.Equal(stateValue))
}

func TestUseStateForUnknownPlanModifyObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	attributeTypes := map[string]attr.Type{"name": types.StringType}
	stateValue := types.ObjectValueMust(attributeTypes, map[string]attr.Value{
		"name": types.StringValue("existing"),
	})
	resp := &planmodifier.ObjectResponse{}

	UseStateForUnknown().PlanModifyObject(ctx, planmodifier.ObjectRequest{
		State:       priorNonNullState(),
		ConfigValue: types.ObjectNull(attributeTypes),
		PlanValue:   types.ObjectUnknown(attributeTypes),
		StateValue:  stateValue,
	}, resp)

	assert.True(t, resp.PlanValue.Equal(stateValue))
}

func priorNonNullState() tfsdk.State {
	return tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
	}
}
