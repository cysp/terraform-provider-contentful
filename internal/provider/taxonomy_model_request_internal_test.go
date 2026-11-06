package provider

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyLabelMapRequestStates(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	states := map[string]struct {
		value       types.Map
		expected    cm.OptLocalizedStringList
		expectError bool
	}{
		"null":    {value: types.MapNull(listType)},
		"unknown": {value: types.MapUnknown(listType), expectError: true},
		"known empty": {
			value:    types.MapValueMust(listType, map[string]attr.Value{}),
			expected: cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {}}),
		},
		"known populated": {
			value: types.MapValueMust(listType, map[string]attr.Value{
				"en-US": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Label")}),
			}),
			expected: cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {"Label"}}),
		},
	}

	attributes := map[string]struct {
		set func(*TaxonomyConceptModel, types.Map)
		get func(cm.TaxonomyConceptRequest) cm.OptLocalizedStringList
	}{
		"alt_labels": {
			set: func(model *TaxonomyConceptModel, value types.Map) { model.AltLabels = value },
			get: func(request cm.TaxonomyConceptRequest) cm.OptLocalizedStringList { return request.AltLabels },
		},
		"hidden_labels": {
			set: func(model *TaxonomyConceptModel, value types.Map) { model.HiddenLabels = value },
			get: func(request cm.TaxonomyConceptRequest) cm.OptLocalizedStringList { return request.HiddenLabels },
		},
	}

	for attributeName, attribute := range attributes {
		for stateName, state := range states {
			t.Run(attributeName+"/"+stateName, func(t *testing.T) {
				t.Parallel()

				model := taxonomyConceptUpdatePlan()
				attribute.set(&model, state.value)

				prepared, diags := prepareTaxonomyConceptMutation(model, model)
				if state.expectError {
					require.True(t, diags.HasError())
					assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

					return
				}

				require.False(t, diags.HasError(), diags)
				request, requestDiags := prepared.planRequest(t.Context())
				require.False(t, requestDiags.HasError(), requestDiags)

				assert.Equal(t, state.expected, attribute.get(request))
			})
		}
	}
}

func TestTaxonomyConceptListRequestStates(t *testing.T) {
	t.Parallel()

	attributes := map[string]struct {
		set   func(*TaxonomyConceptModel, types.List)
		check func(*testing.T, cm.TaxonomyConceptRequest, taxonomyStringListTestCase)
	}{
		"notations": {
			set: func(model *TaxonomyConceptModel, value types.List) { model.Notations = value },
			check: func(t *testing.T, request cm.TaxonomyConceptRequest, state taxonomyStringListTestCase) {
				t.Helper()

				assert.Equal(t, state.expectedStrings, request.Notations)
			},
		},
		"broader_concept_ids": {
			set: func(model *TaxonomyConceptModel, value types.List) { model.BroaderConceptIDs = value },
			check: func(t *testing.T, request cm.TaxonomyConceptRequest, state taxonomyStringListTestCase) {
				t.Helper()

				assert.Equal(t, state.expectedLinks, request.Broader)
			},
		},
		"related_concept_ids": {
			set: func(model *TaxonomyConceptModel, value types.List) { model.RelatedConceptIDs = value },
			check: func(t *testing.T, request cm.TaxonomyConceptRequest, state taxonomyStringListTestCase) {
				t.Helper()

				assert.Equal(t, state.expectedLinks, request.Related)
			},
		},
	}

	for attributeName, attribute := range attributes {
		for stateName, state := range taxonomyStringListStates() {
			t.Run(attributeName+"/"+stateName, func(t *testing.T) {
				t.Parallel()

				model := taxonomyConceptUpdatePlan()
				attribute.set(&model, state.value)

				prepared, diags := prepareTaxonomyConceptMutation(model, model)
				if state.expectError {
					require.True(t, diags.HasError())
					assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

					return
				}

				require.False(t, diags.HasError(), diags)
				request, requestDiags := prepared.planRequest(t.Context())
				require.False(t, requestDiags.HasError(), requestDiags)

				attribute.check(t, request, state)
			})
		}
	}
}

