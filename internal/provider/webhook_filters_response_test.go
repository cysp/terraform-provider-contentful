package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWebhookFilterValueFromResponsePreservesRepresentableAlternatives(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, cm.WebhookDefinitionFilter{})

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		assert.True(t, actual.Value().Not.IsNull())
		assert.True(t, actual.Value().Equals.IsNull())
		assert.True(t, actual.Value().In.IsNull())
		assert.True(t, actual.Value().Regexp.IsNull())
	})

	t.Run("multiple top-level alternatives", func(t *testing.T) {
		t.Parallel()

		input := cm.WebhookDefinitionFilter{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
			In:     cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)},
		}

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, input)

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		assert.False(t, actual.Value().Equals.IsNull())
		assert.False(t, actual.Value().In.IsNull())
		assert.Equal(t, "Entry", actual.Value().Equals.Value().Value.ValueString())
		assert.Equal(t, []types.String{types.StringValue("entry")}, actual.Value().In.Value().Values.Elements())
	})

	t.Run("multiple negated alternatives", func(t *testing.T) {
		t.Parallel()

		input := cm.WebhookDefinitionFilter{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
				In:     cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)},
			}),
		}

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, input)

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		require.False(t, actual.Value().Not.IsNull())
		assert.False(t, actual.Value().Not.Value().Equals.IsNull())
		assert.False(t, actual.Value().Not.Value().In.IsNull())
	})
}

func TestReadWebhookFilterValueFromResponsePreservesValidJSONShapeMismatch(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)
	tests := map[string]struct {
		input        cm.WebhookDefinitionFilter
		expectedPath path.Path
	}{
		"malformed binary array": {
			input: cm.WebhookDefinitionFilter{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`)},
			},
			expectedPath: filterPath.AtName("equals"),
		},
		"incompatible term": {
			input: cm.WebhookDefinitionFilter{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`123`)},
			},
			expectedPath: filterPath.AtName("equals").AtName("value"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, test.input)

			require.False(t, diags.HasError())
			require.Len(t, diags.Warnings(), 1)
			assert.Equal(t, test.expectedPath.String(), webhookWarningPaths(t, diags)[0])
			assert.False(t, actual.IsNull())
		})
	}
}

func TestReadWebhookFiltersListValueFromResponsePreservesSiblingsAndPositions(t *testing.T) {
	t.Parallel()

	input := cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
		},
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`)},
		},
	})

	actual, diags := ReadWebhookFiltersListValueFromResponse(t.Context(), path.Root("filters"), input)

	require.False(t, diags.HasError())
	require.Len(t, diags.Warnings(), 1)
	assert.Equal(t, "filters[1].equals", webhookWarningPaths(t, diags)[0])
	require.Len(t, actual.Elements(), 2)
	assert.False(t, actual.Elements()[0].Value().Equals.IsNull())
	assert.True(t, actual.Elements()[1].Value().Equals.IsNull())
}

func TestWebhookFilterResponseDecoderRetainsUnsupportedProperties(t *testing.T) {
	t.Parallel()

	var decoded cm.OptNilWebhookDefinitionFilterArray
	require.NoError(t, decoded.UnmarshalJSON([]byte(`[
  {
    "equals": [{"doc": "sys.type"}, "Entry"],
    "futureTop": {"mode": "preview"}
  },
  {
    "not": {
      "in": [{"doc": "sys.id"}, ["entry"]],
      "futureNot": [1, 2]
    }
  }
]`)))

	filters, ok := decoded.Get()
	require.True(t, ok)
	require.Len(t, filters, 2)
	require.Len(t, filters[0].AdditionalProps, 1)
	assert.JSONEq(t, `{"mode":"preview"}`, string(filters[0].AdditionalProps["futureTop"]))

	negated, ok := filters[1].Not.Get()
	require.True(t, ok)
	require.Len(t, negated.AdditionalProps, 1)
	assert.JSONEq(t, `[1,2]`, string(negated.AdditionalProps["futureNot"]))
}

