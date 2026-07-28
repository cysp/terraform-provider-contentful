package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToEnvironmentLinks(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("test")

	tests := map[string]struct {
		value               TypedList[types.String]
		expected            []cm.EnvironmentLink
		expectedDiagnostics []string
	}{
		"null": {
			value:    NewTypedListNull[types.String](),
			expected: nil,
		},
		"unknown": {
			value:    NewTypedListUnknown[types.String](),
			expected: nil,
		},
		"unknown element": {
			value: NewTypedList([]types.String{
				types.StringUnknown(),
			}),
			expected:            nil,
			expectedDiagnostics: []string{"test[0]"},
		},
		"known and unknown elements": {
			value: NewTypedList([]types.String{
				types.StringValue("a"),
				types.StringUnknown(),
				types.StringValue("c"),
			}),
			expected:            nil,
			expectedDiagnostics: []string{"test[1]"},
		},
		"null element": {
			value: NewTypedList([]types.String{
				types.StringNull(),
			}),
			expected:            nil,
			expectedDiagnostics: []string{"test[0]"},
		},
		"empty": {
			value:    NewTypedList([]types.String{}),
			expected: []cm.EnvironmentLink{},
		},
		"known elements": {
			value: NewTypedList([]types.String{
				types.StringValue("env1"),
				types.StringValue("env2"),
			}),
			expected: []cm.EnvironmentLink{
				cm.NewEnvironmentLink("env1"),
				cm.NewEnvironmentLink("env2"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToEnvironmentLinks(ctx, path, test.value)

			assert.Equal(t, test.expected, result)

			if len(test.expectedDiagnostics) > 0 {
				require.True(t, diags.HasError())
				assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
			} else {
				assert.Empty(t, diags)
			}
		})
	}
}

func TestDeliveryAPIKeyEnvironmentRequestEncoding(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		environments TypedList[types.String]
		expectedJSON string
	}{
		"null is omitted": {
			environments: NewTypedListNull[types.String](),
			expectedJSON: `{"name":"key","description":null}`,
		},
		"unknown is omitted": {
			environments: NewTypedListUnknown[types.String](),
			expectedJSON: `{"name":"key","description":null}`,
		},
		"known empty is explicit": {
			environments: NewTypedList([]types.String{}),
			expectedJSON: `{"name":"key","description":null,"environments":[]}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := DeliveryAPIKeyModel{
				Name:         types.StringValue("key"),
				Description:  types.StringNull(),
				Environments: test.environments,
			}

			request, diags := model.ToAPIKeyRequestFields(t.Context())
			require.False(t, diags.HasError(), diags.Errors())

			actualJSON, err := request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedJSON, string(actualJSON))
		})
	}
}

func TestNewEnvironmentIDsListValueFromEnvironmentLinks(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("test")

	environmentLinks := []cm.EnvironmentLink{
		cm.NewEnvironmentLink("env1"),
		cm.NewEnvironmentLink("env2"),
	}

	expected := NewTypedList([]types.String{
		types.StringValue("env1"),
		types.StringValue("env2"),
	})

	result, diags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path, environmentLinks)
	assert.Empty(t, diags)
	assert.Equal(t, expected, result)
}
