package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMutationStateRetainsCompleteResponseOnAnyOwnedMismatch(t *testing.T) {
	t.Parallel()

	plannedFields := contentTypeMutationStateTestFields("planned-field", "Planned field", NewTypedObjectNull[ContentTypeFieldItemsValue]())
	mutationResponse := contentTypeMutationStateTestResponse()

	plan := ContentTypeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("planned-space", "planned-environment", "planned-content-type"),
		ContentTypeIdentityModel: ContentTypeIdentityModel{
			SpaceID:       types.StringValue("planned-space"),
			EnvironmentID: types.StringValue("planned-environment"),
			ContentTypeID: types.StringValue("planned-content-type"),
		},
		Name:             types.StringValue("Planned name"),
		Description:      types.StringValue("Planned description"),
		DisplayField:     types.StringValue("planned-field"),
		PublishedVersion: types.Int64Value(99),
		Fields:           plannedFields,
		Metadata:         NewTypedObjectNull[ContentTypeMetadataValue](),
	}
	mutationState, mutationStateDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), mutationResponse, plan, plan)
	require.False(t, mutationStateDiags.HasError())
	assert.True(t, consistencyDiags.HasError())

	assert.Equal(t, types.StringValue("planned-space/planned-environment/planned-content-type"), mutationState.ID)
	assert.Equal(t, types.StringValue("planned-space"), mutationState.SpaceID)
	assert.Equal(t, types.StringValue("planned-environment"), mutationState.EnvironmentID)
	assert.Equal(t, types.StringValue("planned-content-type"), mutationState.ContentTypeID)
	assert.Equal(t, types.Int64Value(7), mutationState.PublishedVersion)
	assert.Equal(t, types.StringValue("Response name"), mutationState.Name)
	assert.Equal(t, types.StringValue("Response description"), mutationState.Description)
	assert.Equal(t, types.StringValue("response-field"), mutationState.DisplayField)
	assert.False(t, mutationState.Fields.Equal(plannedFields))
}

func TestContentTypeMutationStateRejectsUnknownOwnedFieldsWithoutPublishingUnknownState(t *testing.T) {
	t.Parallel()

	nestedUnknownItems := NewTypedObject(ContentTypeFieldItemsValue{
		ItemsType:   types.StringValue("Link"),
		LinkType:    types.StringValue("Entry"),
		Validations: NewTypedListUnknown[jsontypes.Normalized](),
	})

	for name, plannedFields := range map[string]TypedList[TypedObject[ContentTypeFieldValue]]{
		"unknown list": NewTypedListUnknown[TypedObject[ContentTypeFieldValue]](),
		"nested unknown": contentTypeMutationStateTestFields(
			"planned-field",
			"Planned field",
			nestedUnknownItems,
		),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := ContentTypeModel{
				Name:         types.StringValue("Response name"),
				Description:  types.StringValue("Response description"),
				DisplayField: types.StringValue("response-field"),
				Fields:       plannedFields,
				Metadata:     NewTypedObjectUnknown[ContentTypeMetadataValue](),
			}
			mutationState, mutationStateDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), contentTypeMutationStateTestResponse(), plan, plan)
			require.False(t, mutationStateDiags.HasError())
			assert.True(t, consistencyDiags.HasError())
			assert.False(t, mutationState.Fields.IsUnknown())
			require.Len(t, mutationState.Fields.Elements(), 1)
			assert.Equal(t, "response-field", mutationState.Fields.Elements()[0].Value().ID.ValueString())
			assert.Equal(t, "Response field", mutationState.Fields.Elements()[0].Value().Name.ValueString())
			assert.Equal(t, types.Int64Value(7), mutationState.PublishedVersion)
		})
	}
}

