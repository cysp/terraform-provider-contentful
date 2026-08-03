package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamModelToTeamData(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		model               provider.TeamModel
		expected            cm.TeamData
		expectedDiagnostics []string
	}{
		"known": {
			model: provider.TeamModel{
				Name:        types.StringValue("Editors"),
				Description: types.StringValue("Content editors"),
			},
			expected: cm.TeamData{
				Name:        "Editors",
				Description: cm.NewNilString("Content editors"),
			},
		},
		"known empty description": {
			model: provider.TeamModel{
				Name:        types.StringValue("Editors"),
				Description: types.StringValue(""),
			},
			expected: cm.TeamData{
				Name:        "Editors",
				Description: cm.NewNilString(""),
			},
		},
		"null description": {
			model: provider.TeamModel{
				Name:        types.StringValue("Editors"),
				Description: types.StringNull(),
			},
			expected: cm.TeamData{
				Name:        "Editors",
				Description: cm.NewNilStringNull(),
			},
		},
		"unknown description": {
			model: provider.TeamModel{
				Name:        types.StringValue("Editors"),
				Description: types.StringUnknown(),
			},
			expectedDiagnostics: []string{"description"},
		},
		"unknown name": {
			model: provider.TeamModel{
				Name:        types.StringUnknown(),
				Description: types.StringValue("Content editors"),
			},
			expectedDiagnostics: []string{"name"},
		},
		"null name and unknown description": {
			model: provider.TeamModel{
				Name:        types.StringNull(),
				Description: types.StringUnknown(),
			},
			expectedDiagnostics: []string{"name", "description"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.model.ToTeamData(t.Context(), path.Empty())

			if len(test.expectedDiagnostics) == 0 {
				require.False(t, diags.HasError(), diags.Errors())
				assert.Equal(t, test.expected, actual)

				return
			}

			assert.Equal(t, cm.TeamData{}, actual)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
		})
	}
}
