package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMutationStateRestoresKnownPlanOwnedValues(t *testing.T) {
	t.Parallel()

	plannedFields := contentTypeMutationStateTestFields("planned-field", "Planned field", NewTypedObjectNull[ContentTypeFieldItemsValue]())
	mutationResponse := contentTypeMutationStateTestResponse()

	mutationState, mutationStateDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{
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
	})
	require.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 1)

	assert.Equal(t, types.StringValue("response-space/response-environment/response-content-type"), mutationState.ID)
	assert.Equal(t, types.StringValue("response-space"), mutationState.SpaceID)
	assert.Equal(t, types.StringValue("response-environment"), mutationState.EnvironmentID)
	assert.Equal(t, types.StringValue("response-content-type"), mutationState.ContentTypeID)
	assert.Equal(t, types.Int64Value(7), mutationState.PublishedVersion)
	assert.Equal(t, types.StringValue("Planned name"), mutationState.Name)
	assert.Equal(t, types.StringValue("Planned description"), mutationState.Description)
	assert.Equal(t, types.StringValue("planned-field"), mutationState.DisplayField)
	assert.True(t, mutationState.Fields.Equal(plannedFields))
}

func TestContentTypeMutationStateUsesResponseFieldsWhenPlanFieldsAreNotFullyKnown(t *testing.T) {
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

			mutationState, mutationStateDiags := NewContentTypeResourceModelForMutationState(t.Context(), contentTypeMutationStateTestResponse(), ContentTypeModel{
				Name:         types.StringValue("Planned name"),
				Description:  types.StringValue("Planned description"),
				DisplayField: types.StringValue("planned-field"),
				Fields:       plannedFields,
				Metadata:     NewTypedObjectUnknown[ContentTypeMetadataValue](),
			})
			require.False(t, mutationStateDiags.HasError())
			assert.Len(t, mutationStateDiags.Warnings(), 1)
			assert.False(t, mutationState.Fields.IsUnknown())
			require.Len(t, mutationState.Fields.Elements(), 1)
			assert.Equal(t, "response-field", mutationState.Fields.Elements()[0].Value().ID.ValueString())
			assert.Equal(t, "Response field", mutationState.Fields.Elements()[0].Value().Name.ValueString())
			assert.Equal(t, types.Int64Value(7), mutationState.PublishedVersion)
		})
	}
}

func contentTypeMutationStateTestResponse() cm.ContentType {
	response := cm.ContentType{
		Sys:          cm.NewContentTypeSys("response-space", "response-environment", "response-content-type"),
		Name:         "Response name",
		Description:  cm.NewOptNilString("Response description"),
		DisplayField: cm.NewNilString("response-field"),
		Fields: []cm.ContentTypeFieldsItem{{
			ID:          "response-field",
			Name:        "Response field",
			Type:        "Symbol",
			Validations: []jx.Raw{},
		}},
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkType("future-link-type"),
				},
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

func TestContentTypeMutationStateRestoresKnownPlanOwnedMetadata(t *testing.T) {
	t.Parallel()

	plannedMetadata := NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedNull(),
		Taxonomy: NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
			NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
					ID:       types.StringValue("furniture"),
					Required: types.BoolValue(true),
				}),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			}),
		}),
	})
	mutationResponse := cm.ContentType{
		Sys: cm.NewContentTypeSys("space", "environment", "content_type"),
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkType("future-link-type"),
				},
			}},
		}),
	}

	mutationState, mutationStateDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{Metadata: plannedMetadata})
	assert.False(t, mutationStateDiags.HasError())
	assert.Len(t, mutationStateDiags.Warnings(), 1)
	assert.True(t, mutationState.Metadata.Equal(plannedMetadata))

	readState, readDiags := NewContentTypeResourceModelFromResponse(t.Context(), mutationResponse)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 1)
	assert.True(t, readState.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())
	assert.True(t, readState.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConceptScheme.IsNull())

	unknownPlanState, unknownPlanDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{Metadata: NewTypedObjectUnknown[ContentTypeMetadataValue]()})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanState.Metadata.IsUnknown())
	assert.True(t, unknownPlanState.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	nullPlanState, nullPlanDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{Metadata: NewTypedObjectNull[ContentTypeMetadataValue]()})
	assert.False(t, nullPlanDiags.HasError())
	assert.False(t, nullPlanState.Metadata.IsNull())
	assert.True(t, nullPlanState.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	knownMetadataUnknownTaxonomyState, knownMetadataUnknownTaxonomyDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{
		Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedValue(`{"configured":true}`),
			Taxonomy:    NewTypedListUnknown[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		}),
	})
	assert.False(t, knownMetadataUnknownTaxonomyDiags.HasError())
	assert.Equal(t, `{"configured":true}`, knownMetadataUnknownTaxonomyState.Metadata.Value().Annotations.ValueString())
	assert.False(t, knownMetadataUnknownTaxonomyState.Metadata.Value().Taxonomy.IsUnknown())
	assert.True(t, knownMetadataUnknownTaxonomyState.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	knownMetadataUnknownAnnotationsState, knownMetadataUnknownAnnotationsDiags := NewContentTypeResourceModelForMutationState(t.Context(), mutationResponse, ContentTypeModel{
		Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedUnknown(),
			Taxonomy:    plannedMetadata.Value().Taxonomy,
		}),
	})
	assert.False(t, knownMetadataUnknownAnnotationsDiags.HasError())
	assert.False(t, knownMetadataUnknownAnnotationsState.Metadata.Value().Annotations.IsUnknown())
	assert.True(t, knownMetadataUnknownAnnotationsState.Metadata.Value().Annotations.IsNull())
	assert.True(t, knownMetadataUnknownAnnotationsState.Metadata.Value().Taxonomy.Equal(plannedMetadata.Value().Taxonomy))
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
