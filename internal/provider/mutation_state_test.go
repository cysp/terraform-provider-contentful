package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
