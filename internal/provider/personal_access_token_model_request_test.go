package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonalAccessTokenModelToRequestScopes(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		scopes              TypedList[types.String]
		expectedScopes      []string
		expectedDiagnostics []string
	}{
		"known": {
			scopes: NewTypedList([]types.String{
				types.StringValue("scope-a"),
				types.StringValue("scope-b"),
			}),
			expectedScopes: []string{"scope-a", "scope-b"},
		},
		"known empty": {
			scopes:         NewTypedList([]types.String{}),
			expectedScopes: []string{},
		},
		"null container": {
			scopes:              NewTypedListNull[types.String](),
			expectedDiagnostics: []string{"scopes"},
		},
		"unknown container": {
			scopes:              NewTypedListUnknown[types.String](),
			expectedDiagnostics: []string{"scopes"},
		},
		"invalid children": {
			scopes: NewTypedList([]types.String{
				types.StringValue("scope-a"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("scope-b"),
			}),
			expectedDiagnostics: []string{"scopes[1]", "scopes[2]"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := PersonalAccessTokenModel{
				Name:   types.StringValue("token"),
				Scopes: test.scopes,
			}

			request, diags := model.ToPersonalAccessTokenRequestData(t.Context())

			if len(test.expectedDiagnostics) == 0 {
				require.False(t, diags.HasError(), diags.Errors())
				require.NotNil(t, request.Scopes)
				assert.Equal(t, test.expectedScopes, request.Scopes)
				assert.Equal(t, "token", request.Name)

				return
			}

			assert.Equal(t, cm.PersonalAccessTokenRequestData{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
		})
	}
}