func TestWebhookFilterResponseProjectionReportsUnsupportedProperties(t *testing.T) {
	t.Parallel()

	var decoded cm.OptNilWebhookDefinitionFilterArray
	require.NoError(t, decoded.UnmarshalJSON([]byte(`[
  {
    "equals": [{"doc": "sys.type", "zeta": false, "alpha": 1}, "Entry"],
    "futureTop": [{"doc": "sys.id"}, "ignored"]
  },
  {
    "not": {
      "equals": [{"doc": "sys.id"}, "entry"],
      "futureNot": [{"doc": "sys.type"}, "Asset"]
    }
  },
  {
    "futureOnly": [{"doc": "sys.type"}, "Asset"]
  },
  {
    "in": [{"doc": "sys.id"}, ["first", "second"]]
  },
  {
    "equals": [{"futureDoc": "sys.type"}, "Asset"]
  },
  {
    "regexp": [{"doc": "sys.id"}, {"futurePattern": "entry.*"}]
  },
  {
    "in": [{"zeta": false, "alpha": 1}, ["third"]]
  }
]`)))

	model, diags := NewWebhookResourceModelFromResponse(t.Context(), cm.WebhookDefinition{
		Sys:     cm.NewWebhookDefinitionSys("space", "webhook"),
		Name:    "Response webhook",
		URL:     "https://example.com/webhook",
		Topics:  []string{"Entry.create", "Entry.save"},
		Filters: decoded,
		Headers: cm.WebhookDefinitionHeaders{},
	}, map[string]TypedObject[WebhookHeaderValue]{})

	expectedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			webhookEqualsValue(types.StringValue("sys.type"), types.StringValue("Entry")),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObject(webhookNotValue(
				webhookEqualsValue(types.StringValue("sys.id"), types.StringValue("entry")),
				NewTypedObjectNull[WebhookFilterInValue](),
				NewTypedObjectNull[WebhookFilterRegexpValue](),
			)),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			webhookInValue(types.StringValue("sys.id"), NewTypedList([]types.String{
				types.StringValue("first"),
				types.StringValue("second"),
			})),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			webhookEqualsValue(types.StringNull(), types.StringValue("Asset")),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			NewTypedObjectNull[WebhookFilterInValue](),
			webhookRegexpValue(types.StringValue("sys.id"), types.StringNull()),
		)),
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			webhookInValue(types.StringNull(), NewTypedList([]types.String{types.StringValue("third")})),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
	})

	assert.Equal(t, expectedFilters, model.Filters)
	assert.Equal(t, "Response webhook", model.Name.ValueString())
	assert.Equal(t, "https://example.com/webhook", model.URL.ValueString())
	assert.Equal(t, []types.String{types.StringValue("Entry.create"), types.StringValue("Entry.save")}, model.Topics.Elements())

	expectedDiagnostics := []struct {
		path    string
		summary string
		detail  string
	}{
		{
			path:    "filters[0]",
			summary: "Unsupported webhook filter response properties",
			detail:  `Contentful returned webhook filter properties ["futureTop"] that this provider cannot represent. The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[0].equals.doc",
			summary: "Unsupported webhook filter term response properties",
			detail:  `Contentful returned unsupported properties ["alpha" "zeta"] in a webhook filter term object whose supported property is "doc". The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[1].not",
			summary: "Unsupported webhook filter response properties",
			detail:  `Contentful returned webhook filter properties ["futureNot"] that this provider cannot represent. The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[2]",
			summary: "Unsupported webhook filter response properties",
			detail:  `Contentful returned webhook filter properties ["futureOnly"] that this provider cannot represent. The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[4].equals.doc",
			summary: "Unsupported webhook filter term response properties",
			detail:  `Contentful returned unsupported properties ["futureDoc"] in a webhook filter term object whose supported property is "doc". The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[4].equals.doc",
			summary: "Unsupported webhook filter value",
			detail:  `Contentful returned an object without the required "doc" property. Terraform state retains a null value; a later request conversion will reject it until configured.`,
		},
		{
			path:    "filters[5].regexp.pattern",
			summary: "Unsupported webhook filter term response properties",
			detail:  `Contentful returned unsupported properties ["futurePattern"] in a webhook filter term object whose supported property is "pattern". The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[5].regexp.pattern",
			summary: "Unsupported webhook filter value",
			detail:  `Contentful returned an object without the required "pattern" property. Terraform state retains a null value; a later request conversion will reject it until configured.`,
		},
		{
			path:    "filters[6].in.doc",
			summary: "Unsupported webhook filter term response properties",
			detail:  `Contentful returned unsupported properties ["alpha" "zeta"] in a webhook filter term object whose supported property is "doc". The unsupported properties are omitted from Terraform state, and a later Terraform update to this webhook cannot preserve them.`,
		},
		{
			path:    "filters[6].in.doc",
			summary: "Unsupported webhook filter value",
			detail:  `Contentful returned an object without the required "doc" property. Terraform state retains a null value; a later request conversion will reject it until configured.`,
		},
	}

	require.Len(t, diags, len(expectedDiagnostics))

	for index, expected := range expectedDiagnostics {
		diagnostic := diags[index]
		assert.Equal(t, diag.SeverityWarning, diagnostic.Severity())
		assert.Equal(t, expected.summary, diagnostic.Summary())
		assert.Equal(t, expected.detail, diagnostic.Detail())

		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)
		assert.Equal(t, expected.path, withPath.Path().String())
	}
}

