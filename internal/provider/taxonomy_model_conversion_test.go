//nolint:testpackage
package provider

import (
	"encoding/json"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyPatch(t *testing.T) {
	t.Parallel()

	current := cm.TaxonomyConceptSchemeRequest{
		URI:         cm.NewOptNilPointerString(nil),
		PrefLabel:   cm.LocalizedString{"en-US": "Products"},
		Definition:  nullLocalizedString(),
		TopConcepts: []cm.TaxonomyConceptLink{},
		Concepts:    []cm.TaxonomyConceptLink{},
	}

	unchanged, err := taxonomyPatch(current, current)
	require.NoError(t, err)
	require.Empty(t, unchanged)

	desired := current
	desired.PrefLabel = cm.LocalizedString{"en-US": "Product catalog"}

	patch, err := taxonomyPatch(current, desired)
	require.NoError(t, err)
	require.Len(t, patch, 1)
	assert.Equal(t, cm.TaxonomyPatchItemOpAdd, patch[0].Op)
	assert.Equal(t, "/prefLabel", patch[0].Path)
	assert.JSONEq(t, `{"en-US":"Product catalog"}`, string(patch[0].Value))
}

func TestTaxonomyPatchComparesJSONMembersStructurally(t *testing.T) {
	t.Parallel()

	current := map[string]json.RawMessage{
		"prefLabel": json.RawMessage(`{"en-US":"Chair","fr-FR":"Chaise"}`),
	}
	desired := map[string]json.RawMessage{
		"prefLabel": json.RawMessage(`{"fr-FR":"Chaise","en-US":"Chair"}`),
	}

	patch, err := taxonomyPatch(current, desired)
	require.NoError(t, err)
	assert.Empty(t, patch)
}

func TestTaxonomyPatchAddsAbsentDesiredMember(t *testing.T) {
	t.Parallel()

	patch, err := taxonomyPatch(map[string]json.RawMessage{}, map[string]json.RawMessage{
		"altLabels": json.RawMessage(`{"en-US":[]}`),
	})
	require.NoError(t, err)
	require.Len(t, patch, 1)
	assert.Equal(t, "/altLabels", patch[0].Path)
	assert.JSONEq(t, `{"en-US":[]}`, string(patch[0].Value))
}

func TestTaxonomyPatchDoesNotChangeUnchangedMultiLocaleRequest(t *testing.T) {
	t.Parallel()

	request := cm.TaxonomyConceptRequest{
		PrefLabel: cm.LocalizedString{"en-US": "Chair", "fr-FR": "Chaise"},
		AltLabels: cm.NewOptLocalizedStringList(cm.LocalizedStringList{
			"en-US": {"Seat"},
			"fr-FR": {"Siege"},
		}),
		HiddenLabels: cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
	}

	for range 100 {
		patch, err := taxonomyPatch(request, request)
		require.NoError(t, err)
		assert.Empty(t, patch)
	}
}

func TestTaxonomyPatchSortsChangedFields(t *testing.T) {
	t.Parallel()

	current := cm.TaxonomyConceptRequest{
		URI:           cm.NewOptNilPointerString(nil),
		PrefLabel:     cm.LocalizedString{"en-US": "Chair"},
		AltLabels:     cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		HiddenLabels:  cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		Notations:     []string{"CHAIR"},
		Broader:       []cm.TaxonomyConceptLink{},
		Related:       []cm.TaxonomyConceptLink{},
		Note:          nullLocalizedString(),
		ChangeNote:    nullLocalizedString(),
		Definition:    nullLocalizedString(),
		EditorialNote: nullLocalizedString(),
		Example:       nullLocalizedString(),
		HistoryNote:   nullLocalizedString(),
		ScopeNote:     nullLocalizedString(),
	}
	desired := current
	desired.Notations = []string{"SEAT"}
	desired.PrefLabel = cm.LocalizedString{"en-US": "Seat"}

	patch, err := taxonomyPatch(current, desired)
	require.NoError(t, err)
	require.Len(t, patch, 2)
	assert.Equal(t, "/notations", patch[0].Path)
	assert.Equal(t, "/prefLabel", patch[1].Path)
}

func TestTaxonomyPatchValueShapes(t *testing.T) {
	t.Parallel()

	current := completeTaxonomyConceptRequest()
	tests := map[string]struct {
		mutate func(*cm.TaxonomyConceptRequest)
		path   string
		value  string
	}{
		"nullable value to null": {
			mutate: func(desired *cm.TaxonomyConceptRequest) { desired.Definition = nullLocalizedString() },
			path:   "/definition", value: `null`,
		},
		"primitive list to empty": {
			mutate: func(desired *cm.TaxonomyConceptRequest) { desired.Notations = []string{} },
			path:   "/notations", value: `[]`,
		},
		"localized list map to empty": {
			mutate: func(desired *cm.TaxonomyConceptRequest) {
				desired.AltLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{})
			},
			path: "/altLabels", value: `{}`,
		},
		"relationship replacement": {
			mutate: func(desired *cm.TaxonomyConceptRequest) {
				desired.Broader = []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("replacement")}
			},
			path: "/broader", value: `[{
				"sys":{"type":"Link","linkType":"TaxonomyConcept","id":"replacement"}
			}]`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			desired := cloneTaxonomyConceptRequest(t, current)
			test.mutate(&desired)
			desiredBeforePatch := cloneTaxonomyConceptRequest(t, desired)

			patch, err := taxonomyPatch(current, desired)
			require.NoError(t, err)
			require.Len(t, patch, 1)
			assert.Equal(t, cm.TaxonomyPatchItemOpAdd, patch[0].Op)
			assert.Equal(t, test.path, patch[0].Path)
			assert.JSONEq(t, test.value, string(patch[0].Value))

			assert.Equal(t, completeTaxonomyConceptRequest(), current)
			assert.Equal(t, desiredBeforePatch, desired)
		})
	}
}

