//nolint:testpackage // These tests deliberately cover unexported lifecycle helpers without adding production seams.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func taxonomyEmptyLocalizedString() types.Map {
	return types.MapValueMust(types.StringType, map[string]attr.Value{})
}

func taxonomyLocalizedString(value string) types.Map {
	return types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue(value)})
}

func TestTaxonomyNullableLocalizedStringMutationEquivalence(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		planned types.Map
		remote  types.Map
		want    bool
	}{
		"exact empty":                   {taxonomyEmptyLocalizedString(), taxonomyEmptyLocalizedString(), true},
		"exact null":                    {types.MapNull(types.StringType), types.MapNull(types.StringType), true},
		"planned empty remote null":     {taxonomyEmptyLocalizedString(), types.MapNull(types.StringType), true},
		"planned null remote empty":     {types.MapNull(types.StringType), taxonomyEmptyLocalizedString(), false},
		"planned nonempty remote null":  {taxonomyLocalizedString("plan"), types.MapNull(types.StringType), false},
		"planned empty remote nonempty": {taxonomyEmptyLocalizedString(), taxonomyLocalizedString("remote"), false},
		"planned unknown":               {types.MapUnknown(types.StringType), types.MapNull(types.StringType), false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, taxonomyNullableLocalizedStringEquivalentAfterMutation(test.planned, test.remote))
		})
	}
}

func TestTaxonomyNullableLocalizedStringAfterRefresh(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prior  types.Map
		remote types.Map
		want   types.Map
	}{
		"prior empty preserves canonical null": {taxonomyEmptyLocalizedString(), types.MapNull(types.StringType), taxonomyEmptyLocalizedString()},
		"import null remains null":             {types.MapNull(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType)},
		"import null exposes nonempty":         {types.MapNull(types.StringType), taxonomyLocalizedString("remote"), taxonomyLocalizedString("remote")},
		"prior nonempty exposes null":          {taxonomyLocalizedString("prior"), types.MapNull(types.StringType), types.MapNull(types.StringType)},
		"prior empty exposes nonempty":         {taxonomyEmptyLocalizedString(), taxonomyLocalizedString("remote"), taxonomyLocalizedString("remote")},
		"prior unknown does not enter state":   {types.MapUnknown(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, taxonomyNullableLocalizedStringAfterRefresh(test.prior, test.remote))
		})
	}
}

func TestPreparedTaxonomyConceptMutationPreservesAllCanonicalEmptyLocalizedStrings(t *testing.T) {
	t.Parallel()

	remote := taxonomyConceptResponseWithLabelMaps(map[string][]string{}, map[string][]string{})
	plan, diags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, diags.HasError())

	for _, set := range []func(types.Map){
		func(value types.Map) { plan.Note = value },
		func(value types.Map) { plan.ChangeNote = value },
		func(value types.Map) { plan.Definition = value },
		func(value types.Map) { plan.EditorialNote = value },
		func(value types.Map) { plan.Example = value },
		func(value types.Map) { plan.HistoryNote = value },
		func(value types.Map) { plan.ScopeNote = value },
	} {
		set(taxonomyEmptyLocalizedString())
	}

	prepared, diags := prepareTaxonomyConceptMutation(plan, plan)
	require.False(t, diags.HasError())
	state, responseDiags, consistencyDiags, _ := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.False(t, consistencyDiags.HasError())
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.Note)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.ChangeNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.Definition)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.EditorialNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.Example)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.HistoryNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.ScopeNote)
}

