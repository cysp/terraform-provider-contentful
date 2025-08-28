package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToEnvironmentLinks(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("test")

	tests := map[string]struct {
		value         TypedList[types.String]
		expected      []cm.EnvironmentLink
		expectedDiags bool
	}{
		"unknown": {
			value:    NewTypedListUnknown[types.String](),
			expected: nil,
		},
		"unknown element": {
			value: NewTypedList([]types.String{
				types.StringUnknown(),
			}),
			expected: []cm.EnvironmentLink{
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "",
					},
				},
			},
			expectedDiags: true,
		},
		"known and unknown elements": {
			value: NewTypedList([]types.String{
				types.StringValue("a"),
				types.StringUnknown(),
				types.StringValue("c"),
			}),
			expected: []cm.EnvironmentLink{
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "a",
					},
				},
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "",
					},
				},
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "c",
					},
				},
			},
			expectedDiags: true,
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
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "env1",
					},
				},
				{
					Sys: cm.EnvironmentLinkSys{
						Type:     cm.EnvironmentLinkSysTypeLink,
						LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
						ID:       "env2",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToEnvironmentLinks(ctx, path, test.value)

			assert.Equal(t, test.expected, result)

			if test.expectedDiags {
				assert.NotEmpty(t, diags)
			} else {
				assert.Empty(t, diags)
			}
		})
	}
}

func TestNewEnvironmentIDsListValueFromEnvironmentLinks(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("test")

	environmentLinks := []cm.EnvironmentLink{
		{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       "env1",
			},
		},
		{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       "env2",
			},
		},
	}

	expected := NewTypedList([]types.String{
		types.StringValue("env1"),
		types.StringValue("env2"),
	})

	result, diags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path, environmentLinks)
	assert.Empty(t, diags)
	assert.Equal(t, expected, result)
}