func TestContentTypeMutationStateRejectsContradictoryOwnedResponseValues(t *testing.T) {
	t.Parallel()

	base := contentTypeMutationStateTestResponse()
	plan, planDiags := NewContentTypeResourceModelFromResponse(t.Context(), base)
	require.False(t, planDiags.HasError())
	// This represents configured ownership for both metadata children.
	config := plan

	cases := map[string]struct {
		mutation  func(*cm.ContentType)
		valuePath path.Path
	}{
		"name":           {func(v *cm.ContentType) { v.Name = "different" }, path.Root("name")},
		"display field":  {func(v *cm.ContentType) { v.DisplayField = cm.NewNilString("different") }, path.Root("display_field")},
		"field property": {func(v *cm.ContentType) { v.Fields[0].Required = cm.NewOptBool(true) }, path.Root("fields").AtListIndex(0).AtName("required")},
		"missing field":  {func(v *cm.ContentType) { v.Fields = nil }, path.Root("fields").AtListIndex(0)},
		"extra field":    {func(v *cm.ContentType) { v.Fields = append(v.Fields, v.Fields[0]) }, path.Root("fields").AtListIndex(1)},
		"metadata annotations": {func(v *cm.ContentType) {
			metadata := v.Metadata.Or(cm.ContentTypeMetadata{})
			metadata.Annotations = []byte(`{"different":true}`)
			v.Metadata = cm.NewOptContentTypeMetadata(metadata)
		}, path.Root("metadata").AtName("annotations")},
		"metadata taxonomy": {func(v *cm.ContentType) {
			metadata := v.Metadata.Or(cm.ContentTypeMetadata{})
			metadata.Taxonomy = nil
			v.Metadata = cm.NewOptContentTypeMetadata(metadata)
		}, path.Root("metadata").AtName("taxonomy")},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := base
			response.Fields = append([]cm.ContentTypeFieldsItem(nil), base.Fields...)
			mutate.mutation(&response)
			state, responseDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), response, config, plan)
			require.False(t, responseDiags.HasError())
			require.True(t, consistencyDiags.HasError())
			diagnostic, ok := consistencyDiags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, mutate.valuePath, diagnostic.Path())

			switch name {
			case "name":
				assert.Equal(t, "different", state.Name.ValueString())
			case "display field":
				assert.Equal(t, "different", state.DisplayField.ValueString())
			case "field property":
				assert.True(t, state.Fields.Elements()[0].Value().Required.ValueBool())
			case "missing field":
				assert.Empty(t, state.Fields.Elements())
			case "extra field":
				assert.Len(t, state.Fields.Elements(), 2)
			case "metadata annotations":
				assert.JSONEq(t, `{"different":true}`, state.Metadata.Value().Annotations.ValueString())
			case "metadata taxonomy":
				assert.True(t, state.Metadata.Value().Taxonomy.IsNull())
			}

			assert.Equal(t, "Response description", state.Description.ValueString())
			assert.Equal(t, "response-space/response-environment/response-content-type", state.ID.ValueString())
			assert.Equal(t, types.Int64Value(7), state.PublishedVersion)
		})
	}
}

func TestContentTypeMutationStateTreatsResolvedUnknownConfiguredTaxonomyAsOwned(t *testing.T) {
	t.Parallel()

	response := contentTypeMutationStateTestResponse()
	plan, responseDiags := NewContentTypeResourceModelFromResponse(t.Context(), response)
	require.False(t, responseDiags.HasError())

	metadata := response.Metadata.Or(cm.ContentTypeMetadata{})
	metadata.Taxonomy = nil
	response.Metadata = cm.NewOptContentTypeMetadata(metadata)
	_, mutationDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), response, ContentTypeModel{Metadata: NewTypedObjectUnknown[ContentTypeMetadataValue]()}, plan)
	require.False(t, mutationDiags.HasError())
	assert.True(t, consistencyDiags.HasError())
}