func TestPreparedTaxonomyConceptMutationPreservesDistinctLocalizedStrings(t *testing.T) {
	t.Parallel()

	// The all-empty test covers every canonicalization. Distinct values here
	// additionally expose swapped field destinations, including one canonicalized value.
	plan := TaxonomyConceptModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("organization", "concept"),
		OrganizationID:  types.StringValue("organization"), ConceptID: types.StringValue("concept"),
		URI: types.StringNull(), PrefLabel: taxonomyLocalizedString("Concept"),
		AltLabels: types.MapNull(types.ListType{ElemType: types.StringType}), HiddenLabels: types.MapNull(types.ListType{ElemType: types.StringType}),
		Notations: types.ListNull(types.StringType), BroaderConceptIDs: types.ListNull(types.StringType), RelatedConceptIDs: types.ListNull(types.StringType),
		Note: taxonomyLocalizedString("A note"), ChangeNote: taxonomyEmptyLocalizedString(), Definition: taxonomyLocalizedString("A definition"),
		EditorialNote: taxonomyLocalizedString("Editorial guidance"), Example: taxonomyLocalizedString("An example"),
		HistoryNote: taxonomyLocalizedString("History"), ScopeNote: taxonomyLocalizedString("Scope"),
	}
	remote := cm.TaxonomyConcept{
		Sys:           cm.TaxonomyConceptSys{Organization: cm.NewOrganizationLink("organization"), ID: "concept", Version: 2},
		PrefLabel:     cm.LocalizedString{"en-US": "Concept"},
		Note:          cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "A note"}),
		Definition:    cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "A definition"}),
		EditorialNote: cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "Editorial guidance"}),
		Example:       cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "An example"}),
		HistoryNote:   cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "History"}),
		ScopeNote:     cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "Scope"}),
	}
	remote.ChangeNote.SetToNull()

	prepared, diags := prepareTaxonomyConceptMutation(plan, plan)
	require.Empty(t, diags)
	state, responseDiags, consistencyDiags, identityConfirmed := prepared.ProjectResponse(t.Context(), remote)
	require.Empty(t, responseDiags)
	require.Empty(t, consistencyDiags)
	assert.True(t, identityConfirmed)
	assert.Equal(t, taxonomyLocalizedString("A note"), state.Note)
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.ChangeNote)
	assert.Equal(t, taxonomyLocalizedString("A definition"), state.Definition)
	assert.Equal(t, taxonomyLocalizedString("Editorial guidance"), state.EditorialNote)
	assert.Equal(t, taxonomyLocalizedString("An example"), state.Example)
	assert.Equal(t, taxonomyLocalizedString("History"), state.HistoryNote)
	assert.Equal(t, taxonomyLocalizedString("Scope"), state.ScopeNote)
}

func TestPreparedTaxonomyConceptSchemeMismatchKeepsCanonicalDefinitionResponse(t *testing.T) {
	t.Parallel()

	plan := TaxonomyConceptSchemeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("organization", "scheme"),
		OrganizationID:  types.StringValue("organization"), ConceptSchemeID: types.StringValue("scheme"),
		URI: types.StringNull(), PrefLabel: taxonomyLocalizedString("Scheme"), Definition: taxonomyEmptyLocalizedString(),
		TopConceptIDs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("planned-top")}),
		ConceptIDs:    types.ListUnknown(types.StringType),
	}
	config := plan
	config.ConceptIDs = types.ListNull(types.StringType)
	remote := cm.TaxonomyConceptScheme{
		Sys:         cm.TaxonomyConceptSchemeSys{Organization: cm.NewOrganizationLink("organization"), ID: "scheme", Version: 2},
		PrefLabel:   cm.LocalizedString{"en-US": "Scheme"},
		TopConcepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("returned-top")},
		Concepts:    []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("returned-concept")}, TotalConcepts: 1,
	}
	remote.Definition.SetToNull()

	prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
	require.Empty(t, diags)
	state, responseDiags, consistencyDiags, identityConfirmed := prepared.ProjectResponse(t.Context(), remote)
	require.Empty(t, responseDiags)
	require.Len(t, consistencyDiags, 1)
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("top_concept_ids").AtListIndex(0))
	assert.True(t, identityConfirmed)
	assert.Equal(t, types.MapNull(types.StringType), state.Definition, "a later collection contradiction must discard earlier canonicalization")
	assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("returned-top")}), state.TopConceptIDs)
	assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("returned-concept")}), state.ConceptIDs)
	assert.Equal(t, types.Int64Value(1), state.TotalConcepts)
}