func TestReadWebhookDefinitionFilterTermString(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`"abc"`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`123`),
			expectError: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := ReadWebhookDefinitionFilterTermString(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func webhookWarningPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Warnings()))
	for _, diagnostic := range diags.Warnings() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}

func TestReadWebhookDefinitionFilterTermStringArray(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`["abc"]`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`"abc"`),
			expectError: false,
		},
		"wrong element type": {
			input:       []byte(`["abc",123]`),
			expectError: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := ReadWebhookDefinitionFilterTermStringArray(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func TestReadWebhookDefinitionFilterTermStringObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		name        string
		expectNull  bool
		expectValue string
		expectError bool
	}{
		"valid": {
			input:       []byte(`{"value":"abc"}`),
			name:        "value",
			expectValue: "abc",
		},
		"valid with excess": {
			input:       []byte(`{"a":"b","value":"abc","c":"d"}`),
			name:        "value",
			expectValue: "abc",
		},
		"value absent": {
			input:      []byte(`{"a":"b"}`),
			name:       "value",
			expectNull: true,
		},
		"value wrong type": {
			input:      []byte(`{"value":123}`),
			name:       "value",
			expectNull: true,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			name:        "value",
			expectError: true,
		},
		"wrong type": {
			input:      []byte(`123`),
			name:       "value",
			expectNull: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := ReadWebhookDefinitionFilterTermStringObject(
				ctx,
				path.Root("test"),
				testcase.name,
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())

				return
			}

			require.False(t, diags.HasError())

			if testcase.expectNull {
				assert.True(t, value.IsNull())
			} else {
				assert.Equal(t, testcase.expectValue, value.ValueString())
			}
		})
	}
}