func TestContentTypeMutationStateAcceptsNormalizedJSONEquivalence(t *testing.T) {
	t.Parallel()

	response := contentTypeMutationStateTestResponse()
	response.Fields[0].Type = "Array"
	response.Fields[0].Items = cm.NewOptContentTypeFieldsItemItems(cm.ContentTypeFieldsItemItems{
		Type:        cm.NewOptString("Symbol"),
		Validations: []jx.Raw{jx.Raw(`{"size":{"min":1,"max":5}}`)},
	})
	metadata := response.Metadata.Or(cm.ContentTypeMetadata{})
	metadata.Annotations = []byte(`{"a":1,"b":2}`)
	response.Metadata = cm.NewOptContentTypeMetadata(metadata)
	plan, responseDiags := NewContentTypeResourceModelFromResponse(t.Context(), response)
	require.False(t, responseDiags.HasError())

	planMetadata := plan.Metadata.Value()
	planMetadata.Annotations = jsontypes.NewNormalizedValue("{\n  \"b\": 2, \"a\": 1\n}")
	plan.Metadata = NewTypedObject(planMetadata)
	planField := plan.Fields.Elements()[0].Value()
	planField.DefaultValue = jsontypes.NewNormalizedValue(`{"z":2,"a":1}`)
	planField.Validations = NewTypedList([]jsontypes.Normalized{
		jsontypes.NewNormalizedValue(`{"size":{"max":20,"min":1}}`),
	})
	planItems := planField.Items.Value()
	planItems.Validations = NewTypedList([]jsontypes.Normalized{
		jsontypes.NewNormalizedValue(`{"size":{"max":5,"min":1}}`),
	})
	planField.Items = NewTypedObject(planItems)
	plan.Fields = NewTypedList([]TypedObject[ContentTypeFieldValue]{NewTypedObject(planField)})
	state, mutationDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), response, plan, plan)
	require.False(t, mutationDiags.HasError())
	assert.False(t, consistencyDiags.HasError())
	assert.Equal(t, plan.Metadata.Value().Annotations.ValueString(), state.Metadata.Value().Annotations.ValueString())
	assert.Equal(t, planField.DefaultValue.ValueString(), state.Fields.Elements()[0].Value().DefaultValue.ValueString())
	assert.Equal(t, planField.Validations.Elements()[0].ValueString(), state.Fields.Elements()[0].Value().Validations.Elements()[0].ValueString())
	assert.Equal(t, planItems.Validations.Elements()[0].ValueString(), state.Fields.Elements()[0].Value().Items.Value().Validations.Elements()[0].ValueString())
}

func TestContentTypeMutationStateMissingMetadataUsesNullChildSemantics(t *testing.T) {
	t.Parallel()

	response := contentTypeMutationStateTestResponse()
	response.Metadata = cm.OptContentTypeMetadata{}
	plan, responseDiags := NewContentTypeResourceModelFromResponse(t.Context(), response)
	require.False(t, responseDiags.HasError())

	plan.Metadata = NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedNull(),
		Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
	})

	_, mutationDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), response, plan, plan)
	require.False(t, mutationDiags.HasError())
	assert.False(t, consistencyDiags.HasError())

	configured := plan
	configured.Metadata = NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedValue(`{"configured":true}`),
		Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
	})
	state, mutationDiags, consistencyDiags := ProjectContentTypeMutationResponse(t.Context(), response, configured, configured)
	require.False(t, mutationDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	diagnostic, ok := consistencyDiags.Errors()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, path.Root("metadata").AtName("annotations"), diagnostic.Path())
	assert.True(t, state.Metadata.IsNull())
	assert.Equal(t, types.Int64Value(7), state.PublishedVersion)
}

func contentTypeMutationStateTestResponse() cm.ContentType {
	response := cm.ContentType{
		Sys:          cm.NewContentTypeSys("response-space", "response-environment", "response-content-type"),
		Name:         "Response name",
		Description:  cm.NewOptNilString("Response description"),
		DisplayField: cm.NewNilString("response-field"),
		Fields: []cm.ContentTypeFieldsItem{{
			ID:           "response-field",
			Name:         "Response field",
			Type:         "Symbol",
			DefaultValue: jx.Raw(`{"a":1,"z":2}`),
			Validations:  []jx.Raw{jx.Raw(`{"size":{"min":1,"max":20}}`)},
		}},
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
					ID:       "taxonomy-concept",
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
				},
				Required: cm.NewOptBool(false),
			}},
		}),
	}
	response.Sys.Version = 8
	response.Sys.PublishedVersion.SetTo(7)

	return response
}

func contentTypeMutationStateTestFields(fieldID, fieldName string, fieldItems TypedObject[ContentTypeFieldItemsValue]) TypedList[TypedObject[ContentTypeFieldValue]] {
	fieldType := "Symbol"
	if !fieldItems.IsNull() {
		fieldType = "Array"
	}

	return NewTypedList([]TypedObject[ContentTypeFieldValue]{
		NewTypedObject(ContentTypeFieldValue{
			ID:               types.StringValue(fieldID),
			Name:             types.StringValue(fieldName),
			FieldType:        types.StringValue(fieldType),
			LinkType:         types.StringNull(),
			Disabled:         types.BoolValue(false),
			Omitted:          types.BoolValue(false),
			Required:         types.BoolValue(true),
			DefaultValue:     jsontypes.NewNormalizedNull(),
			Items:            fieldItems,
			Localized:        types.BoolValue(false),
			Validations:      NewTypedList([]jsontypes.Normalized{}),
			AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
		}),
	})
}

