package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpaceEnablementsModifyPlanValidatesCoupledConfiguration(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		crossSpaceLinks types.Bool
		spaceTemplates  types.Bool
		stateExists     bool
		wantPath        string
	}{
		"create rejects neither configured": {
			crossSpaceLinks: types.BoolNull(),
			spaceTemplates:  types.BoolNull(),
			wantPath:        "cross_space_links",
		},
		"create rejects only cross space links": {
			crossSpaceLinks: types.BoolValue(true),
			spaceTemplates:  types.BoolNull(),
			wantPath:        "space_templates",
		},
		"create rejects only space templates": {
			crossSpaceLinks: types.BoolNull(),
			spaceTemplates:  types.BoolValue(true),
			wantPath:        "cross_space_links",
		},
		"create rejects unequal values": {
			crossSpaceLinks: types.BoolValue(true),
			spaceTemplates:  types.BoolValue(false),
			wantPath:        "space_templates",
		},
		"create accepts false pair": {
			crossSpaceLinks: types.BoolValue(false),
			spaceTemplates:  types.BoolValue(false),
		},
		"planning defers unknown configured pair": {
			crossSpaceLinks: types.BoolUnknown(),
			spaceTemplates:  types.BoolUnknown(),
		},
		"existing response-owned pair is valid": {
			crossSpaceLinks: types.BoolNull(),
			spaceTemplates:  types.BoolNull(),
			stateExists:     true,
		},
		"update rejects configuration combined with state": {
			crossSpaceLinks: types.BoolValue(true),
			spaceTemplates:  types.BoolNull(),
			stateExists:     true,
			wantPath:        "space_templates",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			resourceSchema := SpaceEnablementsResourceSchema(ctx)
			configModel := spaceEnablementsPlanTestModel(test.crossSpaceLinks, test.spaceTemplates)
			configPlan := tfsdk.Plan{Schema: resourceSchema}
			require.False(t, configPlan.Set(ctx, configModel).HasError())

			plan := tfsdk.Plan{Schema: resourceSchema}
			require.False(t, plan.Set(ctx, configModel).HasError())

			state := tfsdk.State{Schema: resourceSchema}
			if test.stateExists {
				require.False(t, state.Set(ctx, spaceEnablementsPlanTestModel(types.BoolValue(true), types.BoolValue(true))).HasError())
			} else {
				state.Raw = tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil)
			}

			request := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Raw: configPlan.Raw, Schema: resourceSchema},
				State:  state,
				Plan:   plan,
			}
			response := resource.ModifyPlanResponse{Plan: plan}

			resourceWithModifyPlan, ok := NewSpaceEnablementsResource().(resource.ResourceWithModifyPlan)
			require.True(t, ok)

			resourceWithModifyPlan.ModifyPlan(ctx, request, &response)

			if test.wantPath == "" {
				assert.False(t, response.Diagnostics.HasError(), response.Diagnostics)

				return
			}

			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, []string{test.wantPath}, attributeDiagnosticPaths(t, response.Diagnostics))
		})
	}
}

func spaceEnablementsPlanTestModel(crossSpaceLinks, spaceTemplates types.Bool) SpaceEnablementsModel {
	return SpaceEnablementsModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringNull()},
		SpaceEnablementsIdentityModel: SpaceEnablementsIdentityModel{
			SpaceID: types.StringValue("space"),
		},
		CrossSpaceLinks:   crossSpaceLinks,
		SpaceTemplates:    spaceTemplates,
		StudioExperiences: types.BoolNull(),
		SuggestConcepts:   types.BoolNull(),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
		})},
	}
}
