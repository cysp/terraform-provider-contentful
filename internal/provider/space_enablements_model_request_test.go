package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpaceEnablementsRequestSendsKnownUnequalValues(t *testing.T) {
	t.Parallel()

	model := SpaceEnablementsModel{
		CrossSpaceLinks:   types.BoolValue(true),
		SpaceTemplates:    types.BoolValue(false),
		StudioExperiences: types.BoolNull(),
		SuggestConcepts:   types.BoolNull(),
	}
	config := SpaceEnablementsModel{
		CrossSpaceLinks:   types.BoolValue(true),
		SpaceTemplates:    types.BoolValue(false),
		StudioExperiences: types.BoolNull(),
		SuggestConcepts:   types.BoolNull(),
	}

	actual, diags := model.ToSpaceEnablementData(t.Context(), config)
	require.False(t, diags.HasError(), diags.Errors())

	crossSpaceLinks, ok := actual.CrossSpaceLinks.Get()
	require.True(t, ok)
	assert.True(t, crossSpaceLinks.Enabled)

	spaceTemplates, ok := actual.SpaceTemplates.Get()
	require.True(t, ok)
	assert.False(t, spaceTemplates.Enabled)
}

func TestSpaceEnablementsRequestIndependentOptionalComputedOwnership(t *testing.T) {
	t.Parallel()

	fields := []string{"cross_space_links", "space_templates", "studio_experiences", "suggest_concepts"}
	tests := map[string]struct {
		configured  types.Bool
		planned     types.Bool
		wantSet     bool
		wantEnabled bool
		wantError   bool
	}{
		"response-owned null is omitted": {
			configured: types.BoolNull(),
			planned:    types.BoolNull(),
		},
		"response-owned unknown is omitted": {
			configured: types.BoolNull(),
			planned:    types.BoolUnknown(),
		},
		"response-owned known value is sent": {
			configured:  types.BoolNull(),
			planned:     types.BoolValue(true),
			wantSet:     true,
			wantEnabled: true,
		},
		"configuration-owned known false is sent": {
			configured: types.BoolValue(false),
			planned:    types.BoolValue(false),
			wantSet:    true,
		},
		"configuration-owned unknown fails": {
			configured: types.BoolValue(true),
			planned:    types.BoolUnknown(),
			wantError:  true,
		},
	}

	for _, fieldName := range fields {
		for testName, test := range tests {
			t.Run(fieldName+"/"+testName, func(t *testing.T) {
				t.Parallel()

				model := SpaceEnablementsModel{
					CrossSpaceLinks:   types.BoolNull(),
					SpaceTemplates:    types.BoolNull(),
					StudioExperiences: types.BoolNull(),
					SuggestConcepts:   types.BoolNull(),
				}
				config := SpaceEnablementsModel{
					CrossSpaceLinks:   types.BoolNull(),
					SpaceTemplates:    types.BoolNull(),
					StudioExperiences: types.BoolNull(),
					SuggestConcepts:   types.BoolNull(),
				}

				switch fieldName {
				case "cross_space_links":
					model.CrossSpaceLinks = test.planned
					config.CrossSpaceLinks = test.configured
				case "space_templates":
					model.SpaceTemplates = test.planned
					config.SpaceTemplates = test.configured
				case "studio_experiences":
					model.StudioExperiences = test.planned
					config.StudioExperiences = test.configured
				case "suggest_concepts":
					model.SuggestConcepts = test.planned
					config.SuggestConcepts = test.configured
				}

				actual, diags := model.ToSpaceEnablementData(t.Context(), config)

				if test.wantError {
					assert.Equal(t, cm.SpaceEnablementData{}, actual)
					require.True(t, diags.HasError())
					assert.Equal(t, []string{fieldName}, attributeDiagnosticPaths(t, diags))

					return
				}

				require.False(t, diags.HasError(), diags.Errors())

				var field cm.OptSpaceEnablementField

				switch fieldName {
				case "cross_space_links":
					field = actual.CrossSpaceLinks
				case "space_templates":
					field = actual.SpaceTemplates
				case "studio_experiences":
					field = actual.StudioExperiences
				case "suggest_concepts":
					field = actual.SuggestConcepts
				}

				value, ok := field.Get()

				assert.Equal(t, test.wantSet, ok)

				if ok {
					assert.Equal(t, test.wantEnabled, value.Enabled)
				}
			})
		}
	}
}