func TestContentTypeMutationStateLeavesOmittedTaxonomyResponseOwned(t *testing.T) {
	t.Parallel()

	response := contentTypeMutationStateTestResponse()
	planTaxonomy := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
		NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
				ID:       types.StringValue("configured-only-in-plan"),
				Required: types.BoolValue(true),
			}),
			TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		}),
	})
	plan := ContentTypeModel{
		IDIdentityModel:          NewIDIdentityModelFromMultipartID("response-space", "response-environment", "response-content-type"),
		ContentTypeIdentityModel: NewContentTypeIdentityModel("response-space", "response-environment", "response-content-type"),
		Name:                     types.StringValue("Response name"),
		Description:              types.StringValue("Response description"),
		DisplayField:             types.StringValue("response-field"),
		PublishedVersion:         types.Int64Value(7),
		Fields: NewTypedList([]TypedObject[ContentTypeFieldValue]{
			NewTypedObject(ContentTypeFieldValue{
				ID:               types.StringValue("response-field"),
				Name:             types.StringValue("Response field"),
				FieldType:        types.StringValue("Symbol"),
				LinkType:         types.StringNull(),
				Disabled:         types.BoolNull(),
				Omitted:          types.BoolNull(),
				Required:         types.BoolNull(),
				DefaultValue:     jsontypes.NewNormalizedValue(`{"a":1,"z":2}`),
				Items:            NewTypedObjectNull[ContentTypeFieldItemsValue](),
				Localized:        types.BoolNull(),
				Validations:      NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(`{"size":{"min":1,"max":20}}`)}),
				AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
			}),
		}),
		Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    planTaxonomy,
		}),
		Timeouts: TimeoutsNull(),
	}

	state, mutationDiags, consistencyDiags := ProjectContentTypeMutationResponse(
		t.Context(),
		response,
		ContentTypeModel{Metadata: NewTypedObjectNull[ContentTypeMetadataValue]()},
		plan,
	)
	require.False(t, mutationDiags.HasError())
	require.False(t, consistencyDiags.HasError(), "%v", consistencyDiags)
	assert.True(t, state.Metadata.Value().Annotations.IsNull())
	require.Len(t, state.Metadata.Value().Taxonomy.Elements(), 1)
	responseTaxonomy := state.Metadata.Value().Taxonomy.Elements()[0].Value()
	assert.Equal(t, "taxonomy-concept", responseTaxonomy.TaxonomyConcept.Value().ID.ValueString())
	assert.False(t, responseTaxonomy.TaxonomyConcept.Value().Required.ValueBool())
	assert.False(t, state.Metadata.Value().Taxonomy.Equal(planTaxonomy))
}

func editorLayoutPlanGroup(id, name string, items TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]]) TypedObject[EditorInterfaceEditorLayoutItemValue] {
	return NewTypedObject(EditorInterfaceEditorLayoutItemValue{
		Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupValue{
			GroupID: types.StringValue(id),
			Name:    types.StringValue(name),
			Items:   items,
		}),
	})
}

func editorLayoutResponseGroup(id, name string) cm.EditorInterfaceEditorLayoutItem {
	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: id,
		Name:    name,
		Items:   []cm.EditorInterfaceEditorLayoutItem{},
	})
}

func TestEditorInterfaceMutationStateRejectsLossyLayoutProjection(t *testing.T) {
	t.Parallel()

	plannedLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
		editorLayoutPlanGroup("layout-group", "Layout group", NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{})),
	})
	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "title"}),
		}),
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: plannedLayout})
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 1)
	assert.True(t, consistencyDiags.HasError())
	assert.False(t, mutationState.EditorLayout.Equal(plannedLayout))
	assert.True(t, mutationState.EditorLayout.Elements()[0].Value().Group.IsNull())

	unknownPlanState, unknownPlanDiags, unknownConsistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, unknownPlanDiags.HasError())
	assert.Empty(t, unknownConsistencyDiags)
	assert.False(t, unknownPlanState.EditorLayout.IsUnknown())
	assert.True(t, unknownPlanState.EditorLayout.Elements()[0].Value().Group.IsNull())

	nullPlanState, nullPlanDiags, nullConsistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, nullPlanDiags.HasError())
	assert.True(t, nullConsistencyDiags.HasError())
	assert.False(t, nullPlanState.EditorLayout.IsNull())
	assert.True(t, nullPlanState.EditorLayout.Elements()[0].Value().Group.IsNull())
}

