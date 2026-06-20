package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTeamResourceModelFromResponseDescription(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		description cm.NilString
		expected    types.String
	}{
		"non-empty": {
			description: cm.NewNilString("Team description"),
			expected:    types.StringValue("Team description"),
		},
		"empty": {
			description: cm.NewNilString(""),
			expected:    types.StringValue(""),
		},
		"null": {
			description: cm.NewNilStringNull(),
			expected:    types.StringNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, diags := provider.NewTeamResourceModelFromResponse(t.Context(), cm.Team{
				Sys:         cm.NewTeamSys("organization-id", "team-id"),
				Name:        "Test Team",
				Description: test.description,
			})
			require.False(t, diags.HasError())

			assert.Equal(t, test.expected, model.Description)
		})
	}
}
