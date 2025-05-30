package provider_test

import (
	"testing"

	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webhookFiltersListForTesting(t *testing.T) provider.TypedList[provider.WebhookFilterValue] {
	t.Helper()

	ctx := t.Context()

	filterEquals := provider.NewWebhookFilterEqualsValueKnown()
	filterEquals.Doc = types.StringValue("sys.type")
	filterEquals.Value = types.StringValue("abc")

	filterIn := provider.NewWebhookFilterInValueKnown(ctx)
	filterIn.Doc = types.StringValue("sys.type")
	filterIn.Values = DiagsNoErrorsMust(provider.NewTypedList(ctx, []types.String{types.StringValue("abc"), types.StringValue("def")}))

	filterRegexp := provider.NewWebhookFilterRegexpValueKnown()
	filterRegexp.Doc = types.StringValue("sys.type")
	filterRegexp.Pattern = types.StringValue("abc.*")

	filterNotEquals := provider.NewWebhookFilterNotValueKnown()
	filterNotEquals.Equals = filterEquals

	filterNotIn := provider.NewWebhookFilterNotValueKnown()
	filterNotIn.In = filterIn

	filterNotRegexp := provider.NewWebhookFilterNotValueKnown()
	filterNotRegexp.Regexp = filterRegexp

	filterFilterEquals := provider.NewWebhookFilterValueKnown()
	filterFilterEquals.Equals = filterEquals

	filterFilterIn := provider.NewWebhookFilterValueKnown()
	filterFilterIn.In = filterIn

	filterFilterRegexp := provider.NewWebhookFilterValueKnown()
	filterFilterRegexp.Regexp = filterRegexp

	filterFilterNotEquals := provider.NewWebhookFilterValueKnown()
	filterFilterNotEquals.Not = filterNotEquals

	filterFilterNotIn := provider.NewWebhookFilterValueKnown()
	filterFilterNotIn.Not = filterNotIn

	filterFilterNotRegexp := provider.NewWebhookFilterValueKnown()
	filterFilterNotRegexp.Not = filterNotRegexp

	filters, filtersDiags := provider.NewTypedList(ctx, []provider.WebhookFilterValue{
		filterFilterEquals,
		filterFilterIn,
		filterFilterRegexp,
		filterFilterNotEquals,
		filterFilterNotIn,
		filterFilterNotRegexp,
	})

	assert.Empty(t, filtersDiags)

	return filters
}

// func TestWebhookFilterTypeValueFromObject(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()

// 	typs := []provider.AttrTypeWithValueFromObject{
// 		provider.WebhookFilterValue{}.CustomType(ctx),
// 		provider.WebhookFilterNotType{},
// 		provider.WebhookFilterEqualsType{},
// 		provider.WebhookFilterInType{},
// 		provider.WebhookFilterRegexpType{},
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

	values := []provider.AttrValueWithToObjectValue{
		provider.NewWebhookFilterValueUnknown(),
		provider.NewWebhookFilterNotValueUnknown(),
		provider.NewWebhookFilterEqualsValueUnknown(),
		provider.NewWebhookFilterInValueUnknown(),
		provider.NewWebhookFilterRegexpValueUnknown(),
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

	values := []provider.AttrValueWithToObjectValue{
		provider.NewWebhookFilterValueKnown(),
		provider.NewWebhookFilterNotValueKnown(),
		provider.NewWebhookFilterEqualsValueKnown(),
		provider.NewWebhookFilterInValueKnown(ctx),
		provider.NewWebhookFilterRegexpValueKnown(),
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
		provider.NewWebhookFilterValueNull(),
		provider.NewWebhookFilterNotValueNull(),
		provider.NewWebhookFilterEqualsValueNull(),
		provider.NewWebhookFilterInValueNull(),
		provider.NewWebhookFilterRegexpValueNull(),
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
		provider.NewWebhookFilterValueUnknown(),
		provider.NewWebhookFilterNotValueUnknown(),
		provider.NewWebhookFilterEqualsValueUnknown(),
		provider.NewWebhookFilterInValueUnknown(),
		provider.NewWebhookFilterRegexpValueUnknown(),
	}

	for _, value := range values {
		objectValue, err := value.ToTerraformValue(ctx)
		require.NoError(t, err)

		assert.False(t, objectValue.IsKnown())
	}
}
