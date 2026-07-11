//nolint:testpackage
package provider

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func TestLabelMapWithConfiguredKeys(t *testing.T) {
	t.Parallel()

	configured := types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
		"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chair")}),
	})
	returned := types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
		"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chair")}),
		"en-US": types.ListValueMust(types.StringType, []attr.Value{}),
	})

	filtered := labelMapWithConfiguredKeys(configured, returned)
	assert.Equal(t, map[string]attr.Value{
		"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Chair")}),
	}, filtered.Elements())
	assert.Equal(t, returned, labelMapWithConfiguredKeys(types.MapNull(configured.ElementType(t.Context())), returned))
}

func TestValidateTaxonomyConceptResponse(t *testing.T) {
	t.Parallel()

	request := cm.TaxonomyConceptRequest{
		PrefLabel:    cm.LocalizedString{"en-US": "Chair"},
		AltLabels:    cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-GB": {"Seat"}}),
		HiddenLabels: cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-GB": {"Stool"}}),
	}
	response := cm.TaxonomyConcept{
		PrefLabel:    cm.LocalizedString{"en-US": "Chair"},
		AltLabels:    cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-GB": {"Seat"}, "en-US": {}}),
		HiddenLabels: cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-GB": {"Stool"}, "en-US": {}}),
	}

	require.NoError(t, validateTaxonomyConceptResponse(request, response))

	response.PrefLabel = cm.LocalizedString{"en-GB": "Chair"}
	err := validateTaxonomyConceptResponse(request, response)
	require.ErrorIs(t, err, errTaxonomyLocaleNotPreserved)

	response.PrefLabel = cm.LocalizedString{"en-US": "Chair"}

	response.AltLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {}})
	err = validateTaxonomyConceptResponse(request, response)
	require.ErrorIs(t, err, errTaxonomyLocaleNotPreserved)

	response.AltLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-GB": {"Seat"}, "en-US": {}})
	response.HiddenLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {}})
	err = validateTaxonomyConceptResponse(request, response)
	require.ErrorIs(t, err, errTaxonomyLocaleNotPreserved)
}

func TestTaxonomyCollectionConversions(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]types.Map{
		"null":    types.MapNull(types.StringType),
		"unknown": types.MapUnknown(types.StringType),
	} {
		t.Run("string map "+name, func(t *testing.T) {
			t.Parallel()

			actual, diags := stringMap(t.Context(), value)
			require.False(t, diags.HasError())
			assert.Empty(t, actual)
		})
	}

	listType := types.ListType{ElemType: types.StringType}
	for name, value := range map[string]types.Map{
		"null":    types.MapNull(listType),
		"unknown": types.MapUnknown(listType),
	} {
		t.Run("string list map "+name, func(t *testing.T) {
			t.Parallel()

			actual, diags := stringListMap(t.Context(), value)
			require.False(t, diags.HasError())
			assert.Empty(t, actual)
		})
	}
}

func TestTaxonomyConceptToRequestAddsPreferredLocalesToLabelMaps(t *testing.T) {
	t.Parallel()

	model := TaxonomyConceptModel{
		PrefLabel:         types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Chair")}),
		AltLabels:         types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{"en-GB": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Seat")})}),
		HiddenLabels:      types.MapNull(types.ListType{ElemType: types.StringType}),
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

func TestValidateTaxonomyConceptSchemeResponse(t *testing.T) {
	t.Parallel()

	request := cm.TaxonomyConceptSchemeRequest{PrefLabel: cm.LocalizedString{"en-US": "Products"}}
	response := cm.TaxonomyConceptScheme{PrefLabel: cm.LocalizedString{"en-US": "Products", "en-GB": "Products"}}
	require.NoError(t, validateTaxonomyConceptSchemeResponse(request, response))

	response.PrefLabel = cm.LocalizedString{"en-GB": "Products"}
	err := validateTaxonomyConceptSchemeResponse(request, response)
	require.ErrorIs(t, err, errTaxonomyLocaleNotPreserved)
}

func TestLocalizedStringConversions(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	diags := diag.Diagnostics{}
	null := nullableLocalizedString(ctx, types.MapNull(types.StringType), &diags)
	require.False(t, diags.HasError())
	assert.True(t, null.IsNull())
	assert.True(t, localizedStringValue(ctx, null, &diags).IsNull())

	configured := types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Description")})
	known := nullableLocalizedString(ctx, configured, &diags)
	require.False(t, diags.HasError())

	value, ok := known.Get()
	require.True(t, ok)
	assert.Equal(t, cm.NullableLocalizedString{"en-US": "Description"}, value)
	assert.Equal(t, configured, localizedStringValue(ctx, known, &diags))
}
