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

func TestWebhookFilterArrayPreservesNullAndEmptyAndRejectsUnknownContainer(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters")
	tests := map[string]struct {
		value               TypedList[TypedObject[WebhookFilterValue]]
		expected            cm.OptNilWebhookDefinitionFilterArray
		expectedWire        string
		expectedRequestWire string
		expectedPaths       []string
	}{
		"null is explicit null": {
			value:               NewTypedListNull[TypedObject[WebhookFilterValue]](),
			expected:            cm.NewOptNilWebhookDefinitionFilterArrayNull(),
			expectedWire:        `null`,
			expectedRequestWire: `{"name":"","url":"","topics":[],"filters":null}`,
			expectedPaths:       []string{},
		},
		"known empty list is preserved": {
			value:               NewTypedList([]TypedObject[WebhookFilterValue]{}),
			expected:            cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{}),
			expectedWire:        `[]`,
			expectedRequestWire: `{"name":"","url":"","topics":[],"filters":[]}`,
			expectedPaths:       []string{},
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
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))

			if test.expectedWire != "" {
				encoded, err := actual.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, test.expectedWire, string(encoded))

				request := cm.WebhookDefinitionData{Filters: actual}
				encodedRequest, err := request.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, test.expectedRequestWire, string(encodedRequest))
			}
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
	assert.Equal(t, []string{"filters[1]", "filters[2]"}, attributeDiagnosticPaths(t, diags))
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

			actual, diags := ToWebhookDefinitionFilter(filterPath, test.value)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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

			actual, diags := ToWebhookDefinitionFilterNot(filterPath, test.value)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestWebhookFilterEqualsOperands(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("filters").AtListIndex(0).AtName("equals")

	for name, test := range map[string]struct {
		value         WebhookFilterEqualsValue
		expectedPaths []string
	}{
		"unknown doc": {
			value:         WebhookFilterEqualsValue{Doc: types.StringUnknown(), Value: types.StringValue("Entry")},
			expectedPaths: []string{"filters[0].equals.doc"},
		},
		"null value": {
			value:         WebhookFilterEqualsValue{Doc: types.StringValue("sys.type"), Value: types.StringNull()},
			expectedPaths: []string{"filters[0].equals.value"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionFilterEquals(valuePath, test.value)

			assert.Nil(t, actual)
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
		})
	}

	result, diags := ToWebhookDefinitionFilterEquals(valuePath, WebhookFilterEqualsValue{
		Doc:   types.StringValue(""),
		Value: types.StringValue(""),
	})
	require.Empty(t, diags.Errors())

	encodedResult, err := result.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `[{"doc":""},""]`, string(encodedResult))

	filter, filterDiags := ToWebhookDefinitionFilter(
		path.Root("filters").AtListIndex(0),
		webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			webhookEqualsValue(types.StringValue(""), types.StringValue("")),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		),
	)
	require.Empty(t, filterDiags.Errors())

	encodedFilter, err := filter.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"equals":[{"doc":""},""]}`, string(encodedFilter))
}

func TestWebhookFilterInOperands(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("filters").AtListIndex(0).AtName("in")

	for name, test := range map[string]struct {
		value         WebhookFilterInValue
		expectedPaths []string
	}{
		"null doc": {
			value:         WebhookFilterInValue{Doc: types.StringNull(), Values: NewTypedList([]types.String{})},
			expectedPaths: []string{"filters[0].in.doc"},
		},
		"unknown values": {
			value:         WebhookFilterInValue{Doc: types.StringValue("sys.id"), Values: NewTypedListUnknown[types.String]()},
			expectedPaths: []string{"filters[0].in.values"},
		},
		"null values": {
			value:         WebhookFilterInValue{Doc: types.StringValue("sys.id"), Values: NewTypedListNull[types.String]()},
			expectedPaths: []string{"filters[0].in.values"},
		},
		"unresolved elements": {
			value: WebhookFilterInValue{
				Doc: types.StringValue("sys.id"),
				Values: NewTypedList([]types.String{
					types.StringValue("entry"),
					types.StringNull(),
					types.StringUnknown(),
				}),
			},
			expectedPaths: []string{"filters[0].in.values[1]", "filters[0].in.values[2]"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionFilterIn(valuePath, test.value)

			assert.Nil(t, actual)
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
		})
	}

	result, diags := ToWebhookDefinitionFilterIn(valuePath, WebhookFilterInValue{
		Doc:    types.StringValue("sys.id"),
		Values: NewTypedList([]types.String{}),
	})
	require.Empty(t, diags.Errors())

	encodedResult, err := result.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `[{"doc":"sys.id"},[]]`, string(encodedResult))

	filter, filterDiags := ToWebhookDefinitionFilter(
		path.Root("filters").AtListIndex(0),
		webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			webhookInValue(types.StringValue("sys.id"), NewTypedList([]types.String{})),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		),
	)
	require.Empty(t, filterDiags.Errors())

	encodedFilter, err := filter.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"in":[{"doc":"sys.id"},[]]}`, string(encodedFilter))
}

func TestWebhookFilterRegexpOperands(t *testing.T) {
	t.Parallel()

	actual, diags := ToWebhookDefinitionFilterRegexp(
		path.Root("filters").AtListIndex(0).AtName("regexp"),
		WebhookFilterRegexpValue{Doc: types.StringUnknown(), Pattern: types.StringUnknown()},
	)

	assert.Nil(t, actual)
	assert.Equal(t, []string{"filters[0].regexp.doc", "filters[0].regexp.pattern"}, attributeDiagnosticPaths(t, diags))
}

func TestWebhookFilterArrayReturnsNoPartialOutputAfterElementError(t *testing.T) {
	t.Parallel()

	filters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
		webhookEqualsFilter(types.StringUnknown(), types.StringValue("Entry")),
	})

	actual, diags := ToOptNilWebhookDefinitionFilterArray(t.Context(), path.Root("filters"), filters)

	assert.Equal(t, cm.OptNilWebhookDefinitionFilterArray{}, actual)
	assert.Equal(t, []string{"filters[1].equals.doc"}, attributeDiagnosticPaths(t, diags))
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
