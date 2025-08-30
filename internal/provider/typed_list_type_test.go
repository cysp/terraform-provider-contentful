package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedListTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	listType := TypedList[types.String]{}.Type(ctx)

	testcases := map[string]struct {
		tfval       tftypes.Value
		expectError bool
		expected    TypedList[types.String]
	}{
		"null type": {
			tfval:    tftypes.NewValue(nil, nil),
			expected: NewTypedListNull[types.String](ctx),
		},
		"unknown": {
			tfval:    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
			expected: NewTypedListUnknown[types.String](ctx),
		},
		"null": {
			tfval:    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			expected: NewTypedListNull[types.String](ctx),
		},
		"incorrect type": {
			tfval:       tftypes.NewValue(tftypes.String, "string"),
			expectError: true,
		},
		"incorrect element type": {
			tfval:       tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, []tftypes.Value{}),
			expectError: true,
		},
		"empty": {
			tfval:    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
			expected: DiagsNoErrorsMust(NewTypedList(ctx, []types.String{})),
		},
		"with elements": {
			tfval: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "value1"),
				tftypes.NewValue(tftypes.String, "value2"),
			}),
			expected: DiagsNoErrorsMust(NewTypedList(ctx, []types.String{
				types.StringValue("value1"),
				types.StringValue("value2"),
			})),
		},
		"with interior unknown element": {
			tfval: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "value1"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "value2"),
			}),
			expected: DiagsNoErrorsMust(NewTypedList(ctx, []types.String{
				types.StringValue("value1"),
				types.StringUnknown(),
				types.StringValue("value2"),
			})),
		},
		"with interior null element": {
			tfval: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "value1"),
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "value2"),
			}),
			expected: DiagsNoErrorsMust(NewTypedList(ctx, []types.String{
				types.StringValue("value1"),
				types.StringNull(),
				types.StringValue("value2"),
			})),
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			actual, err := listType.ValueFromTerraform(ctx, testcase.tfval)

			if testcase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			actualTypedList, ok := actual.(TypedList[types.String])
			require.True(t, ok, "Expected TypedList[types.String] but got %T", actual)

			assert.Equal(t, testcase.expected, actualTypedList)
		})
	}
}