func TestEditorInterfaceMutationStateRejectsRepresentableLayoutContradiction(t *testing.T) {
	t.Parallel()

	plannedLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
		editorLayoutPlanGroup("planned-group", "Planned group", NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{})),
	})
	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			editorLayoutResponseGroup("response-group", "Response group"),
		}),
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(
		t.Context(), response, EditorInterfaceModel{EditorLayout: plannedLayout},
	)

	require.False(t, mutationStateDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"editor_layout"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "Contentful returned a different Editor Interface layout", consistencyDiags.Errors()[0].Summary())
	assert.Equal(t, "response-group", mutationState.EditorLayout.Elements()[0].Value().Group.Value().GroupID.ValueString())

	nullPlanState, nullPlanResponseDiags, nullPlanConsistencyDiags := ReconcileEditorInterfaceMutationResponse(
		t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()},
	)

	assert.Empty(t, nullPlanResponseDiags)
	require.True(t, nullPlanConsistencyDiags.HasError())
	assert.Equal(t, []string{"editor_layout"}, attributeDiagnosticPaths(t, nullPlanConsistencyDiags))
	require.Len(t, nullPlanState.EditorLayout.Elements(), 1)
	nullPlanLayoutItem := nullPlanState.EditorLayout.Elements()[0].Value()
	require.False(t, nullPlanLayoutItem.Group.IsNull())
	assert.Equal(t, "response-group", nullPlanLayoutItem.Group.Value().GroupID.ValueString())
	assert.Equal(t, "Response group", nullPlanLayoutItem.Group.Value().Name.ValueString())
	assert.Empty(t, nullPlanLayoutItem.Group.Value().Items.Elements())
}

func TestEditorInterfaceMutationStateAcceptsEquivalentResponse(t *testing.T) {
	t.Parallel()

	plannedLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
		editorLayoutPlanGroup("group", "Group", NewTypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]](nil)),
	})
	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			editorLayoutResponseGroup("group", "Group"),
		}),
	}

	mutationState, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(
		t.Context(), response, EditorInterfaceModel{EditorLayout: plannedLayout},
	)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, mutationState.EditorLayout.Equal(plannedLayout))
}

func TestEditorInterfaceMutationStateTreatsLayoutOrderAsMeaningful(t *testing.T) {
	t.Parallel()

	plannedLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
		editorLayoutPlanGroup("first", "first", NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{})),
		editorLayoutPlanGroup("second", "second", NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{})),
	})
	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			editorLayoutResponseGroup("second", "second"),
			editorLayoutResponseGroup("first", "first"),
		}),
	}

	mutationState, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(
		t.Context(), response, EditorInterfaceModel{EditorLayout: plannedLayout},
	)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, "second", mutationState.EditorLayout.Elements()[0].Value().Group.Value().GroupID.ValueString())
}