func TestReadWebhookDefinitionFilterTermsWarnForRepresentableShapeMismatches(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("filter")

	stringValue, stringDiags := ReadWebhookDefinitionFilterTermString(t.Context(), valuePath, []byte(`123`))
	assert.True(t, stringValue.IsNull())
	assert.False(t, stringDiags.HasError())
	assert.Equal(t, []string{valuePath.String()}, webhookWarningPaths(t, stringDiags))
	nullStringValue, nullStringDiags := ReadWebhookDefinitionFilterTermString(t.Context(), valuePath, []byte(`null`))
	assert.True(t, nullStringValue.IsNull())
	assert.False(t, nullStringDiags.HasError())
	assert.Equal(t, []string{valuePath.String()}, webhookWarningPaths(t, nullStringDiags))

	arrayValue, arrayDiags := ReadWebhookDefinitionFilterTermStringArray(t.Context(), valuePath, []byte(`["valid",123]`))
	assert.False(t, arrayDiags.HasError())
	assert.Equal(t, []string{valuePath.AtListIndex(1).String()}, webhookWarningPaths(t, arrayDiags))
	assert.Equal(t, []types.String{types.StringValue("valid"), types.StringNull()}, arrayValue.Elements())
	nullArrayValue, nullArrayDiags := ReadWebhookDefinitionFilterTermStringArray(t.Context(), valuePath, []byte(`["valid",null]`))
	assert.False(t, nullArrayDiags.HasError())
	assert.Equal(t, []string{valuePath.AtListIndex(1).String()}, webhookWarningPaths(t, nullArrayDiags))
	assert.Equal(t, []types.String{types.StringValue("valid"), types.StringNull()}, nullArrayValue.Elements())

	objectValue, objectDiags := ReadWebhookDefinitionFilterTermStringObject(t.Context(), valuePath, "doc", []byte(`{"doc":123}`))
	assert.True(t, objectValue.IsNull())
	assert.False(t, objectDiags.HasError())
	assert.Equal(t, []string{valuePath.String()}, webhookWarningPaths(t, objectDiags))
	nullObjectValue, nullObjectDiags := ReadWebhookDefinitionFilterTermStringObject(t.Context(), valuePath, "doc", []byte(`{"doc":null}`))
	assert.True(t, nullObjectValue.IsNull())
	assert.False(t, nullObjectDiags.HasError())
	assert.Equal(t, []string{valuePath.String()}, webhookWarningPaths(t, nullObjectDiags))
}

func TestReadWebhookDefinitionFilterTermStringArrayDistinguishesNullFromEmpty(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("filter")

	nullValue, nullDiags := ReadWebhookDefinitionFilterTermStringArray(t.Context(), valuePath, []byte(`null`))
	assert.True(t, nullValue.IsNull())
	assert.False(t, nullDiags.HasError())
	assert.Equal(t, []string{valuePath.String()}, webhookWarningPaths(t, nullDiags))

	emptyValue, emptyDiags := ReadWebhookDefinitionFilterTermStringArray(t.Context(), valuePath, []byte(`[]`))
	assert.False(t, emptyValue.IsNull())
	assert.Empty(t, emptyValue.Elements())
	assert.Empty(t, emptyDiags)
}

func TestWebhookMutationStateDoesNotManufactureEqualityFromLossyFallback(t *testing.T) {
	t.Parallel()

	plannedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringNull()),
	})
	plan := WebhookModel{
		Name:    types.StringUnknown(),
		URL:     types.StringUnknown(),
		Filters: plannedFilters,
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		Topics:  NewTypedList([]types.String{}),
	}
	response := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys("space", "webhook"),
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`123`)}},
		}),
	}

	mutationState, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Len(t, responseDiags.Warnings(), 1)
	require.True(t, consistencyDiags.HasError())
	assert.True(t, mutationState.Filters.Equal(plannedFilters))
	assert.Equal(t, []string{"filters[0].equals.value"}, webhookWarningPaths(t, responseDiags))
	assert.Equal(t, []string{"filters"}, attributeDiagnosticPaths(t, consistencyDiags))
}

