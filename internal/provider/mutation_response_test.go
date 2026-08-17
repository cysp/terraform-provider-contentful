package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeMutationResponsePreservesKnownMetadataAfterLossyProjection(t *testing.T) {
	t.Parallel()

	configuredMetadata := NewTypedObject(ContentTypeMetadataValue{
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
	response := cm.ContentType{
		Sys: cm.NewContentTypeSys("space", "environment", "content_type"),
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkType("future-link-type"),
				},
			}},
		}),
	}

	mutationModel, mutationDiags := NewContentTypeResourceModelFromMutationResponse(t.Context(), response, ContentTypeModel{Metadata: configuredMetadata})
	assert.False(t, mutationDiags.HasError())
	assert.Len(t, mutationDiags.Warnings(), 1)
	assert.True(t, mutationModel.Metadata.Equal(configuredMetadata))

	readModel, readDiags := NewContentTypeResourceModelFromResponse(t.Context(), response)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 1)
	assert.True(t, readModel.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())
	assert.True(t, readModel.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConceptScheme.IsNull())

	unknownPlanModel, unknownPlanDiags := NewContentTypeResourceModelFromMutationResponse(t.Context(), response, ContentTypeModel{Metadata: NewTypedObjectUnknown[ContentTypeMetadataValue]()})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanModel.Metadata.IsUnknown())
	assert.True(t, unknownPlanModel.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	nullPlanModel, nullPlanDiags := NewContentTypeResourceModelFromMutationResponse(t.Context(), response, ContentTypeModel{Metadata: NewTypedObjectNull[ContentTypeMetadataValue]()})
	assert.False(t, nullPlanDiags.HasError())
	assert.False(t, nullPlanModel.Metadata.IsNull())
	assert.True(t, nullPlanModel.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	knownMetadataUnknownTaxonomyModel, knownMetadataUnknownTaxonomyDiags := NewContentTypeResourceModelFromMutationResponse(t.Context(), response, ContentTypeModel{
		Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedValue(`{"configured":true}`),
			Taxonomy:    NewTypedListUnknown[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		}),
	})
	assert.False(t, knownMetadataUnknownTaxonomyDiags.HasError())
	assert.Equal(t, `{"configured":true}`, knownMetadataUnknownTaxonomyModel.Metadata.Value().Annotations.ValueString())
	assert.False(t, knownMetadataUnknownTaxonomyModel.Metadata.Value().Taxonomy.IsUnknown())
	assert.True(t, knownMetadataUnknownTaxonomyModel.Metadata.Value().Taxonomy.Elements()[0].Value().TaxonomyConcept.IsNull())

	knownMetadataUnknownAnnotationsModel, knownMetadataUnknownAnnotationsDiags := NewContentTypeResourceModelFromMutationResponse(t.Context(), response, ContentTypeModel{
		Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedUnknown(),
			Taxonomy:    configuredMetadata.Value().Taxonomy,
		}),
	})
	assert.False(t, knownMetadataUnknownAnnotationsDiags.HasError())
	assert.False(t, knownMetadataUnknownAnnotationsModel.Metadata.Value().Annotations.IsUnknown())
	assert.True(t, knownMetadataUnknownAnnotationsModel.Metadata.Value().Annotations.IsNull())
	assert.True(t, knownMetadataUnknownAnnotationsModel.Metadata.Value().Taxonomy.Equal(configuredMetadata.Value().Taxonomy))
}

func TestEditorInterfaceMutationResponsePreservesKnownLayoutAfterLossyProjection(t *testing.T) {
	t.Parallel()

	configuredLayout := NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
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

	mutationModel, mutationDiags := NewEditorInterfaceResourceModelFromMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: configuredLayout})
	assert.False(t, mutationDiags.HasError())
	assert.Len(t, mutationDiags.Warnings(), 1)
	assert.True(t, mutationModel.EditorLayout.Equal(configuredLayout))

	readModel, readDiags := NewEditorInterfaceResourceModelFromResponse(t.Context(), response)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 1)
	assert.True(t, readModel.EditorLayout.Elements()[0].Value().Group.IsNull())

	unknownPlanModel, unknownPlanDiags := NewEditorInterfaceResourceModelFromMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanModel.EditorLayout.IsUnknown())
	assert.True(t, unknownPlanModel.EditorLayout.Elements()[0].Value().Group.IsNull())

	nullPlanModel, nullPlanDiags := NewEditorInterfaceResourceModelFromMutationResponse(t.Context(), response, EditorInterfaceModel{EditorLayout: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()})
	assert.False(t, nullPlanDiags.HasError())
	assert.True(t, nullPlanModel.EditorLayout.IsNull())
}

func TestRoleMutationResponsePreservesKnownPermissionsAndPoliciesAfterLossyProjection(t *testing.T) {
	t.Parallel()

	configuredPermissions := NewTypedMap(map[string]TypedList[types.String]{
		"Entry": NewTypedList([]types.String{types.StringValue("read")}),
	})
	configuredPolicies := NewTypedList([]TypedObject[RolePolicyValue]{
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

	mutationModel, mutationDiags := NewRoleResourceModelFromMutationResponse(t.Context(), response, RoleModel{Permissions: configuredPermissions, Policies: configuredPolicies})
	assert.False(t, mutationDiags.HasError())
	assert.Len(t, mutationDiags.Warnings(), 3)
	assert.True(t, mutationModel.Permissions.Equal(configuredPermissions))
	assert.True(t, mutationModel.Policies.Equal(configuredPolicies))

	readModel, readDiags := NewRoleResourceModelFromResponse(t.Context(), response)
	assert.False(t, readDiags.HasError())
	assert.Len(t, readDiags.Warnings(), 3)
	assert.True(t, readModel.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, readModel.Policies.Elements()[0].Value().Effect.IsNull())
	assert.True(t, readModel.Policies.Elements()[0].Value().Actions.IsNull())

	unknownPlanModel, unknownPlanDiags := NewRoleResourceModelFromMutationResponse(t.Context(), response, RoleModel{
		Permissions: NewTypedMapUnknown[TypedList[types.String]](),
		Policies:    NewTypedListUnknown[TypedObject[RolePolicyValue]](),
	})
	assert.False(t, unknownPlanDiags.HasError())
	assert.False(t, unknownPlanModel.Permissions.IsUnknown())
	assert.False(t, unknownPlanModel.Policies.IsUnknown())
	assert.True(t, unknownPlanModel.Permissions.Elements()["Entry"].IsNull())
	assert.True(t, unknownPlanModel.Policies.Elements()[0].Value().Effect.IsNull())
}
