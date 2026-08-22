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

func TestEditorInterfaceMutationStateRestoresKnownPlanOwnedLayout(t *testing.T) {
	t.Parallel()

	plannedLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
		NewTypedObject(EditorInterfaceEditorLayoutItemValue{
			Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupValue{
				GroupID: types.StringValue("layout-group"),
				Name:    types.StringValue("Layout group"),
				Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{}),
			}),
		}),
	})
	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "title"}),
		}),
	}

	mutationState, mutationStateDiags := NewEditorInterfaceResourceModelForMutationState(t.Context(), response, EditorInterfaceModel{EditorLayout: plannedLayout})
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 1)
	assert.True(t, mutationState.EditorLayout.Equal(plannedLayout))

	readState, readDiags := NewEditorInterfaceResourceModelFromResponse(t.Context(), response)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 1)
	assert.True(t, readState.EditorLayout.Elements()[0].Value().Group.IsNull())

	unknownPlanState, unknownPlanDiags := NewEditorInterfaceResourceModelForMutationState(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanState.EditorLayout.IsUnknown())
	assert.True(t, unknownPlanState.EditorLayout.Elements()[0].Value().Group.IsNull())

	nullPlanState, nullPlanDiags := NewEditorInterfaceResourceModelForMutationState(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, nullPlanDiags.HasError())
	assert.True(t, nullPlanState.EditorLayout.IsNull())
}

func TestRoleMutationStateRestoresKnownPlanOwnedPermissionsAndPolicies(t *testing.T) {
	t.Parallel()

	plannedPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read")}),
	})
	plannedPolicies := NewTypedList([]TypedObject[RolePolicyValue]{
		NewTypedObject(RolePolicyValue{
			Actions:    NewTypedList([]types.String{types.StringValue("read")}),
			Constraint: jsontypes.NewNormalizedNull(),
			Effect:     types.StringValue("allow"),
		}),
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

	mutationState, mutationStateDiags := NewRoleResourceModelForMutationState(t.Context(), response, RoleModel{Permissions: plannedPermissions, Policies: plannedPolicies})
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 3)
	assert.True(t, mutationState.Permissions.Equal(plannedPermissions))
	assert.True(t, mutationState.Policies.Equal(plannedPolicies))

	readState, readDiags := NewRoleResourceModelFromResponse(t.Context(), response)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 3)
	assert.True(t, readState.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, readState.Policies.Elements()[0].Value().Effect.IsNull())
	assert.True(t, readState.Policies.Elements()[0].Value().Actions.IsNull())

	unknownPlanState, unknownPlanDiags := NewRoleResourceModelForMutationState(t.Context(), response, RoleModel{
		Permissions: NewTypedMapUnknown[TypedList[types.String]](),
		Policies:    NewTypedListUnknown[TypedObject[RolePolicyValue]](),
	})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanState.Permissions.IsUnknown())
	assert.False(t, unknownPlanState.Policies.IsUnknown())
	assert.True(t, unknownPlanState.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, unknownPlanState.Policies.Elements()[0].Value().Effect.IsNull())

	nullPlanState, nullPlanDiags := NewRoleResourceModelForMutationState(t.Context(), response, RoleModel{
		Permissions: NewTypedMapNull[TypedList[types.String]](),
		Policies:    NewTypedListNull[TypedObject[RolePolicyValue]](),
	})
	assert.False(t, nullPlanDiags.HasError())
	assert.False(t, nullPlanState.Permissions.IsNull())
	assert.False(t, nullPlanState.Policies.IsNull())
	assert.True(t, nullPlanState.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, nullPlanState.Policies.Elements()[0].Value().Effect.IsNull())
}