func TestWebhookMutationStateRejectsRepresentableFilterContradiction(t *testing.T) {
	t.Parallel()

	plannedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
	})
	plan := WebhookModel{
		Name:    types.StringUnknown(),
		URL:     types.StringUnknown(),
		Filters: plannedFilters,
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		Topics:  NewTypedList([]types.String{}),
	}
	response := cm.WebhookDefinition{
		Sys:  cm.NewWebhookDefinitionSys("space", "webhook"),
		Name: "Response webhook",
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Asset"`)}},
		}),
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	require.False(t, mutationStateDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"filters"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "Contentful returned different webhook filters", consistencyDiags.Errors()[0].Summary())
	assert.Equal(t, "Asset", mutationState.Filters.Elements()[0].Value().Equals.Value().Value.ValueString())
	assert.Equal(t, "Response webhook", mutationState.Name.ValueString())
}

func TestWebhookMutationStateRestoresSemanticallyEquivalentReorderedPlan(t *testing.T) {
	t.Parallel()

	plannedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			webhookInValue(types.StringValue("sys.id"), NewTypedList([]types.String{
				types.StringValue("a"), types.StringValue("b"), types.StringValue("a"),
			})),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
		NewTypedObject(webhookFilterValue(
			NewTypedObject(WebhookFilterNotValue{
				Equals: NewTypedObjectNull[WebhookFilterEqualsValue](),
				In: webhookInValue(types.StringValue("sys.environment.sys.id"), NewTypedList([]types.String{
					types.StringValue("master"), types.StringValue("preview"),
				})),
				Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
			}),
			NewTypedObjectNull[WebhookFilterEqualsValue](),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
	})
	response := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys("space", "webhook"),
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.environment.sys.id"}`), []byte(`["preview","master"]`)},
			})},
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)}},
			{In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["a","a","b"]`)}},
		}),
	}
	plan := WebhookModel{
		Name:    types.StringUnknown(),
		URL:     types.StringUnknown(),
		Filters: plannedFilters,
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		Topics:  NewTypedList([]types.String{}),
	}

	mutationState, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, mutationState.Filters.Equal(plannedFilters))
	assert.Equal(t, []types.String{types.StringValue("a"), types.StringValue("b"), types.StringValue("a")}, mutationState.Filters.Elements()[0].Value().In.Value().Values.Elements())
}

func TestWebhookMutationStatePreservesFilterDuplicateMultiplicity(t *testing.T) {
	t.Parallel()

	plannedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
		webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry")),
	})
	response := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys("space", "webhook"),
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)}},
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Asset"`)}},
		}),
	}
	plan := WebhookModel{
		Name:    types.StringUnknown(),
		URL:     types.StringUnknown(),
		Filters: plannedFilters,
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		Topics:  NewTypedList([]types.String{}),
	}

	mutationState, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, "Asset", mutationState.Filters.Elements()[1].Value().Equals.Value().Value.ValueString())
}

func TestWebhookMutationStateUsesResponseForUnknownFilters(t *testing.T) {
	t.Parallel()

	response := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys("space", "webhook"),
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)}},
		}),
	}
	plan := WebhookModel{
		Name:    types.StringUnknown(),
		URL:     types.StringUnknown(),
		Filters: NewTypedListUnknown[TypedObject[WebhookFilterValue]](),
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		Topics:  NewTypedList([]types.String{}),
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.False(t, mutationStateDiags.HasError())
	assert.Empty(t, consistencyDiags)
	assert.False(t, mutationState.Filters.IsUnknown())
	assert.False(t, mutationState.Filters.IsNull())
	assert.Equal(t, "Entry", mutationState.Filters.Elements()[0].Value().Equals.Value().Value.ValueString())

	plan.Filters = NewTypedListNull[TypedObject[WebhookFilterValue]]()
	nullPlanState, nullPlanDiags, nullConsistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)
	assert.False(t, nullPlanDiags.HasError())
	assert.True(t, nullConsistencyDiags.HasError())
	assert.False(t, nullPlanState.Filters.IsNull())
	assert.Equal(t, "Entry", nullPlanState.Filters.Elements()[0].Value().Equals.Value().Value.ValueString())
}
