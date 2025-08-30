package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeFieldValueToTerraformValueRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input ContentTypeFieldValue
		check func(t *testing.T, v tftypes.Value)
	}{
		"unknown": {
			input: NewContentTypeFieldValueUnknown(),
			check: func(t *testing.T, v tftypes.Value) {
				t.Helper()

				assert.False(t, v.IsKnown())
				assert.False(t, v.IsNull())
			},
		},
		"null": {
			input: NewContentTypeFieldValueNull(),
			check: func(t *testing.T, v tftypes.Value) {
				t.Helper()

				assert.True(t, v.IsKnown())
				assert.True(t, v.IsNull())
			},
		},
		"known": {
			input: DiagsNoErrorsMust(NewContentTypeFieldValueKnownFromAttributes(ctx, map[string]attr.Value{
				"type":        types.StringValue("Link"),
				"link_type":   types.StringValue("Entry"),
				"validations": DiagsNoErrorsMust(NewTypedList(ctx, []jsontypes.Normalized{})),
				"id":          types.StringValue("id"),
				"name":        types.StringValue("name"),
				"items": DiagsNoErrorsMust(NewContentTypeFieldItemsValueKnownFromAttributes(ctx, map[string]attr.Value{
					"type":        types.StringValue("Link"),
					"link_type":   types.StringValue("Entry"),
					"validations": DiagsNoErrorsMust(NewTypedList(ctx, []jsontypes.Normalized{})),
				})),
				"default_value":     jsontypes.NewNormalizedNull(),
				"localized":         types.BoolValue(true),
				"disabled":          types.BoolValue(false),
				"omitted":           types.BoolValue(false),
				"required":          types.BoolValue(true),
				"allowed_resources": NewTypedListNull[ContentTypeFieldAllowedResourceItemValue](ctx),
			})),
			check: func(t *testing.T, v tftypes.Value) {
				t.Helper()

				assert.True(t, v.IsKnown())
				assert.False(t, v.IsNull())
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := testcase.input.ToTerraformValue(ctx)
			require.NoError(t, err)

			testcase.check(t, actual)

			roundtrip, err := testcase.input.Type(ctx).ValueFromTerraform(ctx, actual)
			require.NoError(t, err)

			assert.True(t, testcase.input.Equal(roundtrip))
		})
	}
}
