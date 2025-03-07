package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/assert"
)

func TestModelTypeEqual(t *testing.T) {
	t.Parallel()

	type InequalType struct {
		attr.Type
	}

	types := []attr.Type{
		provider.ControlsType{},
		provider.EditorLayoutType{},
		provider.FieldsType{},
		provider.GroupControlsType{},
		provider.ItemsType{},
		provider.SidebarType{},
		provider.RolePolicyType{},
		provider.WebhookFilterEqualsType{},
		provider.WebhookFilterInType{},
		provider.WebhookFilterNotType{},
		provider.WebhookFilterRegexpType{},
		provider.WebhookFilterType{},
		provider.WebhookHeaderType{},
		provider.WebhookTransformationType{},
	}

	for aIndex, aType := range types {
		t.Run(aType.String(), func(t *testing.T) {
			t.Parallel()

			t.Run(aType.String(), func(t *testing.T) {
				t.Parallel()

				assert.True(t, aType.Equal(aType)) //nolint:gocritic
			})

			t.Run("inequal", func(t *testing.T) {
				t.Parallel()

				assert.False(t, aType.Equal(InequalType{aType}))
			})

			for bIndex, bType := range types {
				if aIndex == bIndex {
					continue
				}

				t.Run(bType.String(), func(t *testing.T) {
					t.Parallel()

					assert.False(t, aType.Equal(InequalType{aType}))
				})
			}
		})
	}
}