func completeTaxonomyConceptRequest() cm.TaxonomyConceptRequest {
	return cm.TaxonomyConceptRequest{
		URI:           cm.NewOptNilPointerString(new("https://example.com/concepts/chair")),
		PrefLabel:     cm.LocalizedString{"en-US": "Chair"},
		AltLabels:     cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {"Seat"}}),
		HiddenLabels:  cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		Notations:     []string{"CHAIR"},
		Note:          nullLocalizedString(),
		ChangeNote:    nullLocalizedString(),
		Definition:    cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "A seat"}),
		EditorialNote: nullLocalizedString(),
		Example:       nullLocalizedString(),
		HistoryNote:   nullLocalizedString(),
		ScopeNote:     nullLocalizedString(),
		Broader:       []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("furniture")},
		Related:       []cm.TaxonomyConceptLink{},
	}
}

func cloneTaxonomyConceptRequest(t *testing.T, request cm.TaxonomyConceptRequest) cm.TaxonomyConceptRequest {
	t.Helper()

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var result cm.TaxonomyConceptRequest
	require.NoError(t, json.Unmarshal(data, &result))

	return result
}

func nullLocalizedString() cm.OptNilNullableLocalizedString {
	var value cm.OptNilNullableLocalizedString
	value.SetToNull()

	return value
}

func TestNewTaxonomyConceptRefreshStateProjectsLabelMapDrift(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	prior := TaxonomyConceptModel{
		AltLabels: types.MapValueMust(listType, map[string]attr.Value{
			"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chair")}),
			"de-DE": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Stuhl")}),
		}),
		HiddenLabels: types.MapValueMust(listType, map[string]attr.Value{
			"en-US": types.ListValueMust(types.StringType, nil),
		}),
	}
	remote := taxonomyConceptResponseWithLabelMaps(
		map[string][]string{"de-DE": {}, "en-GB": {"Chair"}, "en-US": {}, "fr-FR": {"Chaise"}},
		map[string][]string{},
	)

	actual, diags := newTaxonomyConceptRefreshState(t.Context(), prior, remote)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]attr.Value{
		"de-DE": types.ListValueMust(types.StringType, []attr.Value{}),
		"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chair")}),
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chaise")}),
	}, actual.AltLabels.Elements())
	assert.Equal(t, map[string]attr.Value{}, actual.HiddenLabels.Elements())
}

func TestNewTaxonomyConceptRefreshStateUsesCompleteRemoteMapWithoutPriorValue(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	prior := TaxonomyConceptModel{
		AltLabels:    types.MapNull(listType),
		HiddenLabels: types.MapUnknown(listType),
	}
	remote := taxonomyConceptResponseWithLabelMaps(
		map[string][]string{"en-US": {}, "fr-FR": {"Chaise"}},
		map[string][]string{"en-US": {}},
	)

	actual, diags := newTaxonomyConceptRefreshState(t.Context(), prior, remote)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chaise")}),
	}, actual.AltLabels.Elements())
	assert.Equal(t, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
	}, actual.HiddenLabels.Elements())
}

func taxonomyConceptResponseWithLabelMaps(altLabels, hiddenLabels map[string][]string) cm.TaxonomyConcept {
	return cm.TaxonomyConcept{
		Sys: cm.TaxonomyConceptSys{
			Organization: cm.OrganizationLink{Sys: cm.OrganizationLinkSys{ID: "organization-id"}},
			ID:           "concept-id",
			Version:      1,
		},
		PrefLabel:      cm.LocalizedString{"en-US": "Chair"},
		AltLabels:      cm.NewOptLocalizedStringList(cm.LocalizedStringList(altLabels)),
		HiddenLabels:   cm.NewOptLocalizedStringList(cm.LocalizedStringList(hiddenLabels)),
		Notations:      []string{},
		Broader:        []cm.TaxonomyConceptLink{},
		Related:        []cm.TaxonomyConceptLink{},
		ConceptSchemes: []cm.TaxonomyConceptSchemeLink{},
	}
}

