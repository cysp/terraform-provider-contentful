//nolint:testpackage
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contentTypeCheckpointPrivateData struct{ data map[string][]byte }

func (data *contentTypeCheckpointPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return data.data[key], nil
}

func (data *contentTypeCheckpointPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	data.data[key] = value

	return nil
}

func TestContentTypeActivationRecoveryAddsNoMarkerState(t *testing.T) {
	t.Parallel()

	resourceSchema := ContentTypeResourceSchema(t.Context())
	assert.ElementsMatch(t, []string{
		"id",
		"space_id",
		"environment_id",
		"content_type_id",
		"name",
		"description",
		"display_field",
		"published_version",
		"fields",
		"metadata",
		"timeouts",
	}, mapKeys(resourceSchema.Attributes))
}

func TestContentTypeMutationCheckpointUsesOnlyVersionPrivateKey(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	resourceSchema := ContentTypeResourceSchema(ctx)
	identitySchema := identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"space_id":        identityschema.StringAttribute{RequiredForImport: true},
		"environment_id":  identityschema.StringAttribute{RequiredForImport: true},
		"content_type_id": identityschema.StringAttribute{RequiredForImport: true},
	}}
	state := &tfsdk.State{
		Schema: resourceSchema,
		Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil),
	}
	identity := &tfsdk.ResourceIdentity{
		Schema: identitySchema,
		Raw:    tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil),
	}
	private := &contentTypeCheckpointPrivateData{data: map[string][]byte{}}

	diags := setContentTypeIdentityStateAndVersion(ctx, identity, state, private, contentTypeMutationResult{
		state:   contentTypeLifecycleTestModel(),
		version: 42,
	})

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, map[string][]byte{"version": []byte("42")}, private.data)
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	return keys
}

func TestContentTypeDraftMutationRequired(t *testing.T) {
	t.Parallel()

	base := contentTypeLifecycleTestModel()

	tests := map[string]struct {
		mutatePlan   func(*ContentTypeModel)
		mutateState  func(*ContentTypeModel)
		mutateConfig func(*TypedObject[ContentTypeMetadataValue])
		expected     bool
	}{
		"unchanged": {
			expected: false,
		},
		"name": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Name = types.StringValue("Changed")
			},
			expected: true,
		},
		"description": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Description = types.StringValue("Changed")
			},
			expected: true,
		},
		"display field": {
			mutatePlan: func(model *ContentTypeModel) {
				model.DisplayField = types.StringValue("slug")
			},
			expected: true,
		},
		"fields": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Fields = NewTypedList([]TypedObject[ContentTypeFieldValue]{
					NewTypedObject(ContentTypeFieldValue{}),
				})
			},
			expected: true,
		},
		"field JSON representation only": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Fields = contentTypeLifecycleJSONFields(
					jsontypes.NewNormalizedValue(`{"b":2,"a":1}`),
					jsontypes.NewNormalizedValue(`{"size":{"max":20,"min":1}}`),
				)
			},
			mutateState: func(model *ContentTypeModel) {
				model.Fields = contentTypeLifecycleJSONFields(
					jsontypes.NewNormalizedValue("{\n  \"a\": 1, \"b\": 2\n}"),
					jsontypes.NewNormalizedValue(`{"size":{"min":1,"max":20}}`),
				)
			},
			expected: false,
		},
		"known metadata differs": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"ContentType":[]}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			mutateConfig: func(metadata *TypedObject[ContentTypeMetadataValue]) {
				*metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"ContentType":[]}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			expected: true,
		},
		"known metadata equal": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"ContentType":[]}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			mutateState: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"ContentType":[]}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			mutateConfig: func(metadata *TypedObject[ContentTypeMetadataValue]) {
				*metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"ContentType":[]}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			expected: false,
		},
		"metadata JSON representation only": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"b":2,"a":1}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			mutateState: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue("{\n  \"a\": 1, \"b\": 2\n}"),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			mutateConfig: func(metadata *TypedObject[ContentTypeMetadataValue]) {
				*metadata = NewTypedObject(ContentTypeMetadataValue{
					Annotations: jsontypes.NewNormalizedValue(`{"b":2,"a":1}`),
					Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
				})
			},
			expected: false,
		},
		"unknown planned name": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Name = types.StringUnknown()
			},
			expected: true,
		},
		"unknown planned fields": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Fields = NewTypedListUnknown[TypedObject[ContentTypeFieldValue]]()
			},
			expected: true,
		},
		"unknown planned metadata omitted from configuration": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObjectUnknown[ContentTypeMetadataValue]()
			},
			expected: false,
		},
		"unknown planned metadata explicitly configured": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Metadata = NewTypedObjectUnknown[ContentTypeMetadataValue]()
			},
			mutateConfig: func(metadata *TypedObject[ContentTypeMetadataValue]) {
				*metadata = NewTypedObjectUnknown[ContentTypeMetadataValue]()
			},
			expected: true,
		},
		"id is operational": {
			mutatePlan: func(model *ContentTypeModel) {
				model.ID = types.StringValue("different")
			},
			expected: false,
		},
		"identity is operational": {
			mutatePlan: func(model *ContentTypeModel) {
				model.SpaceID = types.StringValue("different")
				model.EnvironmentID = types.StringValue("different")
				model.ContentTypeID = types.StringValue("different")
			},
			expected: false,
		},
		"published version is operational": {
			mutatePlan: func(model *ContentTypeModel) {
				model.PublishedVersion = types.Int64Value(99)
			},
			expected: false,
		},
		"timeouts are operational": {
			mutatePlan: func(model *ContentTypeModel) {
				model.Timeouts = timeouts.Value{
					Object: types.ObjectUnknown(map[string]attr.Type{
						"create": types.StringType,
						"read":   types.StringType,
						"update": types.StringType,
						"delete": types.StringType,
					}),
				}
			},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan, state := base, base
			configMetadata := NewTypedObjectNull[ContentTypeMetadataValue]()

			if test.mutatePlan != nil {
				test.mutatePlan(&plan)
			}

			if test.mutateState != nil {
				test.mutateState(&state)
			}

			if test.mutateConfig != nil {
				test.mutateConfig(&configMetadata)
			}

			assert.Equal(t, test.expected, contentTypeDraftMutationRequired(configMetadata, plan, state))
		})
	}
}

