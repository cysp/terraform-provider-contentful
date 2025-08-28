package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webhookFiltersListForTesting(t *testing.T) TypedList[WebhookFilterValue] {
	t.Helper()

	ctx := t.Context()

	filterEquals := NewWebhookFilterEqualsValueKnown()
	filterEquals.Doc = types.StringValue("sys.type")
	filterEquals.Value = types.StringValue("abc")

	filterIn := NewWebhookFilterInValueKnown(ctx)
	filterIn.Doc = types.StringValue("sys.type")
	filterIn.Values = NewTypedList([]types.String{types.StringValue("abc"), types.StringValue("def")})

	filterRegexp := NewWebhookFilterRegexpValueKnown()
	filterRegexp.Doc = types.StringValue("sys.type")
	filterRegexp.Pattern = types.StringValue("abc.*")

	filterNotEquals := NewWebhookFilterNotValueKnown()
	filterNotEquals.Equals = filterEquals

	filterNotIn := NewWebhookFilterNotValueKnown()
	filterNotIn.In = filterIn

	filterNotRegexp := NewWebhookFilterNotValueKnown()
	filterNotRegexp.Regexp = filterRegexp

	filterFilterEquals := NewWebhookFilterValueKnown()
	filterFilterEquals.Equals = filterEquals

	filterFilterIn := NewWebhookFilterValueKnown()
	filterFilterIn.In = filterIn

	filterFilterRegexp := NewWebhookFilterValueKnown()
	filterFilterRegexp.Regexp = filterRegexp

	filterFilterNotEquals := NewWebhookFilterValueKnown()
	filterFilterNotEquals.Not = filterNotEquals

	filterFilterNotIn := NewWebhookFilterValueKnown()
	filterFilterNotIn.Not = filterNotIn

	filterFilterNotRegexp := NewWebhookFilterValueKnown()
	filterFilterNotRegexp.Not = filterNotRegexp

	filters := NewTypedList([]WebhookFilterValue{
		filterFilterEquals,
		filterFilterIn,
		filterFilterRegexp,
		filterFilterNotEquals,
		filterFilterNotIn,
		filterFilterNotRegexp,
	})

	return filters
}

// func TestWebhookFilterTypeValueFromObject(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()

// 	typs := []AttrTypeWithValueFromObject{
// 		WebhookFilterValue{}.CustomType(ctx),
// 		WebhookFilterNotType{},
// 		WebhookFilterEqualsType{},
// 		WebhookFilterInType{},
// 		WebhookFilterRegexpType{},
// 	}

// 	tfvalniltype := tftypes.NewValue(nil, nil)

// 	for _, typ := range typs {

// 		// types.ObjectNull(typ.ValType(ctx).)
// 		// tftyp := typ.TerraformType(ctx)

// 		// tfvalnull := tftypes.NewValue(tftyp, nil)
// 		// tfvalunknown := tftypes.NewValue(tftyp, tftypes.UnknownValue)

// 		// valueNil, err := typ.ValueFromObject(ctx, tfvalniltype)
// 		// assert.NoError(t, err)
// 		// assert.True(t, valueNil.IsNull())

// 		// valueNull, err := typ.ValueFromObject(ctx, tfvalnull)
// 		// assert.NoError(t, err)
// 		// assert.True(t, valueNull.IsNull())

// 		// valueUnknown, err := typ.ValueFromObject(ctx, typs)
// 		// assert.NoError(t, err)
// 		// assert.True(t, valueUnknown.IsUnknown())
// 	}
// }

func TestWebhookFilterValueEqual(t *testing.T) {
	t.Parallel()

	filtersList := webhookFiltersListForTesting(t)

	//nolint:gocritic
	assert.True(t, filtersList.Equal(filtersList))
}

func TestWebhookFilterValueToObjectValueUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []AttrValueWithToObjectValue{
		NewWebhookFilterValueUnknown(),
		NewWebhookFilterNotValueUnknown(),
		NewWebhookFilterEqualsValueUnknown(),
		NewWebhookFilterInValueUnknown(),
		NewWebhookFilterRegexpValueUnknown(),
	}

	for _, value := range values {
		assert.True(t, value.IsUnknown())

		objectValue, objectValueDiags := value.ToObjectValue(ctx)
		assert.Empty(t, objectValueDiags)

		assert.True(t, objectValue.IsUnknown())
	}
}

func TestWebhookFilterValueToObjectValue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []AttrValueWithToObjectValue{
		NewWebhookFilterValueKnown(),
		NewWebhookFilterNotValueKnown(),
		NewWebhookFilterEqualsValueKnown(),
		NewWebhookFilterInValueKnown(ctx),
		NewWebhookFilterRegexpValueKnown(),
	}

	for _, value := range values {
		objectValue, objectValueDiags := value.ToObjectValue(ctx)
		assert.Empty(t, objectValueDiags)

		assert.False(t, objectValue.IsNull())
		assert.False(t, objectValue.IsUnknown())
	}
}

func TestWebhookFilterValueToTerraformValueNull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []attr.Value{
		NewWebhookFilterValueNull(),
		NewWebhookFilterNotValueNull(),
		NewWebhookFilterEqualsValueNull(),
		NewWebhookFilterInValueNull(),
		NewWebhookFilterRegexpValueNull(),
	}

	for _, value := range values {
		objectValue, err := value.ToTerraformValue(ctx)
		require.NoError(t, err)

		assert.True(t, objectValue.IsKnown())
		assert.True(t, objectValue.IsNull())
	}
}

func TestWebhookFilterValueToTerraformValueUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []attr.Value{
		NewWebhookFilterValueUnknown(),
		NewWebhookFilterNotValueUnknown(),
		NewWebhookFilterEqualsValueUnknown(),
		NewWebhookFilterInValueUnknown(),
		NewWebhookFilterRegexpValueUnknown(),
	}

	for _, value := range values {
		objectValue, err := value.ToTerraformValue(ctx)
		require.NoError(t, err)

		assert.False(t, objectValue.IsKnown())
	}
}