func TestNewTaxonomyConceptModelFromResponsePreservesLabelMapPresence(t *testing.T) {
	t.Parallel()

	unset := taxonomyConceptResponseWithLabelMaps(nil, nil)
	unset.AltLabels.Reset()
	unset.HiddenLabels.Reset()
	unsetModel, unsetDiags := NewTaxonomyConceptModelFromResponse(t.Context(), unset)
	require.False(t, unsetDiags.HasError())
	assert.True(t, unsetModel.AltLabels.IsNull())
	assert.True(t, unsetModel.HiddenLabels.IsNull())

	empty := taxonomyConceptResponseWithLabelMaps(map[string][]string{}, map[string][]string{})
	emptyModel, emptyDiags := NewTaxonomyConceptModelFromResponse(t.Context(), empty)
	require.False(t, emptyDiags.HasError())
	assert.False(t, emptyModel.AltLabels.IsNull() || emptyModel.AltLabels.IsUnknown())
	assert.False(t, emptyModel.HiddenLabels.IsNull() || emptyModel.HiddenLabels.IsUnknown())
	assert.Empty(t, emptyModel.AltLabels.Elements())
	assert.Empty(t, emptyModel.HiddenLabels.Elements())
}

func TestTaxonomyLabelMapRequestStates(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	states := []struct {
		name  string
		value types.Map
	}{
		{name: "null", value: types.MapNull(listType)},
		{name: "unknown", value: types.MapUnknown(listType)},
		{name: "known empty", value: types.MapValueMust(listType, map[string]attr.Value{})},
		{name: "known populated", value: types.MapValueMust(listType, map[string]attr.Value{
			"en-US": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Label")}),
		})},
	}

	attributes := []struct {
		name string
		set  func(*TaxonomyConceptModel, types.Map)
		get  func(cm.TaxonomyConceptRequest) cm.OptLocalizedStringList
	}{
		{
			name: "alt_labels",
			set:  func(model *TaxonomyConceptModel, value types.Map) { model.AltLabels = value },
			get:  func(request cm.TaxonomyConceptRequest) cm.OptLocalizedStringList { return request.AltLabels },
		},
		{
			name: "hidden_labels",
			set:  func(model *TaxonomyConceptModel, value types.Map) { model.HiddenLabels = value },
			get:  func(request cm.TaxonomyConceptRequest) cm.OptLocalizedStringList { return request.HiddenLabels },
		},
	}

	for _, attribute := range attributes {
		t.Run(attribute.name, func(t *testing.T) {
			t.Parallel()

			for stateIndex, state := range states {
				t.Run(state.name, func(t *testing.T) {
					t.Parallel()

					assertTaxonomyLabelMapRequestState(t, attribute.name, attribute.set, attribute.get, stateIndex, state.value)
				})
			}
		})
	}
}

func assertTaxonomyLabelMapRequestState(
	t *testing.T,
	attributeName string,
	set func(*TaxonomyConceptModel, types.Map),
	get func(cm.TaxonomyConceptRequest) cm.OptLocalizedStringList,
	stateIndex int,
	value types.Map,
) {
	t.Helper()

	model := taxonomyConceptUpdatePlan()
	set(&model, value)

	prepared, diags := prepareTaxonomyConceptMutation(t.Context(), model, model)
	if stateIndex == 1 {
		require.True(t, diags.HasError())
		assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

		return
	}

	require.False(t, diags.HasError())

	labels := get(prepared.CreateRequest())
	if stateIndex == 0 {
		assert.False(t, labels.IsSet())

		return
	}

	actual, ok := labels.Get()
	require.True(t, ok)

	if stateIndex == 2 {
		assert.Equal(t, cm.LocalizedStringList{"en-US": {}}, actual)

		return
	}

	assert.Equal(t, cm.LocalizedStringList{"en-US": {"Label"}}, actual)
}

func TestTaxonomyConceptListRequestStates(t *testing.T) {
	t.Parallel()

	for _, attributeName := range []string{"notations", "broader_concept_ids", "related_concept_ids"} {
		t.Run(attributeName, func(t *testing.T) {
			t.Parallel()

			for stateIndex, value := range taxonomyStringListStates("value") {
				t.Run([]string{"null", "unknown", "known empty", "known populated"}[stateIndex], func(t *testing.T) {
					t.Parallel()

					model := taxonomyConceptUpdatePlan()

					switch attributeName {
					case "notations":
						model.Notations = value
					case "broader_concept_ids":
						model.BroaderConceptIDs = value
					case "related_concept_ids":
						model.RelatedConceptIDs = value
					}

					prepared, diags := prepareTaxonomyConceptMutation(t.Context(), model, model)
					if stateIndex == 1 {
						require.True(t, diags.HasError())
						assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

						return
					}

					require.False(t, diags.HasError())

					request := prepared.CreateRequest()
					actual := request.Notations

					switch attributeName {
					case "broader_concept_ids":
						actual = taxonomyConceptLinkIDsPreservingNil(request.Broader)
					case "related_concept_ids":
						actual = taxonomyConceptLinkIDsPreservingNil(request.Related)
					}

					switch stateIndex {
					case 0:
						assert.Nil(t, actual)
					case 2:
						assert.NotNil(t, actual)
						assert.Empty(t, actual)
					default:
						assert.Equal(t, []string{"value"}, actual)
					}
				})
			}
		})
	}
}

func taxonomyConceptLinkIDsPreservingNil(links []cm.TaxonomyConceptLink) []string {
	if links == nil {
		return nil
	}

	return conceptLinkIDs(links)
}

