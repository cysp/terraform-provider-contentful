package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterArrayPreservesOmissionAndRejectsUnknownContainer(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters")
	tests := map[string]struct {
		value         TypedList[TypedObject[WebhookFilterValue]]
		expected      cm.OptNilWebhookDefinitionFilterArray
		expectedPaths []string
	}{
		"null is omitted": {
			value:         NewTypedListNull[TypedObject[WebhookFilterValue]](),
			expected:      cm.NewOptNilWebhookDefinitionFilterArrayNull(),
			expectedPaths: []string{},
		},
		"known empty list is preserved": {
			value:         NewTypedList([]TypedObject[WebhookFilterValue]{}),
			expected:      cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{}),
			expectedPaths: []string{},
		},
		"unknown is rejected": {
			value:         NewTypedListUnknown[TypedObject[WebhookFilterValue]](),
			expectedPaths: []string{"filters"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToOptNilWebhookDefinitionFilterArray(t.Context(), filterPath, test.value)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestWebhookFilterArrayRejectsUnresolvedElementsAtomically(t *testing.T) {
	t.Parallel()

	filters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
		NewTypedObjectNull[WebhookFilterValue](),
		NewTypedObjectUnknown[WebhookFilterValue](),
	})

	actual, diags := ToOptNilWebhookDefinitionFilterArray(t.Context(), path.Root("filters"), filters)

	assert.Equal(t, cm.OptNilWebhookDefinitionFilterArray{}, actual)
	assert.Equal(t, []string{"filters[1]", "filters[2]"}, diagnosticPaths(t, diags))
}

func TestWebhookFilterRequiresExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)
	equals := webhookEqualsValue(types.StringValue("sys.type"), types.StringValue("Entry"))
	inFilter := webhookInValue(types.StringValue("sys.id"), NewTypedList([]types.String{types.StringValue("entry")}))
	regexp := webhookRegexpValue(types.StringValue("sys.id"), types.StringValue("entry.*"))
	not := NewTypedObject(webhookNotValue(equals, NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()))

	tests := map[string]struct {
		value         WebhookFilterValue
		expected      cm.WebhookDefinitionFilter
		expectedPaths []string
	}{
		"not": {
			value: webhookFilterValue(not, NewTypedObjectNull[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expected: cm.WebhookDefinitionFilter{
				Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
					Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
				}),
			},
			expectedPaths: []string{},
		},
		"equals": {
			value: webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), equals, NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expected: cm.WebhookDefinitionFilter{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
			},
			expectedPaths: []string{},
		},
		"in": {
			value: webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), NewTypedObjectNull[WebhookFilterEqualsValue](), inFilter, NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expected: cm.WebhookDefinitionFilter{
				In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)},
			},
			expectedPaths: []string{},
		},
		"regexp": {
			value: webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), NewTypedObjectNull[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), regexp),
			expected: cm.WebhookDefinitionFilter{
				Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.id"}`), []byte(`{"pattern":"entry.*"}`)},
			},
			expectedPaths: []string{},
		},
		"none": {
			value:         webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), NewTypedObjectNull[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expectedPaths: []string{"filters[0]"},
		},
		"multiple": {
			value:         webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), equals, inFilter, NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expectedPaths: []string{"filters[0]"},
		},
		"unknown alternative": {
			value:         webhookFilterValue(NewTypedObjectNull[WebhookFilterNotValue](), equals, NewTypedObjectUnknown[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expectedPaths: []string{"filters[0].in"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionFilter(t.Context(), filterPath, test.value)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestWebhookFilterNotRequiresExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0).AtName("not")
	equals := webhookEqualsValue(types.StringValue("sys.type"), types.StringValue("Entry"))
	inFilter := webhookInValue(types.StringValue("sys.id"), NewTypedList([]types.String{types.StringValue("entry")}))
	regexp := webhookRegexpValue(types.StringValue("sys.id"), types.StringValue("entry.*"))

	tests := map[string]struct {
		value         WebhookFilterNotValue
		expected      cm.OptWebhookDefinitionFilterNot
		expectedPaths []string
	}{
		"equals": {
			value:         webhookNotValue(equals, NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expected:      cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)}}),
			expectedPaths: []string{},
		},
		"in": {
			value:         webhookNotValue(NewTypedObjectNull[WebhookFilterEqualsValue](), inFilter, NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expected:      cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)}}),
			expectedPaths: []string{},
		},
		"regexp": {
			value:         webhookNotValue(NewTypedObjectNull[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), regexp),
			expected:      cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.id"}`), []byte(`{"pattern":"entry.*"}`)}}),
			expectedPaths: []string{},
		},
		"none": {
			value:         webhookNotValue(NewTypedObjectNull[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expectedPaths: []string{"filters[0].not"},
		},
		"multiple": {
			value:         webhookNotValue(equals, inFilter, NewTypedObjectNull[WebhookFilterRegexpValue]()),
			expectedPaths: []string{"filters[0].not"},
		},
		"unknown alternative": {
			value:         webhookNotValue(NewTypedObjectUnknown[WebhookFilterEqualsValue](), NewTypedObjectNull[WebhookFilterInValue](), regexp),
			expectedPaths: []string{"filters[0].not.equals"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionFilterNot(t.Context(), filterPath, test.value)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestWebhookFilterOperandsRejectUnresolvedValues(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)
	tests := map[string]struct {
		convert       func() (any, []string)
		expectedPaths []string
	}{
		"equals unknown doc": {
			convert:       webhookEqualsConversion(t, filterPath.AtName("equals"), types.StringUnknown(), types.StringValue("Entry")),
			expectedPaths: []string{"filters[0].equals.doc"},
		},
		"equals null value": {
			convert:       webhookEqualsConversion(t, filterPath.AtName("equals"), types.StringValue("sys.type"), types.StringNull()),
			expectedPaths: []string{"filters[0].equals.value"},
		},
		"in null doc": {
			convert:       webhookInConversion(t, filterPath.AtName("in"), types.StringNull(), NewTypedList([]types.String{})),
			expectedPaths: []string{"filters[0].in.doc"},
		},
		"in unknown values": {
			convert:       webhookInConversion(t, filterPath.AtName("in"), types.StringValue("sys.id"), NewTypedListUnknown[types.String]()),
			expectedPaths: []string{"filters[0].in.values"},
		},
		"in null values": {
			convert:       webhookInConversion(t, filterPath.AtName("in"), types.StringValue("sys.id"), NewTypedListNull[types.String]()),
			expectedPaths: []string{"filters[0].in.values"},
		},
		"in unresolved elements": {
			convert: webhookInConversion(t, filterPath.AtName("in"), types.StringValue("sys.id"), NewTypedList([]types.String{
				types.StringValue("entry"),
				types.StringNull(),
				types.StringUnknown(),
			})),
			expectedPaths: []string{"filters[0].in.values[1]", "filters[0].in.values[2]"},
		},
		"regexp unknown doc and pattern": {
			convert:       webhookRegexpConversion(t, filterPath.AtName("regexp"), types.StringUnknown(), types.StringUnknown()),
			expectedPaths: []string{"filters[0].regexp.doc", "filters[0].regexp.pattern"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, paths := test.convert()

			assert.Zero(t, actual)
			assert.Equal(t, test.expectedPaths, paths)
		})
	}
}

func TestWebhookFilterArrayReturnsNoPartialOutputAfterElementError(t *testing.T) {
	t.Parallel()

	filters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
		webhookEqualsFilter(types.StringUnknown(), types.StringValue("Entry")),
	})

	actual, diags := ToOptNilWebhookDefinitionFilterArray(t.Context(), path.Root("filters"), filters)

	assert.Equal(t, cm.OptNilWebhookDefinitionFilterArray{}, actual)
	assert.Equal(t, []string{"filters[1].equals.doc"}, diagnosticPaths(t, diags))
}

func webhookFilterValue(
	not TypedObject[WebhookFilterNotValue],
	equals TypedObject[WebhookFilterEqualsValue],
	in TypedObject[WebhookFilterInValue],
	regexp TypedObject[WebhookFilterRegexpValue],
) WebhookFilterValue {
	return WebhookFilterValue{Not: not, Equals: equals, In: in, Regexp: regexp}
}

func webhookNotValue(
	equals TypedObject[WebhookFilterEqualsValue],
	in TypedObject[WebhookFilterInValue],
	regexp TypedObject[WebhookFilterRegexpValue],
) WebhookFilterNotValue {
	return WebhookFilterNotValue{Equals: equals, In: in, Regexp: regexp}
}

func webhookEqualsValue(doc types.String, value types.String) TypedObject[WebhookFilterEqualsValue] {
	return NewTypedObject(WebhookFilterEqualsValue{Doc: doc, Value: value})
}

func webhookInValue(doc types.String, values TypedList[types.String]) TypedObject[WebhookFilterInValue] {
	return NewTypedObject(WebhookFilterInValue{Doc: doc, Values: values})
}

func webhookRegexpValue(doc types.String, pattern types.String) TypedObject[WebhookFilterRegexpValue] {
	return NewTypedObject(WebhookFilterRegexpValue{Doc: doc, Pattern: pattern})
}

func webhookEqualsFilter(doc types.String, value types.String) TypedObject[WebhookFilterValue] {
	return NewTypedObject(webhookFilterValue(
		NewTypedObjectNull[WebhookFilterNotValue](),
		webhookEqualsValue(doc, value),
		NewTypedObjectNull[WebhookFilterInValue](),
		NewTypedObjectNull[WebhookFilterRegexpValue](),
	))
}

func webhookEqualsConversion(
	t *testing.T,
	valuePath path.Path,
	doc types.String,
	value types.String,
) func() (any, []string) {
	t.Helper()

	return func() (any, []string) {
		actual, diags := ToWebhookDefinitionFilterEquals(t.Context(), valuePath, WebhookFilterEqualsValue{Doc: doc, Value: value})

		return actual, diagnosticPaths(t, diags)
	}
}

func webhookInConversion(
	t *testing.T,
	valuePath path.Path,
	doc types.String,
	values TypedList[types.String],
) func() (any, []string) {
	t.Helper()

	return func() (any, []string) {
		actual, diags := ToWebhookDefinitionFilterIn(t.Context(), valuePath, WebhookFilterInValue{Doc: doc, Values: values})

		return actual, diagnosticPaths(t, diags)
	}
}

func webhookRegexpConversion(
	t *testing.T,
	valuePath path.Path,
	doc types.String,
	pattern types.String,
) func() (any, []string) {
	t.Helper()

	return func() (any, []string) {
		actual, diags := ToWebhookDefinitionFilterRegexp(t.Context(), valuePath, WebhookFilterRegexpValue{Doc: doc, Pattern: pattern})

		return actual, diagnosticPaths(t, diags)
	}
}
