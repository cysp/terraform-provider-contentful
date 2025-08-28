package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webhookFiltersListForTesting(t *testing.T) TypedList[TypedObject[WebhookFilterValue]] {
	t.Helper()

	ctx := t.Context()

	filterEquals := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterEqualsValue](ctx, map[string]attr.Value{
		"doc":   types.StringValue("sys.type"),
		"value": types.StringValue("abc"),
	}))

	filterIn := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterInValue](ctx, map[string]attr.Value{
		"doc":    types.StringValue("sys.type"),
		"values": NewTypedList([]types.String{types.StringValue("abc"), types.StringValue("def")}),
	}))

	filterRegexp := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterRegexpValue](ctx, map[string]attr.Value{
		"doc":     types.StringValue("sys.type"),
		"pattern": types.StringValue("abc.*"),
	}))

	filterNotEquals := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterNotValue](ctx, map[string]attr.Value{
		"equals": filterEquals,
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterNotIn := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterNotValue](ctx, map[string]attr.Value{
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     filterIn,
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterNotRegexp := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterNotValue](ctx, map[string]attr.Value{
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": filterRegexp,
	}))

	filterFilterEquals := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    NewTypedObjectNull[WebhookFilterNotValue](),
		"equals": filterEquals,
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterFilterIn := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    NewTypedObjectNull[WebhookFilterNotValue](),
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     filterIn,
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterFilterRegexp := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    NewTypedObjectNull[WebhookFilterNotValue](),
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": filterRegexp,
	}))

	filterFilterNotEquals := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    filterNotEquals,
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterFilterNotIn := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    filterNotIn,
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filterFilterNotRegexp := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, map[string]attr.Value{
		"not":    filterNotRegexp,
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}))

	filters := NewTypedList([]TypedObject[WebhookFilterValue]{
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
		NewTypedObjectUnknown[WebhookFilterValue](),
		NewTypedObjectUnknown[WebhookFilterNotValue](),
		NewTypedObjectUnknown[WebhookFilterEqualsValue](),
		NewTypedObjectUnknown[WebhookFilterInValue](),
		NewTypedObjectUnknown[WebhookFilterRegexpValue](),
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
		NewTypedObject(WebhookFilterValue{}),
		NewTypedObject(WebhookFilterNotValue{}),
		NewTypedObject(WebhookFilterEqualsValue{}),
		NewTypedObject(WebhookFilterInValue{}),
		NewTypedObject(WebhookFilterRegexpValue{}),
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
		NewTypedObjectNull[WebhookFilterValue](),
		NewTypedObjectNull[WebhookFilterNotValue](),
		NewTypedObjectNull[WebhookFilterEqualsValue](),
		NewTypedObjectNull[WebhookFilterInValue](),
		NewTypedObjectNull[WebhookFilterRegexpValue](),
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
		NewTypedObjectUnknown[WebhookFilterValue](),
		NewTypedObjectUnknown[WebhookFilterNotValue](),
		NewTypedObjectUnknown[WebhookFilterEqualsValue](),
		NewTypedObjectUnknown[WebhookFilterInValue](),
		NewTypedObjectUnknown[WebhookFilterRegexpValue](),
	}

	for _, value := range values {
		objectValue, err := value.ToTerraformValue(ctx)
		require.NoError(t, err)

		assert.False(t, objectValue.IsKnown())
	}
}

func TestWebhookFilterValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	attributes := map[string]attr.Value{
		"not":    NewTypedObjectNull[WebhookFilterNotValue](),
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := NewTypedObjectFromAttributes[WebhookFilterValue](ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