func TestTaxonomyConceptSchemeListRequestStates(t *testing.T) {
	t.Parallel()

	for _, attributeName := range []string{"top_concept_ids", "concept_ids"} {
		t.Run(attributeName, func(t *testing.T) {
			t.Parallel()

			for stateIndex, value := range taxonomyStringListStates("value") {
				t.Run([]string{"null", "unknown", "known empty", "known populated"}[stateIndex], func(t *testing.T) {
					t.Parallel()

					model := taxonomyConceptSchemeUpdatePlan()
					if attributeName == "top_concept_ids" {
						model.TopConceptIDs = value
					} else {
						model.ConceptIDs = value
					}

					prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), model, model)
					if stateIndex == 1 {
						require.True(t, diags.HasError())
						assertTaxonomyDiagnosticPath(t, diags, path.Root(attributeName))

						return
					}

					require.False(t, diags.HasError())

					request := prepared.CreateRequest()

					actual := conceptLinkIDs(request.TopConcepts)
					if attributeName == "concept_ids" {
						actual = conceptLinkIDs(request.Concepts)
					}

					if stateIndex < 3 {
						assert.NotNil(t, actual)
						assert.Empty(t, actual)
					} else {
						assert.Equal(t, []string{"value"}, actual)
					}
				})
			}
		})
	}
}

func taxonomyStringListStates(populated string) []types.List {
	return []types.List{
		types.ListNull(types.StringType),
		types.ListUnknown(types.StringType),
		types.ListValueMust(types.StringType, []attr.Value{}),
		types.ListValueMust(types.StringType, []attr.Value{types.StringValue(populated)}),
	}
}

func assertTaxonomyDiagnosticPath(t *testing.T, diags diag.Diagnostics, want path.Path) {
	t.Helper()

	for _, diagnostic := range diags.Errors() {
		pathDiagnostic, ok := diagnostic.(diag.DiagnosticWithPath)
		if ok && pathDiagnostic.Path().String() == want.String() {
			return
		}
	}

	t.Fatalf("diagnostics did not include path %s: %v", want, diags)
}

func TestPreparedTaxonomyConceptMutationSendsKnownPlanValuesWhenConfigOmitsCollections(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	config := taxonomyConceptUpdatePlan()
	plan := config
	plan.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chaise")}),
	})
	plan.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Siege")}),
	})
	plan.Notations = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("CHAIR")})
	plan.BroaderConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("parent")})
	plan.RelatedConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("related")})

	prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	request := prepared.CreateRequest()

	assert.True(t, request.AltLabels.IsSet())
	assert.True(t, request.HiddenLabels.IsSet())
	assert.Equal(t, []string{"CHAIR"}, request.Notations)
	assert.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent")}, request.Broader)
	assert.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("related")}, request.Related)
}

func TestPreparedTaxonomyConceptSchemeMutationSendsKnownPlanValuesWhenConfigOmitsCollections(t *testing.T) {
	t.Parallel()

	config := taxonomyConceptSchemeUpdatePlan()
	plan := config
	plan.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("top")})
	plan.ConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("concept")})

	prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	request := prepared.CreateRequest()
	assert.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("top")}, request.TopConcepts)
	assert.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("concept")}, request.Concepts)
}

func TestPreparedTaxonomyConceptMutationRejectsUnknownCollectionConfiguration(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	tests := []struct {
		name string
		path path.Path
		set  func(*TaxonomyConceptModel, bool)
	}{
		{"alt labels", path.Root("alt_labels"), func(model *TaxonomyConceptModel, unknown bool) {
			if unknown {
				model.AltLabels = types.MapUnknown(listType)
			} else {
				model.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
			}
		}},
		{"hidden labels", path.Root("hidden_labels"), func(model *TaxonomyConceptModel, unknown bool) {
			if unknown {
				model.HiddenLabels = types.MapUnknown(listType)
			} else {
				model.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{})
			}
		}},
		{"notations", path.Root("notations"), func(model *TaxonomyConceptModel, unknown bool) {
			if unknown {
				model.Notations = types.ListUnknown(types.StringType)
			} else {
				model.Notations = types.ListValueMust(types.StringType, []attr.Value{})
			}
		}},
		{"broader concepts", path.Root("broader_concept_ids"), func(model *TaxonomyConceptModel, unknown bool) {
			if unknown {
				model.BroaderConceptIDs = types.ListUnknown(types.StringType)
			} else {
				model.BroaderConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}
		}},
		{"related concepts", path.Root("related_concept_ids"), func(model *TaxonomyConceptModel, unknown bool) {
			if unknown {
				model.RelatedConceptIDs = types.ListUnknown(types.StringType)
			} else {
				model.RelatedConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}
		}},
	}

	for _, test := range tests {
		for _, planUnknown := range []bool{false, true} {
			t.Run(test.name+"/plan unknown="+strconv.FormatBool(planUnknown), func(t *testing.T) {
				t.Parallel()

				config, plan := taxonomyConceptUpdatePlan(), taxonomyConceptUpdatePlan()
				test.set(&config, true)
				test.set(&plan, planUnknown)
				_, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
				require.True(t, diags.HasError())
				require.Len(t, diags.Errors(), 1)
				assertTaxonomyDiagnosticPath(t, diags, test.path)
			})
		}
	}
}