func TestPreparedTaxonomyNullableLocalizedStringMismatchKeepsCompleteRemoteState(t *testing.T) {
	t.Parallel()

	remote := taxonomyConceptResponseWithLabelMaps(map[string][]string{}, map[string][]string{})
	remote.Note = cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "remote"})
	plan, diags := NewTaxonomyConceptModelFromResponse(t.Context(), remote)
	require.False(t, diags.HasError())

	plan.Note = taxonomyEmptyLocalizedString()

	plan.Definition = taxonomyEmptyLocalizedString()
	prepared, diags := prepareTaxonomyConceptMutation(plan, plan)
	require.False(t, diags.HasError())

	state, responseDiags, consistencyDiags, _ := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assertTaxonomyDiagnosticPath(t, consistencyDiags, path.Root("note").AtMapKey("en-US"))
	assert.Equal(t, taxonomyLocalizedString("remote"), state.Note)
	assert.True(t, state.Definition.IsNull(), "all-or-nothing recovery must not restore an otherwise equivalent empty value")
}

func TestPreparedTaxonomyConceptSchemeMutationPreservesCanonicalEmptyDefinition(t *testing.T) {
	t.Parallel()

	remote := cm.TaxonomyConceptScheme{Sys: cm.TaxonomyConceptSchemeSys{Organization: cm.NewOrganizationLink("organization-id"), ID: "scheme-id", Version: 2}, PrefLabel: cm.LocalizedString{"en-US": "Scheme"}}
	plan, diags := NewTaxonomyConceptSchemeModelFromResponse(t.Context(), remote)
	require.False(t, diags.HasError())

	plan.Definition = taxonomyEmptyLocalizedString()
	prepared, diags := prepareTaxonomyConceptSchemeMutation(plan, plan)
	require.False(t, diags.HasError())

	state, responseDiags, consistencyDiags, _ := prepared.ProjectResponse(t.Context(), remote)
	require.False(t, responseDiags.HasError())
	require.False(t, consistencyDiags.HasError())
	assert.Equal(t, taxonomyEmptyLocalizedString(), state.Definition)
}

func TestTaxonomyRefreshStatePreservesOnlyCanonicalEmptyLocalizedStrings(t *testing.T) {
	t.Parallel()

	concept := taxonomyConceptResponseWithLabelMaps(map[string][]string{}, map[string][]string{})
	prior := TaxonomyConceptModel{}

	for _, set := range []func(types.Map){
		func(value types.Map) { prior.Note = value },
		func(value types.Map) { prior.ChangeNote = value },
		func(value types.Map) { prior.Definition = value },
		func(value types.Map) { prior.EditorialNote = value },
		func(value types.Map) { prior.Example = value },
		func(value types.Map) { prior.HistoryNote = value },
		func(value types.Map) { prior.ScopeNote = value },
	} {
		set(taxonomyEmptyLocalizedString())
	}

	actual, diags := newTaxonomyConceptRefreshState(t.Context(), prior, concept)
	require.False(t, diags.HasError())
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.Note)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.ChangeNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.Definition)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.EditorialNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.Example)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.HistoryNote)
	assert.Equal(t, taxonomyEmptyLocalizedString(), actual.ScopeNote)

	scheme := cm.TaxonomyConceptScheme{Sys: cm.TaxonomyConceptSchemeSys{Organization: cm.NewOrganizationLink("organization-id"), ID: "scheme-id", Version: 1}, PrefLabel: cm.LocalizedString{"en-US": "Scheme"}}
	schemeState, schemeDiags := newTaxonomyConceptSchemeRefreshState(t.Context(), TaxonomyConceptSchemeModel{Definition: taxonomyEmptyLocalizedString()}, scheme)
	require.False(t, schemeDiags.HasError())
	assert.Equal(t, taxonomyEmptyLocalizedString(), schemeState.Definition)
}
