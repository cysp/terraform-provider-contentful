package provider

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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
	assert.JSONEq(t, testJSON(map[string]any{"en-US": "Product catalog"}), string(patch[0].Value))
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
		"altLabels": json.RawMessage(testJSON(map[string]any{"en-US": []any{}})),
	})
	require.NoError(t, err)
	require.Len(t, patch, 1)
	assert.Equal(t, "/altLabels", patch[0].Path)
	assert.JSONEq(t, testJSON(map[string]any{"en-US": []any{}}), string(patch[0].Value))
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
			path:   "/definition", value: testJSON(nil),
		},
		"primitive list to empty": {
			mutate: func(desired *cm.TaxonomyConceptRequest) { desired.Notations = []string{} },
			path:   "/notations", value: testJSON([]any{}),
		},
		"localized list map to empty": {
			mutate: func(desired *cm.TaxonomyConceptRequest) {
				desired.AltLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{})
			},
			path: "/altLabels", value: testJSON(map[string]any{}),
		},
		"relationship replacement": {
			mutate: func(desired *cm.TaxonomyConceptRequest) {
				desired.Broader = []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("replacement")}
			},
			path: "/broader", value: testJSON([]any{
				map[string]any{"sys": map[string]any{"type": "Link", "linkType": "TaxonomyConcept", "id": "replacement"}},
			}),
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