func TestPreparedTaxonomyConceptSchemeMutationRejectsUnknownCollectionConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path path.Path
		set  func(*TaxonomyConceptSchemeModel, bool)
	}{
		{"top concepts", path.Root("top_concept_ids"), func(model *TaxonomyConceptSchemeModel, unknown bool) {
			if unknown {
				model.TopConceptIDs = types.ListUnknown(types.StringType)
			} else {
				model.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}
		}},
		{"concepts", path.Root("concept_ids"), func(model *TaxonomyConceptSchemeModel, unknown bool) {
			if unknown {
				model.ConceptIDs = types.ListUnknown(types.StringType)
			} else {
				model.ConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}
		}},
	}

	for _, test := range tests {
		for _, planUnknown := range []bool{false, true} {
			t.Run(test.name+"/plan unknown="+strconv.FormatBool(planUnknown), func(t *testing.T) {
				t.Parallel()

				config, plan := taxonomyConceptSchemeUpdatePlan(), taxonomyConceptSchemeUpdatePlan()
				test.set(&config, true)
				test.set(&plan, planUnknown)
				_, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
				require.True(t, diags.HasError())
				require.Len(t, diags.Errors(), 1)
				assertTaxonomyDiagnosticPath(t, diags, test.path)
			})
		}
	}
}

func TestPreparedTaxonomyConceptMutationAllowsUnknownComputedPlanWhenConfigOmitted(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	config := taxonomyConceptUpdatePlan()
	plan := config
	plan.AltLabels = types.MapUnknown(listType)
	plan.HiddenLabels = types.MapUnknown(listType)
	plan.Notations = types.ListUnknown(types.StringType)
	plan.BroaderConceptIDs = types.ListUnknown(types.StringType)
	plan.RelatedConceptIDs = types.ListUnknown(types.StringType)

	prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	request := prepared.CreateRequest()

	assert.False(t, request.AltLabels.IsSet())
	assert.False(t, request.HiddenLabels.IsSet())
	assert.Nil(t, request.Notations)
	assert.Nil(t, request.Broader)
	assert.Nil(t, request.Related)
}

func TestPreparedTaxonomyConceptPatchFollowsCollectionOwnership(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	state := taxonomyConceptUpdatePlan()
	state.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chaise")}),
	})
	state.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Siege")}),
	})
	state.Notations = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("CHAIR")})
	state.BroaderConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("parent")})
	state.RelatedConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("related")})

	t.Run("omitted collections remain response-owned", func(t *testing.T) {
		t.Parallel()

		patchState := state
		config := state
		plan := state
		plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Updated")})
		config.PrefLabel = plan.PrefLabel
		config.AltLabels = types.MapNull(listType)
		config.HiddenLabels = types.MapNull(listType)
		config.Notations = types.ListNull(types.StringType)
		config.BroaderConceptIDs = types.ListNull(types.StringType)
		config.RelatedConceptIDs = types.ListNull(types.StringType)
		plan.AltLabels = types.MapUnknown(listType)
		plan.HiddenLabels = types.MapUnknown(listType)
		plan.Notations = types.ListUnknown(types.StringType)
		plan.BroaderConceptIDs = types.ListUnknown(types.StringType)
		plan.RelatedConceptIDs = types.ListUnknown(types.StringType)

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError(), patchDiags)
		require.Len(t, patch, 1)
		assert.Equal(t, "/prefLabel", patch[0].Path)
	})

	t.Run("unchanged owned sparse labels do not create spurious patches", func(t *testing.T) {
		t.Parallel()

		patchState := state
		config := state
		plan := state
		plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Updated")})
		config.PrefLabel = plan.PrefLabel
		config.AltLabels = state.AltLabels
		config.HiddenLabels = state.HiddenLabels
		plan.AltLabels = state.AltLabels
		plan.HiddenLabels = state.HiddenLabels

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError())
		require.Len(t, patch, 1)
		assert.Equal(t, "/prefLabel", patch[0].Path)
	})

	t.Run("known empty collections take ownership and clear remote values", func(t *testing.T) {
		t.Parallel()

		patchState := state
		config := state
		plan := state
		config.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
		config.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{})
		config.Notations = types.ListValueMust(types.StringType, []attr.Value{})
		config.BroaderConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
		config.RelatedConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
		plan.AltLabels = config.AltLabels
		plan.HiddenLabels = config.HiddenLabels
		plan.Notations = config.Notations
		plan.BroaderConceptIDs = config.BroaderConceptIDs
		plan.RelatedConceptIDs = config.RelatedConceptIDs

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError())
		assert.ElementsMatch(t, []string{"/altLabels", "/broader", "/hiddenLabels", "/notations", "/related"}, taxonomyPatchPaths(patch))
	})

	t.Run("owned preferred-empty label is added when state omits it", func(t *testing.T) {
		t.Parallel()

		patchState := state
		config := state
		plan := state
		config.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
			"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
		})
		plan.AltLabels = config.AltLabels
		patchState.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError())
		assert.Contains(t, taxonomyPatchPaths(patch), "/altLabels")
	})

	t.Run("response-added preferred-empty label is not removed", func(t *testing.T) {
		t.Parallel()

		patchState := state
		config := state
		plan := state
		config.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
		plan.AltLabels = config.AltLabels
		patchState.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
			"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
		})

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError())
		assert.NotContains(t, taxonomyPatchPaths(patch), "/altLabels")
	})

	t.Run("preferred locale transition updates owned empty label wire maps", func(t *testing.T) {
		t.Parallel()

		patchState := state
		patchState.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")})
		patchState.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
		patchState.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{})
		config := patchState
		config.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
		config.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{})
		plan := config
		plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"fr-FR": types.StringValue("Chaise")})
		config.PrefLabel = plan.PrefLabel

		prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), patchState)
		require.False(t, patchDiags.HasError(), patchDiags)
		assert.Equal(t, []string{"/altLabels", "/hiddenLabels", "/prefLabel"}, taxonomyPatchPaths(patch))
	})
}

