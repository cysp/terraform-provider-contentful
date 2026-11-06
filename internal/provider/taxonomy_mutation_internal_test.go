package provider

import (
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	prepared, diags := prepareTaxonomyConceptMutation(config, plan)
	require.False(t, diags.HasError())

	request, requestDiags := prepared.planRequest(t.Context())
	require.False(t, requestDiags.HasError(), requestDiags)

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

	prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
	require.False(t, diags.HasError())

	request, requestDiags := prepared.planRequest(t.Context())
	require.False(t, requestDiags.HasError(), requestDiags)
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
				_, diags := prepareTaxonomyConceptMutation(config, plan)
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
				_, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
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

	prepared, diags := prepareTaxonomyConceptMutation(config, plan)
	require.False(t, diags.HasError())

	request, requestDiags := prepared.planRequest(t.Context())
	require.False(t, requestDiags.HasError(), requestDiags)

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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
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

		prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
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

	prepared, diags := prepareTaxonomyConceptMutation(config, plan)
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

	prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
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

	prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
	require.False(t, diags.HasError())

	request, requestDiags := prepared.planRequest(t.Context())
	require.False(t, requestDiags.HasError(), requestDiags)

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

			prepared, prepareDiags := prepareTaxonomyConceptMutation(plan, plan)
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

	prepared, prepareDiags := prepareTaxonomyConceptMutation(plan, plan)
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

	prepared, prepareDiags := prepareTaxonomyConceptMutation(plan, plan)
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
	prepared, diags := prepareTaxonomyConceptMutation(plan, plan)
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
	prepared, diags := prepareTaxonomyConceptSchemeMutation(config, plan)
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