func TestContentTypeActivationRequired(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		publishedVersion types.Int64
		version          int
		expected         bool
	}{
		"null publication": {
			publishedVersion: types.Int64Null(),
			version:          2,
			expected:         true,
		},
		"legacy null publication without private version": {
			publishedVersion: types.Int64Null(),
			version:          0,
			expected:         true,
		},
		"activated": {
			publishedVersion: types.Int64Value(3),
			version:          4,
			expected:         false,
		},
		"stale publication": {
			publishedVersion: types.Int64Value(2),
			version:          4,
			expected:         true,
		},
		"unknown publication": {
			publishedVersion: types.Int64Unknown(),
			version:          4,
			expected:         false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state := contentTypeLifecycleTestModel()
			state.PublishedVersion = test.publishedVersion

			assert.Equal(t, test.expected, contentTypeActivationRequired(state, test.version))
		})
	}
}

func TestContentTypeDraftMutationRequiredUnknownMetadataWithState(t *testing.T) {
	t.Parallel()

	state := contentTypeLifecycleTestModel()
	state.Metadata = NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedNull(),
		Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
	})
	plan := state
	plan.Metadata = NewTypedObjectUnknown[ContentTypeMetadataValue]()

	assert.True(t, contentTypeDraftMutationRequired(NewTypedObjectUnknown[ContentTypeMetadataValue](), plan, state))
}

func contentTypeLifecycleTestModel() ContentTypeModel {
	return ContentTypeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "environment", "content-type"),
		ContentTypeIdentityModel: ContentTypeIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			ContentTypeID: types.StringValue("content-type"),
		},
		Name:             types.StringValue("Test"),
		Description:      types.StringValue("Description"),
		DisplayField:     types.StringValue("name"),
		PublishedVersion: types.Int64Value(1),
		Fields:           NewTypedList([]TypedObject[ContentTypeFieldValue]{}),
		Metadata:         NewTypedObjectNull[ContentTypeMetadataValue](),
		Timeouts:         TimeoutsNull(),
	}
}

func contentTypeLifecycleJSONFields(defaultValue, validation jsontypes.Normalized) TypedList[TypedObject[ContentTypeFieldValue]] {
	return NewTypedList([]TypedObject[ContentTypeFieldValue]{
		NewTypedObject(ContentTypeFieldValue{
			ID:               types.StringValue("value"),
			Name:             types.StringValue("Value"),
			FieldType:        types.StringValue("Object"),
			LinkType:         types.StringNull(),
			Disabled:         types.BoolValue(false),
			Omitted:          types.BoolValue(false),
			Required:         types.BoolValue(false),
			DefaultValue:     defaultValue,
			Items:            NewTypedObjectNull[ContentTypeFieldItemsValue](),
			Localized:        types.BoolValue(false),
			Validations:      NewTypedList([]jsontypes.Normalized{validation}),
			AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
		}),
	})
}