func TestPreparedTaxonomyConceptSchemePatchFollowsCollectionOwnership(t *testing.T) {
	t.Parallel()

	state := taxonomyConceptSchemeUpdatePlan()
	state.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("top")})
	state.ConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("member")})

	t.Run("omitted arrays remain response-owned", func(t *testing.T) {
		t.Parallel()

		config := state
		plan := state
		plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Updated")})
		config.PrefLabel = plan.PrefLabel
		config.TopConceptIDs = types.ListNull(types.StringType)
		config.ConceptIDs = types.ListNull(types.StringType)
		plan.TopConceptIDs = types.ListUnknown(types.StringType)
		plan.ConceptIDs = types.ListUnknown(types.StringType)

		prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), state)
		require.False(t, patchDiags.HasError())
		require.Len(t, patch, 1)
		assert.Equal(t, "/prefLabel", patch[0].Path)
	})

	t.Run("known empty arrays take ownership and clear remote values", func(t *testing.T) {
		t.Parallel()

		config := state
		plan := state
		config.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
		config.ConceptIDs = types.ListValueMust(types.StringType, []attr.Value{})
		plan.TopConceptIDs = config.TopConceptIDs
		plan.ConceptIDs = config.ConceptIDs

		prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
		require.False(t, diags.HasError())
		patch, patchDiags := prepared.PatchFromState(t.Context(), state)
		require.False(t, patchDiags.HasError())
		assert.ElementsMatch(t, []string{"/concepts", "/topConcepts"}, taxonomyPatchPaths(patch))
	})
}

func TestPreparedTaxonomyConceptNoopStatePreservesResponseOwnedValues(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	state := taxonomyConceptUpdatePlan()
	state.URI = types.StringValue("https://example.com/remote")
	state.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chaise")}),
	})
	state.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{
		"fr-FR": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Siege")}),
	})
	state.Notations = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("REMOTE")})
	state.BroaderConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("parent")})
	state.RelatedConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("related")})
	state.ConceptSchemeIDs = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("scheme")})

	plan := state
	plan.URI = types.StringValue("https://example.com/planned")
	plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Planned")})
	config := plan
	config.AltLabels = types.MapNull(listType)
	config.HiddenLabels = types.MapNull(listType)
	config.Notations = types.ListNull(types.StringType)
	config.BroaderConceptIDs = types.ListNull(types.StringType)
	config.RelatedConceptIDs = types.ListNull(types.StringType)

	prepared, diags := prepareTaxonomyConceptMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	actual := prepared.NoopState(state)

	assert.Equal(t, plan.URI, actual.URI)
	assert.Equal(t, plan.PrefLabel, actual.PrefLabel)
	assert.Equal(t, state.AltLabels, actual.AltLabels)
	assert.Equal(t, state.HiddenLabels, actual.HiddenLabels)
	assert.Equal(t, state.Notations, actual.Notations)
	assert.Equal(t, state.BroaderConceptIDs, actual.BroaderConceptIDs)
	assert.Equal(t, state.RelatedConceptIDs, actual.RelatedConceptIDs)
	assert.Equal(t, state.ConceptSchemeIDs, actual.ConceptSchemeIDs)
	assert.False(t, actual.AltLabels.IsUnknown() || actual.HiddenLabels.IsUnknown() || actual.Notations.IsUnknown())
	assert.False(t, actual.BroaderConceptIDs.IsUnknown() || actual.RelatedConceptIDs.IsUnknown() || actual.ConceptSchemeIDs.IsUnknown())
}

