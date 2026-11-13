package provider_test

import (
	"fmt"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Exercise simultaneous ownership decisions, including mismatched Config/Plan
// booleans and multiple failures. The example tests vary one attribute at a time.
// Lifecycle configuration is known; unknown Config also exercises the defensive
// converter rule that any non-null configuration represents ownership.
func TestSpaceEnablementsRequestStateCombinations(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	configuredValues := []types.Bool{types.BoolNull(), types.BoolUnknown(), types.BoolValue(false), types.BoolValue(true)}
	plannedValues := []types.Bool{types.BoolNull(), types.BoolUnknown(), types.BoolValue(false), types.BoolValue(true)}
	// Rows and columns are null/unknown/false/true for Config and Plan respectively.
	rejected := [4][4]bool{{false, false, false, false}, {false, true, false, false}, {false, true, false, false}, {false, true, false, false}}
	wireValues := [4]cm.OptSpaceEnablementField{
		{}, {}, {Set: true, Value: cm.SpaceEnablementField{Enabled: false}}, {Set: true, Value: cm.SpaceEnablementField{Enabled: true}},
	}
	names := [4]string{"cross_space_links", "space_templates", "studio_experiences", "suggest_concepts"}

	rapid.Check(t, func(t *rapid.T) {
		var (
			configured, planned [4]types.Bool
			fields              [4]cm.OptSpaceEnablementField
		)

		wantPaths := []string{}

		for index, name := range names {
			configState := rapid.IntRange(0, 3).Draw(t, name+" config")
			planState := rapid.IntRange(0, 3).Draw(t, name+" plan")
			configured[index], planned[index] = configuredValues[configState], plannedValues[planState]
			fields[index] = wireValues[planState]

			if rejected[configState][planState] {
				wantPaths = append(wantPaths, name)
			}
		}

		model := SpaceEnablementsModel{CrossSpaceLinks: planned[0], SpaceTemplates: planned[1], StudioExperiences: planned[2], SuggestConcepts: planned[3]}
		config := SpaceEnablementsModel{CrossSpaceLinks: configured[0], SpaceTemplates: configured[1], StudioExperiences: configured[2], SuggestConcepts: configured[3]}
		actual, diags := model.ToSpaceEnablementData(ctx, config)
		require.Equal(t, wantPaths, requestPropertyErrorPaths(t, diags))
		require.Empty(t, diags.Warnings())

		if len(wantPaths) != 0 {
			require.Equal(t, cm.SpaceEnablementData{}, actual, "any error discards the complete request")

			return
		}

		require.Equal(t, cm.SpaceEnablementData{CrossSpaceLinks: fields[0], SpaceTemplates: fields[1], StudioExperiences: fields[2], SuggestConcepts: fields[3]}, actual)
	})
}

// A literal catalog defines wire values and relative diagnostic paths without
// using the converters. Lists combine successful unions, unresolved containers,
// invalid cardinalities, and invalid leaves at independently chosen positions.
func TestContentTypeTaxonomyRequestStateCombinations(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	nullConcept := NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue]()
	nullScheme := NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue]()
	concept := NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{ID: types.StringValue(""), Required: types.BoolValue(false)})
	scheme := NewTypedObject(ContentTypeMetadataTaxonomyItemConceptSchemeValue{ID: types.StringValue("scheme"), Required: types.BoolNull()})

	type requestCase struct {
		item  TypedObject[ContentTypeMetadataTaxonomyItemValue]
		wire  cm.ContentTypeMetadataTaxonomyItem
		paths []string
	}

	cases := []requestCase{
		{item: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue](), paths: []string{""}},
		{item: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemValue](), paths: []string{""}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{TaxonomyConcept: nullConcept, TaxonomyConceptScheme: nullScheme}), paths: []string{""}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{TaxonomyConcept: concept, TaxonomyConceptScheme: scheme}), paths: []string{""}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{TaxonomyConcept: concept, TaxonomyConceptScheme: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptSchemeValue]()}), paths: []string{".taxonomy_concept_scheme"}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{ID: types.StringUnknown(), Required: types.BoolUnknown()}), TaxonomyConceptScheme: nullScheme,
		}), paths: []string{".taxonomy_concept.id", ".taxonomy_concept.required"}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{TaxonomyConcept: concept, TaxonomyConceptScheme: nullScheme}), wire: cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{Type: "Link", LinkType: "TaxonomyConcept", ID: ""}, Required: cm.OptBool{Set: true, Value: false},
		}},
		{item: NewTypedObject(ContentTypeMetadataTaxonomyItemValue{TaxonomyConcept: nullConcept, TaxonomyConceptScheme: scheme}), wire: cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{Type: "Link", LinkType: "TaxonomyConceptScheme", ID: "scheme"},
		}},
	}

	rapid.Check(t, func(t *rapid.T) {
		choices := rapid.SliceOfN(rapid.IntRange(0, len(cases)-1), 0, 8).Draw(t, "items")
		items := make([]TypedObject[ContentTypeMetadataTaxonomyItemValue], len(choices))
		want := make([]cm.ContentTypeMetadataTaxonomyItem, len(choices))
		wantPaths := []string{}
		parent := rapid.IntRange(0, 9).Draw(t, "parent index")
		valuePath := path.Root("content_types").AtListIndex(parent).AtName("metadata").AtName("taxonomy")

		for index, choice := range choices {
			test := cases[choice]
			items[index], want[index] = test.item, test.wire

			for _, suffix := range test.paths {
				wantPaths = append(wantPaths, fmt.Sprintf("content_types[%d].metadata.taxonomy[%d]%s", parent, index, suffix))
			}
		}

		actual, diags := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(ctx, valuePath, NewTypedList(items))
		require.Equal(t, wantPaths, requestPropertyErrorPaths(t, diags))
		require.Empty(t, diags.Warnings())

		if len(wantPaths) != 0 {
			require.Nil(t, actual, "valid siblings must not escape a failed request")

			return
		}

		require.Equal(t, want, actual, "preserve order, duplicates, empty IDs, false, and omitted optional booleans")
	})
}

func requestPropertyErrorPaths(t *rapid.T, diags diag.Diagnostics) []string {
	paths := make([]string, 0, len(diags.Errors()))

	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok, "expected attribute diagnostic, got %T", diagnostic)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
