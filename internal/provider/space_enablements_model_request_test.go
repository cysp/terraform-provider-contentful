package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpaceEnablementsRequestRequiresKnownEqualCoupledValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		crossSpaceLinks types.Bool
		spaceTemplates  types.Bool
		wantPath        string
	}{
		"missing cross space links": {
			crossSpaceLinks: types.BoolNull(),
			spaceTemplates:  types.BoolValue(true),
			wantPath:        "cross_space_links",
		},
		"response-owned unknown space templates": {
			crossSpaceLinks: types.BoolValue(true),
			spaceTemplates:  types.BoolUnknown(),
			wantPath:        "space_templates",
		},
		"unequal values": {
			crossSpaceLinks: types.BoolValue(true),
			spaceTemplates:  types.BoolValue(false),
			wantPath:        "space_templates",
		},
		"known false values": {
			crossSpaceLinks: types.BoolValue(false),
			spaceTemplates:  types.BoolValue(false),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := SpaceEnablementsModel{
				CrossSpaceLinks:   test.crossSpaceLinks,
				SpaceTemplates:    test.spaceTemplates,
				StudioExperiences: types.BoolUnknown(),
				SuggestConcepts:   types.BoolUnknown(),
			}
			config := SpaceEnablementsModel{
				CrossSpaceLinks:   types.BoolNull(),
				SpaceTemplates:    types.BoolNull(),
				StudioExperiences: types.BoolNull(),
				SuggestConcepts:   types.BoolNull(),
			}

			actual, diags := model.ToSpaceEnablementData(t.Context(), config)

			if test.wantPath != "" {
				assert.Equal(t, cm.SpaceEnablementData{}, actual)
				require.True(t, diags.HasError())
				assert.Contains(t, attributeDiagnosticPaths(t, diags), test.wantPath)

				return
			}

			require.False(t, diags.HasError(), diags.Errors())

			crossSpaceLinks, ok := actual.CrossSpaceLinks.Get()
			require.True(t, ok)
			assert.False(t, crossSpaceLinks.Enabled)

			spaceTemplates, ok := actual.SpaceTemplates.Get()
			require.True(t, ok)
			assert.False(t, spaceTemplates.Enabled)
			assert.False(t, actual.StudioExperiences.IsSet())
			assert.False(t, actual.SuggestConcepts.IsSet())
		})
	}
}

func TestSpaceEnablementsRequestIndependentOptionalComputedOwnership(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		field      string
		configured types.Bool
		planned    types.Bool
		wantPath   string
	}{
		"response-owned known studio experiences is sent": {
			field:      "studio_experiences",
			configured: types.BoolNull(),
			planned:    types.BoolValue(true),
		},
		"configuration-owned unknown studio experiences fails": {
			field:      "studio_experiences",
			configured: types.BoolValue(true),
			planned:    types.BoolUnknown(),
			wantPath:   "studio_experiences",
		},
		"response-owned known suggest concepts is sent": {
			field:      "suggest_concepts",
			configured: types.BoolNull(),
			planned:    types.BoolValue(true),
		},
		"configuration-owned unknown suggest concepts fails": {
			field:      "suggest_concepts",
			configured: types.BoolValue(true),
			planned:    types.BoolUnknown(),
			wantPath:   "suggest_concepts",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := SpaceEnablementsModel{
				CrossSpaceLinks:   types.BoolValue(false),
				SpaceTemplates:    types.BoolValue(false),
				StudioExperiences: types.BoolNull(),
				SuggestConcepts:   types.BoolNull(),
			}
			config := SpaceEnablementsModel{
				CrossSpaceLinks:   types.BoolNull(),
				SpaceTemplates:    types.BoolNull(),
				StudioExperiences: types.BoolNull(),
				SuggestConcepts:   types.BoolNull(),
			}

			switch test.field {
			case "studio_experiences":
				model.StudioExperiences = test.planned
				config.StudioExperiences = test.configured
			case "suggest_concepts":
				model.SuggestConcepts = test.planned
				config.SuggestConcepts = test.configured
			}

			actual, diags := model.ToSpaceEnablementData(t.Context(), config)

			if test.wantPath != "" {
				assert.Equal(t, cm.SpaceEnablementData{}, actual)
				require.True(t, diags.HasError())
				assert.Equal(t, []string{test.wantPath}, attributeDiagnosticPaths(t, diags))

				return
			}

			require.False(t, diags.HasError(), diags.Errors())

			var field cm.OptSpaceEnablementField
			switch test.field {
			case "studio_experiences":
				field = actual.StudioExperiences
			case "suggest_concepts":
				field = actual.SuggestConcepts
			}

			value, ok := field.Get()
			require.True(t, ok)
			assert.True(t, value.Enabled)
		})
	}
}