func TestPreparedTaxonomyConceptSchemeNoopStatePreservesComputedValues(t *testing.T) {
	t.Parallel()

	state := taxonomyConceptSchemeUpdatePlan()
	state.URI = types.StringValue("https://example.com/remote")
	state.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("remote-top")})
	state.ConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("remote-member")})
	state.TotalConcepts = types.Int64Value(1)

	plan := state
	plan.URI = types.StringValue("https://example.com/planned")
	plan.PrefLabel = types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Planned")})
	config := plan
	config.TopConceptIDs = types.ListNull(types.StringType)
	config.ConceptIDs = types.ListNull(types.StringType)

	prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	actual := prepared.NoopState(state)

	assert.Equal(t, plan.URI, actual.URI)
	assert.Equal(t, plan.PrefLabel, actual.PrefLabel)
	assert.Equal(t, state.TopConceptIDs, actual.TopConceptIDs)
	assert.Equal(t, state.ConceptIDs, actual.ConceptIDs)
	assert.Equal(t, state.TotalConcepts, actual.TotalConcepts)
	assert.False(t, actual.TopConceptIDs.IsUnknown() || actual.ConceptIDs.IsUnknown() || actual.TotalConcepts.IsUnknown())
}

func taxonomyPatchPaths(patch cm.TaxonomyPatch) []string {
	paths := make([]string, 0, len(patch))
	for _, item := range patch {
		paths = append(paths, item.Path)
	}

	return paths
}

func TestPreparedTaxonomyConceptSchemeMutationOmitsUnknownResponseOwnedArrays(t *testing.T) {
	t.Parallel()

	config := taxonomyConceptSchemeUpdatePlan()
	plan := config
	plan.TopConceptIDs = types.ListUnknown(types.StringType)
	plan.ConceptIDs = types.ListUnknown(types.StringType)

	prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())

	request := prepared.CreateRequest()

	assert.Nil(t, request.TopConcepts)
	assert.Nil(t, request.Concepts)
}

func TestPreparedTaxonomyConceptMutationLabelEquivalenceIsDirectional(t *testing.T) {
	t.Parallel()

	listValue := func(values ...string) attr.Value {
		children := make([]attr.Value, 0, len(values))
		for _, value := range values {
			children = append(children, types.StringValue(value))
		}

		return types.ListValueMust(types.StringType, children)
	}

	tests := []struct {
		name         string
		plannedAlt   map[string]attr.Value
		remoteAlt    map[string][]string
		wantMismatch bool
		wantPath     path.Path
	}{
		{
			name:       "preferred empty response addition is allowed",
			plannedAlt: map[string]attr.Value{},
			remoteAlt:  map[string][]string{"en-US": {}},
		},
		{
			name:         "nonpreferred empty response addition is drift",
			plannedAlt:   map[string]attr.Value{},
			remoteAlt:    map[string][]string{"fr-FR": {}},
			wantMismatch: true,
			wantPath:     path.Root("alt_labels").AtMapKey("fr-FR"),
		},
		{
			name:         "configured empty locale omitted by response is drift",
			plannedAlt:   map[string]attr.Value{"en-US": listValue()},
			remoteAlt:    map[string][]string{},
			wantMismatch: true,
			wantPath:     path.Root("alt_labels").AtMapKey("en-US"),
		},
		{
			name:         "nonempty ordering is exact",
			plannedAlt:   map[string]attr.Value{"en-US": listValue("first", "second")},
			remoteAlt:    map[string][]string{"en-US": {"second", "first"}},
			wantMismatch: true,
			wantPath:     path.Root("alt_labels").AtMapKey("en-US").AtListIndex(0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			remote := taxonomyConceptResponseWithLabelMaps(test.remoteAlt, map[string][]string{"en-US": {}})
			plan, modelDiags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
			require.False(t, modelDiags.HasError())

			listType := types.ListType{ElemType: types.StringType}
			plan.AltLabels = types.MapValueMust(listType, test.plannedAlt)

			prepared, prepareDiags := prepareTaxonomyConceptMutation(t.Context(), plan, plan)
			require.False(t, prepareDiags.HasError())

			state, responseDiags, consistencyDiags := prepared.ProjectResponse(t.Context(), remote)
			require.False(t, responseDiags.HasError())

			if !test.wantMismatch {
				require.False(t, consistencyDiags.HasError())
				assert.Equal(t, plan.AltLabels, state.AltLabels)

				return
			}

			require.True(t, consistencyDiags.HasError())
			assertTaxonomyDiagnosticPath(t, consistencyDiags, test.wantPath)
			remoteModel, remoteDiags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
			require.False(t, remoteDiags.HasError())
			assert.Equal(t, remoteModel.AltLabels, state.AltLabels)
		})
	}
}

func TestPreparedTaxonomyConceptMutationReportsIndependentLabelMismatches(t *testing.T) {
	t.Parallel()

	listType := types.ListType{ElemType: types.StringType}
	remote := taxonomyConceptResponseWithLabelMaps(
		map[string][]string{"en-US": {"remote-alt"}},
		map[string][]string{"en-US": {"remote-hidden"}},
	)
	plan, modelDiags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, modelDiags.HasError())

	plan.AltLabels = types.MapValueMust(listType, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("planned-alt")}),
	})
	plan.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("planned-hidden")}),
	})

	prepared, prepareDiags := prepareTaxonomyConceptMutation(t.Context(), plan, plan)
	require.False(t, prepareDiags.HasError())
	state, responseDiags, consistencyDiags := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	require.Len(t, consistencyDiags.Errors(), 2)
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("alt_labels").AtMapKey("en-US").AtListIndex(0))
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("hidden_labels").AtMapKey("en-US").AtListIndex(0))
	remoteModel, remoteDiags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, remoteDiags.HasError())
	assert.Equal(t, remoteModel.AltLabels, state.AltLabels)
	assert.Equal(t, remoteModel.HiddenLabels, state.HiddenLabels)
}

