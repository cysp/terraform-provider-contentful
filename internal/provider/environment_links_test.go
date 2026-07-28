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

	tests := map[string]struct {
		environmentLinks []cm.EnvironmentLink
		expected         TypedList[types.String]
	}{
		"nil": {
			environmentLinks: nil,
			expected:         NewTypedListNull[types.String](),
		},
		"empty": {
			environmentLinks: []cm.EnvironmentLink{},
			expected:         NewTypedList([]types.String{}),
		},
		"known elements": {
			environmentLinks: []cm.EnvironmentLink{
				cm.NewEnvironmentLink("env1"),
				cm.NewEnvironmentLink("env2"),
			},
			expected: NewTypedList([]types.String{
				types.StringValue("env1"),
				types.StringValue("env2"),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path, test.environmentLinks)
			assert.Empty(t, diags)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestEnvironmentLinksRequestResponseRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	valuePath := path.Root("test")

	for name, value := range map[string]TypedList[types.String]{
		"null":  NewTypedListNull[types.String](),
		"empty": NewTypedList([]types.String{}),
		"known elements": NewTypedList([]types.String{
			types.StringValue("env1"),
			types.StringValue("env2"),
		}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			environmentLinks, requestDiags := ToEnvironmentLinks(ctx, valuePath, value)
			require.False(t, requestDiags.HasError(), requestDiags.Errors())

			result, responseDiags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, valuePath, environmentLinks)
			require.False(t, responseDiags.HasError(), responseDiags.Errors())

			assert.Equal(t, value, result)
		})
	}
}

func TestEnvironmentLinksJSONRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	valuePath := path.Root("test")

	for name, test := range map[string]struct {
		value             TypedList[types.String]
		environmentsField string
	}{
		"null": {
			value:             NewTypedListNull[types.String](),
			environmentsField: "",
		},
		"empty": {
			value:             NewTypedList([]types.String{}),
			environmentsField: `"environments":[]`,
		},
		"known elements": {
			value: NewTypedList([]types.String{
				types.StringValue("env1"),
				types.StringValue("env2"),
			}),
			environmentsField: `"environments":[`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			environmentLinks, requestDiags := ToEnvironmentLinks(ctx, valuePath, test.value)
			require.False(t, requestDiags.HasError(), requestDiags.Errors())

			apiKey := cm.ApiKey{
				Sys:          cm.NewAPIKeySys("space", "key"),
				Name:         "key",
				Environments: environmentLinks,
				AccessToken:  "token",
			}

			encoded, err := apiKey.MarshalJSON()
			require.NoError(t, err)

			if test.environmentsField == "" {
				assert.NotContains(t, string(encoded), `"environments"`)
			} else {
				assert.Contains(t, string(encoded), test.environmentsField)
			}

			var decoded cm.ApiKey
			require.NoError(t, decoded.UnmarshalJSON(encoded))

			result, responseDiags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, valuePath, decoded.Environments)
			require.False(t, responseDiags.HasError(), responseDiags.Errors())

			assert.Equal(t, test.value, result)
		})
	}
}