func TestTaxonomyConceptSchemeListRequestStates(t *testing.T) {
	t.Parallel()

	attributes := map[string]struct {
		set func(*TaxonomyConceptSchemeModel, types.List)
		get func(cm.TaxonomyConceptSchemeRequest) []cm.TaxonomyConceptLink
	}{
		"top_concept_ids": {
			set: func(model *TaxonomyConceptSchemeModel, value types.List) { model.TopConceptIDs = value },
			get: func(request cm.TaxonomyConceptSchemeRequest) []cm.TaxonomyConceptLink { return request.TopConcepts },
		},
		"concept_ids": {
			set: func(model *TaxonomyConceptSchemeModel, value types.List) { model.ConceptIDs = value },
			get: func(request cm.TaxonomyConceptSchemeRequest) []cm.TaxonomyConceptLink { return request.Concepts },
		},
	}

	for attributeName, attribute := range attributes {
		for stateName, state := range taxonomyStringListStates() {
			t.Run(attributeName+"/"+stateName, func(t *testing.T) {
				t.Parallel()

				model := taxonomyConceptSchemeUpdatePlan()
				attribute.set(&model, state.value)

				prepared, diags := prepareTaxonomyConceptSchemeMutation(model, model)
				if state.expectError {
					require.True(t, diags.HasError())
					assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

					return
				}

				require.False(t, diags.HasError(), diags)
				request, requestDiags := prepared.planRequest(t.Context())
				require.False(t, requestDiags.HasError(), requestDiags)

				assert.Equal(t, state.expectedLinks, attribute.get(request))
			})
		}
	}
}

type taxonomyStringListTestCase struct {
	value           types.List
	expectedStrings []string
	expectedLinks   []cm.TaxonomyConceptLink
	expectError     bool
}

func taxonomyStringListStates() map[string]taxonomyStringListTestCase {
	return map[string]taxonomyStringListTestCase{
		"null":    {value: types.ListNull(types.StringType)},
		"unknown": {value: types.ListUnknown(types.StringType), expectError: true},
		"known empty": {
			value:           types.ListValueMust(types.StringType, []attr.Value{}),
			expectedStrings: []string{},
			expectedLinks:   []cm.TaxonomyConceptLink{},
		},
		"known populated": {
			value:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value")}),
			expectedStrings: []string{"value"},
			expectedLinks: []cm.TaxonomyConceptLink{{
				Sys: cm.TaxonomyConceptLinkSys{
					Type:     cm.TaxonomyConceptLinkSysTypeLink,
					LinkType: cm.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept,
					ID:       "value",
				},
			}},
		},
	}
}

func TestTaxonomyRequestConversionsReportUnknownCollectionElements(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*testing.T) diag.Diagnostics{
		"concept string map": func(t *testing.T) diag.Diagnostics {
			t.Helper()

			model := taxonomyConceptUpdatePlan()
			model.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{
				"en-US": types.StringUnknown(),
			})
			_, diags := model.ToRequest(t.Context())

			return diags
		},
		"concept string list map": func(t *testing.T) diag.Diagnostics {
			t.Helper()

			model := taxonomyConceptUpdatePlan()
			model.AltLabels = types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
				"en-US": types.ListUnknown(types.StringType),
			})
			_, diags := model.ToRequest(t.Context())

			return diags
		},
		"concept string list": func(t *testing.T) diag.Diagnostics {
			t.Helper()

			model := taxonomyConceptUpdatePlan()
			model.Notations = types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()})
			_, diags := model.ToRequest(t.Context())

			return diags
		},
		"concept scheme string map": func(t *testing.T) diag.Diagnostics {
			t.Helper()

			model := taxonomyConceptSchemeUpdatePlan()
			model.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{
				"en-US": types.StringUnknown(),
			})
			_, diags := model.ToRequest(t.Context())

			return diags
		},
		"concept scheme string list": func(t *testing.T) diag.Diagnostics {
			t.Helper()

			model := taxonomyConceptSchemeUpdatePlan()
			model.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()})
			_, diags := model.ToRequest(t.Context())

			return diags
		},
	}

	for name, convert := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.True(t, convert(t).HasError())
		})
	}
}