func TestPreparedTaxonomyConceptMismatchKeepsCompleteRemoteState(t *testing.T) {
	t.Parallel()

	remote := taxonomyConceptResponseWithLabelMaps(
		map[string][]string{"en-US": {}},
		map[string][]string{"en-US": {"remote-hidden"}},
	)
	remote.URI = cm.NewOptNilPointerString(new("https://example.com/remote"))
	remote.Notations = []string{"REMOTE-NOTATION"}
	remote.Broader = []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent")}
	remote.Related = []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("related")}
	remote.Note = cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "Remote note"})

	remoteModel, remoteDiags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, remoteDiags.HasError())

	plan := remoteModel
	listType := types.ListType{ElemType: types.StringType}
	plan.AltLabels = types.MapValueMust(listType, map[string]attr.Value{}) // preferred-empty response addition is equal
	plan.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("planned-hidden")}),
	})

	prepared, prepareDiags := prepareTaxonomyConceptMutation(t.Context(), plan, plan)
	require.False(t, prepareDiags.HasError())
	actual, responseDiags, consistencyDiags := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	require.Len(t, consistencyDiags.Errors(), 1)
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("hidden_labels").AtMapKey("en-US").AtListIndex(0))
	assert.Equal(t, remoteModel, actual)
}

func TestPreparedTaxonomyConceptResponseIdentityMismatchKeepsRequestedIdentity(t *testing.T) {
	t.Parallel()

	remote := taxonomyConceptResponseWithLabelMaps(map[string][]string{"en-US": {}}, map[string][]string{"en-US": {}})
	plan, diags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, diags.HasError())

	listType := types.ListType{ElemType: types.StringType}
	plan.AltLabels = types.MapValueMust(listType, map[string]attr.Value{})
	plan.HiddenLabels = types.MapValueMust(listType, map[string]attr.Value{})
	remote.Sys.Organization.Sys.ID, remote.Sys.ID, remote.Sys.Version = "other", "other", 9
	prepared, diags := prepareTaxonomyConceptMutation(t.Context(), plan, plan)
	require.False(t, diags.HasError())
	state, responseDiags, consistencyDiags := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("organization_id"))
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("concept_id"))
	assert.Equal(t, plan.OrganizationID, state.OrganizationID)
	assert.Equal(t, plan.ConceptID, state.ConceptID)
	assert.Equal(t, NewIDIdentityModelFromMultipartID(plan.OrganizationID.ValueString(), plan.ConceptID.ValueString()), state.IDIdentityModel)
	assert.False(t, state.ID.IsUnknown())
	assert.Equal(t, types.MapValueMust(listType, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
	}), state.AltLabels)
	assert.Equal(t, types.MapValueMust(listType, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
	}), state.HiddenLabels)
	assert.False(t, state.AltLabels.IsUnknown() || state.HiddenLabels.IsUnknown() || state.Notations.IsUnknown() || state.BroaderConceptIDs.IsUnknown() || state.RelatedConceptIDs.IsUnknown())
}

func TestPreparedTaxonomyConceptSchemeResponseIdentityMismatchKeepsRequestedIdentity(t *testing.T) {
	t.Parallel()

	remote := cm.TaxonomyConceptScheme{Sys: cm.TaxonomyConceptSchemeSys{Organization: cm.NewOrganizationLink("organization"), ID: "scheme", Version: 1}, PrefLabel: cm.LocalizedString{"en-US": "Scheme"}, TopConcepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("canonical-top")}, Concepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("canonical-concept")}}
	plan, diags := NewTaxonomyConceptSchemeModelFromResponse(t.Context(), remote)
	require.False(t, diags.HasError())

	config := plan
	config.TopConceptIDs = types.ListNull(types.StringType)
	config.ConceptIDs = types.ListNull(types.StringType)
	plan.TopConceptIDs = types.ListUnknown(types.StringType)
	plan.ConceptIDs = types.ListUnknown(types.StringType)
	remote.Sys.Organization.Sys.ID, remote.Sys.ID, remote.Sys.Version = "other", "other", 9
	prepared, diags := prepareTaxonomyConceptSchemeMutation(t.Context(), config, plan)
	require.False(t, diags.HasError())
	state, responseDiags, consistencyDiags := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("organization_id"))
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("concept_scheme_id"))
	assert.Equal(t, plan.OrganizationID, state.OrganizationID)
	assert.Equal(t, plan.ConceptSchemeID, state.ConceptSchemeID)
	assert.Equal(t, NewIDIdentityModelFromMultipartID(plan.OrganizationID.ValueString(), plan.ConceptSchemeID.ValueString()), state.IDIdentityModel)
	assert.False(t, state.ID.IsUnknown())
	assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("canonical-top")}), state.TopConceptIDs)
	assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("canonical-concept")}), state.ConceptIDs)
	assert.False(t, state.TopConceptIDs.IsUnknown() || state.ConceptIDs.IsUnknown() || state.TotalConcepts.IsUnknown())
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