func TestOptionalMutationStateReconcilesOmittedNullValues(t *testing.T) {
	t.Parallel()

	editorState, editorResponseDiags, editorConsistencyDiags := ReconcileEditorInterfaceMutationResponse(
		t.Context(),
		cm.EditorInterface{Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default")},
		EditorInterfaceModel{EditorLayout: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()},
	)
	assert.Empty(t, editorResponseDiags)
	assert.Empty(t, editorConsistencyDiags)
	assert.True(t, editorState.EditorLayout.IsNull())

	webhookState, webhookResponseDiags, webhookConsistencyDiags := ReconcileWebhookMutationResponse(
		t.Context(),
		cm.WebhookDefinition{Sys: cm.NewWebhookDefinitionSys("space", "webhook")},
		WebhookModel{
			Name:    types.StringUnknown(),
			URL:     types.StringUnknown(),
			Filters: NewTypedListNull[TypedObject[WebhookFilterValue]](),
			Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
			Topics:  NewTypedList([]types.String{}),
		},
	)
	assert.Empty(t, webhookResponseDiags)
	assert.Empty(t, webhookConsistencyDiags)
	assert.True(t, webhookState.Filters.IsNull())
}

func TestWebhookMutationStateDistinguishesNullAndEmptyFilters(t *testing.T) {
	t.Parallel()

	nullFilters := NewTypedListNull[TypedObject[WebhookFilterValue]]()
	emptyFilters := NewTypedList([]TypedObject[WebhookFilterValue]{})

	for name, test := range map[string]struct {
		plan                TypedList[TypedObject[WebhookFilterValue]]
		response            cm.OptNilWebhookDefinitionFilterArray
		expectedState       TypedList[TypedObject[WebhookFilterValue]]
		expectContradiction bool
	}{
		"null plan and empty response": {
			plan:                nullFilters,
			response:            cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{}),
			expectedState:       emptyFilters,
			expectContradiction: true,
		},
		"empty plan and null response": {
			plan:                emptyFilters,
			response:            cm.NewOptNilWebhookDefinitionFilterArrayNull(),
			expectedState:       nullFilters,
			expectContradiction: true,
		},
		"empty plan and empty response": {
			plan:          emptyFilters,
			response:      cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{}),
			expectedState: emptyFilters,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(
				t.Context(),
				cm.WebhookDefinition{
					Sys:     cm.NewWebhookDefinitionSys("space", "webhook"),
					Filters: test.response,
				},
				WebhookModel{
					Name:    types.StringUnknown(),
					URL:     types.StringUnknown(),
					Filters: test.plan,
					Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
					Topics:  NewTypedList([]types.String{}),
				},
			)

			assert.Empty(t, responseDiags)
			assert.Equal(t, test.expectedState, state.Filters)

			if test.expectContradiction {
				require.True(t, consistencyDiags.HasError())
				assert.Equal(t, []string{"filters"}, attributeDiagnosticPaths(t, consistencyDiags))
				assert.Equal(t, "Contentful returned different webhook filters", consistencyDiags.Errors()[0].Summary())

				return
			}

			assert.Empty(t, consistencyDiags)
		})
	}
}

func TestRoleMutationStateRejectsLossyPermissionsAndPoliciesProjection(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read")}),
	})
	plannedPolicies := NewTypedList([]TypedObject[RolePolicyValue]{
		rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedNull()),
	})
	response := cm.Role{
		Sys: cm.NewRoleSys("space", "role"),
		Permissions: cm.RolePermissions{
			"Entry": {Type: cm.RolePermissionsItemType("future-actions")},
		},
		Policies: []cm.RolePoliciesItem{{
			Effect:  cm.RolePoliciesItemEffect("future-effect"),
			Actions: cm.RolePoliciesItemActions{Type: cm.RolePoliciesItemActionsType("future-actions")},
		}},
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileRoleMutationResponse(t.Context(), response, RoleModel{Name: types.StringUnknown(), Permissions: plannedPermissions, Policies: plannedPolicies})
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 3)
	assert.True(t, consistencyDiags.HasError())
	assert.Len(t, consistencyDiags.Errors(), 2)
	assert.False(t, mutationState.Permissions.Equal(plannedPermissions))
	assert.False(t, mutationState.Policies.Equal(plannedPolicies))

	unknownPlanState, unknownPlanDiags, unknownConsistencyDiags := ReconcileRoleMutationResponse(t.Context(), response, RoleModel{
		Name:        types.StringUnknown(),
		Permissions: NewTypedMapUnknown[TypedList[types.String]](),
		Policies:    NewTypedListUnknown[TypedObject[RolePolicyValue]](),
	})
	assert.False(t, unknownPlanDiags.HasError())
	assert.Empty(t, unknownConsistencyDiags)
	assert.False(t, unknownPlanState.Permissions.IsUnknown())
	assert.False(t, unknownPlanState.Policies.IsUnknown())
	assert.True(t, unknownPlanState.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, unknownPlanState.Policies.Elements()[0].Value().Effect.IsNull())
}

