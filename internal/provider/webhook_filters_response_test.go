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

func TestWebhookMutationStateReconcilesKnownPlannedFilters(t *testing.T) {
	t.Parallel()

	plannedFilters := NewTypedList([]TypedObject[WebhookFilterValue]{
		NewTypedObject(webhookFilterValue(
			NewTypedObjectNull[WebhookFilterNotValue](),
			webhookEqualsValue(types.StringValue("sys.type"), types.StringValue("Entry")),
			NewTypedObjectNull[WebhookFilterInValue](),
			NewTypedObjectNull[WebhookFilterRegexpValue](),
		)),
	})
	plan := WebhookModel{
		Filters: plannedFilters,
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
	}
	response := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys("space", "webhook"),
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
			{Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`123`)}},
		}),
	}

	mutationState, mutationStateDiags := NewWebhookResourceModelForMutationState(t.Context(), response, plan)
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 1)
	assert.True(t, mutationState.Filters.Equal(plannedFilters))

	readState, readDiags := NewWebhookResourceModelFromResponse(t.Context(), response, plan.Headers.Elements())
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 1)
	assert.True(t, readState.Filters.Elements()[0].Value().Equals.Value().Value.IsNull())
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
		Filters: NewTypedListUnknown[TypedObject[WebhookFilterValue]](),
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
	}

	mutationState, mutationStateDiags := NewWebhookResourceModelForMutationState(t.Context(), response, plan)

	assert.False(t, mutationStateDiags.HasError())
	assert.False(t, mutationState.Filters.IsUnknown())
	assert.False(t, mutationState.Filters.IsNull())
	assert.Equal(t, "Entry", mutationState.Filters.Elements()[0].Value().Equals.Value().Value.ValueString())

	plan.Filters = NewTypedListNull[TypedObject[WebhookFilterValue]]()
	nullPlanState, nullPlanDiags := NewWebhookResourceModelForMutationState(t.Context(), response, plan)
	assert.False(t, nullPlanDiags.HasError())
	assert.True(t, nullPlanState.Filters.IsNull())
}
