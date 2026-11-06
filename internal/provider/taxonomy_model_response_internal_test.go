package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