func TestRoleMutationStateRejectsRepresentablePermissionsContradiction(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read")}),
	})
	plannedPolicies := NewTypedList([]TypedObject[RolePolicyValue]{
		rolePolicyValue("allow", []string{"read", "create"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
	})
	response := cm.Role{
		Sys:  cm.NewRoleSys("space", "role"),
		Name: "Response role",
		Permissions: cm.RolePermissions{
			"Entry": cm.NewStringArrayRolePermissionsItem([]string{"manage"}),
		},
		Policies: []cm.RolePoliciesItem{{
			Actions:    cm.NewStringArrayRolePoliciesItemActions([]string{"create", "read"}),
			Constraint: []byte(`{"sys":{"type":"Entry"}}`),
			Effect:     cm.RolePoliciesItemEffect("allow"),
		}},
	}

	mutationState, mutationStateDiags, consistencyDiags := ReconcileRoleMutationResponse(
		t.Context(), response, RoleModel{Name: types.StringUnknown(), Permissions: plannedPermissions, Policies: plannedPolicies},
	)

	require.False(t, mutationStateDiags.HasError())
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"permissions"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "Contentful returned different role permissions", consistencyDiags.Errors()[0].Summary())
	assert.Equal(t, []types.String{types.StringValue("manage")}, mutationState.Permissions.Elements()["Entry"].Elements())
	assert.Equal(t, []types.String{types.StringValue("create"), types.StringValue("read")}, mutationState.Policies.Elements()[0].Value().Actions.Elements())
	assert.Equal(t, "Response role", mutationState.Name.ValueString())
}

func TestRoleMutationStateRestoresSemanticallyEquivalentReorderedPlan(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringValue("create"), types.StringValue("read")}),
	})
	plannedPolicies := NewTypedList([]TypedObject[RolePolicyValue]{
		rolePolicyValue("allow", []string{"read", "create"}, jsontypes.NewNormalizedValue(`{"b":2,"a":1}`)),
		rolePolicyValue("deny", []string{"delete"}, jsontypes.NewNormalizedNull()),
	})
	response := cm.Role{
		Sys:  cm.NewRoleSys("space", "role"),
		Name: "Response role",
		Permissions: cm.RolePermissions{
			"Entry": cm.NewStringArrayRolePermissionsItem([]string{"read", "read", "create"}),
		},
		Policies: []cm.RolePoliciesItem{
			{Actions: cm.NewStringArrayRolePoliciesItemActions([]string{"delete"}), Effect: cm.RolePoliciesItemEffect("deny")},
			{Actions: cm.NewStringArrayRolePoliciesItemActions([]string{"create", "read"}), Constraint: []byte(`{"a":1,"b":2}`), Effect: cm.RolePoliciesItemEffect("allow")},
		},
	}

	mutationState, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(
		t.Context(), response, RoleModel{Name: types.StringUnknown(), Permissions: plannedPermissions, Policies: plannedPolicies},
	)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, mutationState.Permissions.Equal(plannedPermissions))
	assert.True(t, mutationState.Policies.Equal(plannedPolicies))
}

func TestRoleMutationStatePreservesDuplicateMultiplicity(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringValue("read")}),
	})
	plannedPolicies := NewTypedList([]TypedObject[RolePolicyValue]{})
	response := cm.Role{
		Sys: cm.NewRoleSys("space", "role"),
		Permissions: cm.RolePermissions{
			"Entry": cm.NewStringArrayRolePermissionsItem([]string{"read", "create"}),
		},
		Policies: []cm.RolePoliciesItem{},
	}

	mutationState, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(
		t.Context(), response, RoleModel{Name: types.StringUnknown(), Permissions: plannedPermissions, Policies: plannedPolicies},
	)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []types.String{types.StringValue("read"), types.StringValue("create")}, mutationState.Permissions.Elements()["Entry"].Elements())
}

