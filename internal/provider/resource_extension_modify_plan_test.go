package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionModifyPlanReconcilesSourceOwnership(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		configSrc       types.String
		configSrcdoc    types.String
		stateSrc        types.String
		stateSrcdoc     types.String
		stateExists     bool
		plannedSrc      types.String
		plannedSrcdoc   types.String
		wantSrc         types.String
		wantSrcdoc      types.String
		wantDiagnostics bool
		wantPath        string
	}{
		"create requires a source": {
			configSrc:       types.StringNull(),
			configSrcdoc:    types.StringNull(),
			plannedSrc:      types.StringUnknown(),
			plannedSrcdoc:   types.StringUnknown(),
			wantDiagnostics: true,
			wantPath:        "extension.src",
		},
		"create uses configured src": {
			configSrc:     types.StringValue("https://example.com/extension.js"),
			configSrcdoc:  types.StringNull(),
			plannedSrc:    types.StringValue("https://example.com/extension.js"),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringValue("https://example.com/extension.js"),
			wantSrcdoc:    types.StringNull(),
		},
		"planning defers an unknown configured srcdoc": {
			configSrc:     types.StringNull(),
			configSrcdoc:  types.StringUnknown(),
			plannedSrc:    types.StringUnknown(),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringNull(),
			wantSrcdoc:    types.StringUnknown(),
		},
		"omitted configuration preserves state src": {
			configSrc:     types.StringNull(),
			configSrcdoc:  types.StringNull(),
			stateSrc:      types.StringValue("https://state.example/extension.js"),
			stateSrcdoc:   types.StringNull(),
			stateExists:   true,
			plannedSrc:    types.StringUnknown(),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringValue("https://state.example/extension.js"),
			wantSrcdoc:    types.StringNull(),
		},
		"omitted configuration preserves empty state srcdoc": {
			configSrc:     types.StringNull(),
			configSrcdoc:  types.StringNull(),
			stateSrc:      types.StringNull(),
			stateSrcdoc:   types.StringValue(""),
			stateExists:   true,
			plannedSrc:    types.StringUnknown(),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringNull(),
			wantSrcdoc:    types.StringValue(""),
		},
		"legacy empty sibling preserves valid state src": {
			configSrc:     types.StringNull(),
			configSrcdoc:  types.StringNull(),
			stateSrc:      types.StringValue("https://state.example/extension.js"),
			stateSrcdoc:   types.StringValue(""),
			stateExists:   true,
			plannedSrc:    types.StringUnknown(),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringValue("https://state.example/extension.js"),
			wantSrcdoc:    types.StringNull(),
		},
		"two genuine state sources are rejected": {
			configSrc:       types.StringNull(),
			configSrcdoc:    types.StringNull(),
			stateSrc:        types.StringValue("https://state.example/extension.js"),
			stateSrcdoc:     types.StringValue("<p>state</p>"),
			stateExists:     true,
			plannedSrc:      types.StringUnknown(),
			plannedSrcdoc:   types.StringUnknown(),
			wantDiagnostics: true,
			wantPath:        "extension.srcdoc",
		},
		"config switches srcdoc to src": {
			configSrc:     types.StringValue("https://new.example/extension.js"),
			configSrcdoc:  types.StringNull(),
			stateSrc:      types.StringNull(),
			stateSrcdoc:   types.StringValue("<p>old</p>"),
			stateExists:   true,
			plannedSrc:    types.StringValue("https://new.example/extension.js"),
			plannedSrcdoc: types.StringUnknown(),
			wantSrc:       types.StringValue("https://new.example/extension.js"),
			wantSrcdoc:    types.StringNull(),
		},
		"config switches src to srcdoc": {
			configSrc:     types.StringNull(),
			configSrcdoc:  types.StringValue("<p>new</p>"),
			stateSrc:      types.StringValue("https://old.example/extension.js"),
			stateSrcdoc:   types.StringNull(),
			stateExists:   true,
			plannedSrc:    types.StringUnknown(),
			plannedSrcdoc: types.StringValue("<p>new</p>"),
			wantSrc:       types.StringNull(),
			wantSrcdoc:    types.StringValue("<p>new</p>"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			resourceSchema := ExtensionResourceSchema(ctx)
			configModel := extensionPlanTestModel(test.configSrc, test.configSrcdoc)
			plannedModel := extensionPlanTestModel(test.plannedSrc, test.plannedSrcdoc)

			configPlan := tfsdk.Plan{Schema: resourceSchema}
			require.False(t, configPlan.Set(ctx, configModel).HasError())

			plan := tfsdk.Plan{Schema: resourceSchema}
			require.False(t, plan.Set(ctx, plannedModel).HasError())

			state := tfsdk.State{Schema: resourceSchema}
			if test.stateExists {
				require.False(t, state.Set(ctx, extensionPlanTestModel(test.stateSrc, test.stateSrcdoc)).HasError())
			} else {
				state.Raw = tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil)
			}

			request := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Raw: configPlan.Raw, Schema: resourceSchema},
				State:  state,
				Plan:   plan,
			}
			response := resource.ModifyPlanResponse{Plan: plan}

			resourceWithModifyPlan, ok := NewExtensionResource().(resource.ResourceWithModifyPlan)
			require.True(t, ok)

			resourceWithModifyPlan.ModifyPlan(ctx, request, &response)

			assert.Equal(t, test.wantDiagnostics, response.Diagnostics.HasError(), response.Diagnostics)

			if test.wantDiagnostics {
				assert.Equal(t, []string{test.wantPath}, attributeDiagnosticPaths(t, response.Diagnostics))

				return
			}

			var actualSrc, actualSrcdoc types.String
			require.False(t, response.Plan.GetAttribute(ctx, path.Root("extension").AtName("src"), &actualSrc).HasError())
			require.False(t, response.Plan.GetAttribute(ctx, path.Root("extension").AtName("srcdoc"), &actualSrcdoc).HasError())
			assert.Equal(t, test.wantSrc, actualSrc)
			assert.Equal(t, test.wantSrcdoc, actualSrcdoc)
		})
	}
}

func extensionPlanTestModel(src, srcdoc types.String) ExtensionModel {
	return ExtensionModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringNull()},
		ExtensionIdentityModel: ExtensionIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			ExtensionID:   types.StringValue("extension"),
		},
		Extension: &ExtensionModelExtension{
			Name:       types.StringValue("Extension"),
			Src:        src,
			SrcDoc:     srcdoc,
			FieldTypes: []AppDefinitionLocationFieldTypesItem{},
			Sidebar:    types.BoolValue(false),
		},
		Parameters: jsontypes.NewNormalizedNull(),
		Timeouts:   TimeoutsNull(),
	}
}
