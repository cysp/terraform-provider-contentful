package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelTypeEqual(t *testing.T) {
	t.Parallel()

	type InequalType struct {
		attr.Type
	}

	types := []attr.Type{
		provider.ContentTypeFieldItemsType{},
		provider.ContentTypeFieldType{},
		provider.EditorInterfaceControlType{},
		provider.EditorInterfaceEditorLayoutType{},
		provider.EditorInterfaceGroupControlType{},
		provider.EditorInterfaceSidebarType{},
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

func TestModelTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	types := []attr.Type{
		provider.ContentTypeFieldItemsType{},
		provider.ContentTypeFieldType{},
		provider.EditorInterfaceControlType{},
		provider.EditorInterfaceEditorLayoutType{},
		provider.EditorInterfaceGroupControlType{},
		provider.EditorInterfaceSidebarType{},
		provider.RolePolicyType{},
		provider.WebhookFilterEqualsType{},
		provider.WebhookFilterInType{},
		provider.WebhookFilterNotType{},
		provider.WebhookFilterRegexpType{},
		provider.WebhookFilterType{},
		provider.WebhookHeaderType{},
		provider.WebhookTransformationType{},
	}

	tfvalniltype := tftypes.NewValue(nil, nil)

	for _, typ := range types {
		tftyp := typ.TerraformType(ctx)

		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			tfvalunknown := tftypes.NewValue(tftyp, tftypes.UnknownValue)
			valueUnknown, err := typ.ValueFromTerraform(ctx, tfvalunknown)
			require.NoError(t, err)
			assert.True(t, valueUnknown.IsUnknown())
		})

		t.Run("nil", func(t *testing.T) {
			t.Parallel()

			valueNil, err := typ.ValueFromTerraform(ctx, tfvalniltype)
			require.NoError(t, err)
			assert.True(t, valueNil.IsNull())
		})

		t.Run("null", func(t *testing.T) {
			t.Parallel()

			tfvalnull := tftypes.NewValue(tftyp, nil)
			valueNull, err := typ.ValueFromTerraform(ctx, tfvalnull)
			require.NoError(t, err)
			assert.True(t, valueNull.IsNull())
		})
	}
}