func TestTaxonomyConceptToRequestAddsPreferredLocalesToLabelMaps(t *testing.T) {
	t.Parallel()

	model := TaxonomyConceptModel{
		PrefLabel:         types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")}),
		AltLabels:         types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Seat")})}),
		HiddenLabels:      types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{}),
		Notations:         types.ListNull(types.StringType),
		BroaderConceptIDs: types.ListNull(types.StringType),
		RelatedConceptIDs: types.ListNull(types.StringType),
	}

	request, diags := model.ToRequest(t.Context())
	require.False(t, diags.HasError())

	altLabels, ok := request.AltLabels.Get()
	require.True(t, ok)
	assert.Equal(t, cm.LocalizedStringList{"en-GB": {"Seat"}, "en-US": {}}, altLabels)

	hiddenLabels, ok := request.HiddenLabels.Get()
	require.True(t, ok)
	assert.Equal(t, cm.LocalizedStringList{"en-US": {}}, hiddenLabels)
}

func TestTaxonomyConceptToRequestOmitsUnconfiguredLabelMaps(t *testing.T) {
	t.Parallel()

	model := TaxonomyConceptModel{
		PrefLabel:         types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")}),
		AltLabels:         types.MapNull(types.ListType{ElemType: types.StringType}),
		HiddenLabels:      types.MapNull(types.ListType{ElemType: types.StringType}),
		Notations:         types.ListNull(types.StringType),
		BroaderConceptIDs: types.ListNull(types.StringType),
		RelatedConceptIDs: types.ListNull(types.StringType),
	}

	request, diags := model.ToRequest(t.Context())
	require.False(t, diags.HasError())
	assert.False(t, request.AltLabels.IsSet())
	assert.False(t, request.HiddenLabels.IsSet())
}

func TestTaxonomyConceptToRequestRejectsInvalidLabelChildrenWithoutPanicking(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}

	for _, attributeName := range []string{"alt_labels", "hidden_labels"} {
		for valueName, invalidValue := range map[string]attr.Value{
			"null":    types.StringNull(),
			"unknown": types.StringUnknown(),
		} {
			t.Run(attributeName+" "+valueName, func(t *testing.T) {
				t.Parallel()

				invalidLabels := types.MapValueMust(listType, map[string]attr.Value{
					"en-US": types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("known"),
						invalidValue,
					}),
				})
				model := TaxonomyConceptModel{
					PrefLabel:         types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")}),
					AltLabels:         types.MapNull(listType),
					HiddenLabels:      types.MapNull(listType),
					Notations:         types.ListNull(types.StringType),
					BroaderConceptIDs: types.ListNull(types.StringType),
					RelatedConceptIDs: types.ListNull(types.StringType),
				}

				switch attributeName {
				case "alt_labels":
					model.AltLabels = invalidLabels
				case "hidden_labels":
					model.HiddenLabels = invalidLabels
				}

				request, diags := model.ToRequest(t.Context())

				assert.Equal(t, cm.TaxonomyConceptRequest{}, request)
				require.True(t, diags.HasError())
				require.Len(t, diags.Errors(), 1)

				pathDiagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
				require.True(t, ok)
				assert.Equal(t, path.Root(attributeName).AtMapKey("en-US").AtListIndex(1), pathDiagnostic.Path())
			})
		}
	}
}

func TestLocalizedStringConversions(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	diags := diag.Diagnostics{}
	null := nullableLocalizedString(types.MapNull(types.StringType), path.Root("definition"), &diags)
	require.False(t, diags.HasError())
	assert.True(t, null.IsNull())
	assert.True(t, localizedStringValue(ctx, null, &diags).IsNull())

	configured := types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Description")})
	known := nullableLocalizedString(configured, path.Root("definition"), &diags)
	require.False(t, diags.HasError())

	value, ok := known.Get()
	require.True(t, ok)
	assert.Equal(t, cm.NullableLocalizedString{"en-US": "Description"}, value)
	assert.Equal(t, configured, localizedStringValue(ctx, known, &diags))
}

func TestTaxonomyURIRequestValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		value         types.String
		expected      cm.OptNilString
		expectedJSON  string
		expectedError bool
	}{
		{
			name:         "null sends explicit null",
			value:        types.StringNull(),
			expected:     cm.NewOptNilStringNull(),
			expectedJSON: "null",
		},
		{
			name:         "known empty sends empty string",
			value:        types.StringValue(""),
			expected:     cm.NewOptNilString(""),
			expectedJSON: testJSON(""),
		},
		{
			name:         "known sends string",
			value:        types.StringValue("https://example.com/taxonomy"),
			expected:     cm.NewOptNilString("https://example.com/taxonomy"),
			expectedJSON: testJSON("https://example.com/taxonomy"),
		},
		{
			name:          "unknown fails closed",
			value:         types.StringUnknown(),
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			t.Run("concept", func(t *testing.T) {
				t.Parallel()

				model := taxonomyConceptUpdatePlan()
				model.URI = test.value
				request, diags := model.ToRequest(t.Context())

				if test.expectedError {
					assert.Equal(t, cm.TaxonomyConceptRequest{}, request)
					require.True(t, diags.HasError())
					require.Len(t, diags.Errors(), 1)
					assertTaxonomyDiagnosticPath(t, diags, path.Root("uri"))

					return
				}

				require.False(t, diags.HasError(), diags.Errors())
				assertTaxonomyURIWireValue(t, request.URI, test.expected, test.expectedJSON)
			})

			t.Run("concept scheme", func(t *testing.T) {
				t.Parallel()

				model := taxonomyConceptSchemeUpdatePlan()
				model.URI = test.value
				request, diags := model.ToRequest(t.Context())

				if test.expectedError {
					assert.Equal(t, cm.TaxonomyConceptSchemeRequest{}, request)
					require.True(t, diags.HasError())
					require.Len(t, diags.Errors(), 1)
					assertTaxonomyDiagnosticPath(t, diags, path.Root("uri"))

					return
				}

				require.False(t, diags.HasError(), diags.Errors())
				assertTaxonomyURIWireValue(t, request.URI, test.expected, test.expectedJSON)
			})
		})
	}
}

func assertTaxonomyURIWireValue(
	t *testing.T,
	actual cm.OptNilString,
	expected cm.OptNilString,
	expectedJSON string,
) {
	t.Helper()

	assert.Equal(t, expected, actual)

	payload, err := json.Marshal(actual)
	require.NoError(t, err)
	assert.JSONEq(t, expectedJSON, string(payload))
}

func TestTaxonomyToRequestRejectsUnknownOptionalValues(t *testing.T) {
	t.Parallel()

	model := TaxonomyConceptModel{
		URI:               types.StringUnknown(),
		PrefLabel:         types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")}),
		AltLabels:         types.MapNull(types.ListType{ElemType: types.StringType}),
		HiddenLabels:      types.MapNull(types.ListType{ElemType: types.StringType}),
		Notations:         types.ListNull(types.StringType),
		Note:              types.MapUnknown(types.StringType),
		BroaderConceptIDs: types.ListNull(types.StringType),
		RelatedConceptIDs: types.ListNull(types.StringType),
	}

	_, diags := model.ToRequest(t.Context())
	require.True(t, diags.HasError())

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	assert.Contains(t, paths, "uri")
	assert.Contains(t, paths, "note")
}

func TestOptionalComputedStringListValuePreservesChildPathAndFailsClosed(t *testing.T) {
	t.Parallel()

	value := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("known"),
		types.StringNull(),
	})

	actual, diags := optionalComputedStringListValue(
		value,
		path.Root("broader_concept_ids"),
	)

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)

	pathDiagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, path.Root("broader_concept_ids").AtListIndex(1), pathDiagnostic.Path())
}