func TestRoleMutationStateRejectsRepresentablePolicyContradictions(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringValue("create")}),
	})
	for name, test := range map[string]struct {
		planned             []TypedObject[RolePolicyValue]
		response            []cm.RolePoliciesItem
		expectedEffects     []string
		expectedActions     [][]types.String
		expectedConstraints []string
	}{
		"effect": {
			planned: []TypedObject[RolePolicyValue]{
				rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
			},
			response: []cm.RolePoliciesItem{
				rolePolicyResponse("deny", []string{"read"}, `{"sys":{"type":"Entry"}}`),
			},
			expectedEffects:     []string{"deny"},
			expectedActions:     [][]types.String{{types.StringValue("read")}},
			expectedConstraints: []string{`{"sys":{"type":"Entry"}}`},
		},
		"actions": {
			planned: []TypedObject[RolePolicyValue]{
				rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
			},
			response: []cm.RolePoliciesItem{
				rolePolicyResponse("allow", []string{"create"}, `{"sys":{"type":"Entry"}}`),
			},
			expectedEffects:     []string{"allow"},
			expectedActions:     [][]types.String{{types.StringValue("create")}},
			expectedConstraints: []string{`{"sys":{"type":"Entry"}}`},
		},
		"constraint": {
			planned: []TypedObject[RolePolicyValue]{
				rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
			},
			response: []cm.RolePoliciesItem{
				rolePolicyResponse("allow", []string{"read"}, `{"sys":{"type":"Asset"}}`),
			},
			expectedEffects:     []string{"allow"},
			expectedActions:     [][]types.String{{types.StringValue("read")}},
			expectedConstraints: []string{`{"sys":{"type":"Asset"}}`},
		},
		"duplicate multiplicity": {
			planned: []TypedObject[RolePolicyValue]{
				rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
				rolePolicyValue("allow", []string{"read"}, jsontypes.NewNormalizedValue(`{"sys":{"type":"Entry"}}`)),
			},
			response: []cm.RolePoliciesItem{
				rolePolicyResponse("allow", []string{"read"}, `{"sys":{"type":"Entry"}}`),
				rolePolicyResponse("allow", []string{"read"}, `{"sys":{"type":"Asset"}}`),
			},
			expectedEffects:     []string{"allow", "allow"},
			expectedActions:     [][]types.String{{types.StringValue("read")}, {types.StringValue("read")}},
			expectedConstraints: []string{`{"sys":{"type":"Entry"}}`, `{"sys":{"type":"Asset"}}`},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plannedPolicies := NewTypedList(test.planned)
			response := cm.Role{
				Sys:         cm.NewRoleSys("space", "role"),
				Name:        "Response role",
				Permissions: cm.RolePermissions{"Entry": cm.NewStringArrayRolePermissionsItem([]string{"create", "read"})},
				Policies:    test.response,
			}

			mutationState, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(
				t.Context(), response, RoleModel{Name: types.StringUnknown(), Permissions: plannedPermissions, Policies: plannedPolicies},
			)

			assert.Empty(t, responseDiags)
			require.True(t, consistencyDiags.HasError())
			assert.Equal(t, []string{"policies"}, attributeDiagnosticPaths(t, consistencyDiags))
			assert.False(t, mutationState.Policies.Equal(plannedPolicies))
			assert.Equal(t, "Response role", mutationState.Name.ValueString())
			assert.False(t, mutationState.Permissions.Equal(plannedPermissions))
			assert.Equal(t, []types.String{types.StringValue("create"), types.StringValue("read")}, mutationState.Permissions.Elements()["Entry"].Elements())
			require.Len(t, mutationState.Policies.Elements(), len(test.expectedEffects))

			for index, expectedEffect := range test.expectedEffects {
				actualPolicy := mutationState.Policies.Elements()[index].Value()
				assert.Equal(t, expectedEffect, actualPolicy.Effect.ValueString())
				assert.Equal(t, test.expectedActions[index], actualPolicy.Actions.Elements())
				assert.Equal(t, test.expectedConstraints[index], actualPolicy.Constraint.ValueString())
			}
		})
	}
}

func rolePolicyValue(effect string, actions []string, constraint jsontypes.Normalized) TypedObject[RolePolicyValue] {
	typedActions := make([]types.String, len(actions))
	for index, action := range actions {
		typedActions[index] = types.StringValue(action)
	}

	return NewTypedObject(RolePolicyValue{
		Actions:    NewTypedList(typedActions),
		Constraint: constraint,
		Effect:     types.StringValue(effect),
	})
}

func rolePolicyResponse(effect string, actions []string, constraint string) cm.RolePoliciesItem {
	return cm.RolePoliciesItem{
		Actions:    cm.NewStringArrayRolePoliciesItemActions(actions),
		Constraint: []byte(constraint),
		Effect:     cm.RolePoliciesItemEffect(effect),
	}
}
