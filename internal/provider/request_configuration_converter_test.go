package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestConvertersRejectUnknownPlannedConfigurationOwnedValues(t *testing.T) {
	t.Parallel()

	t.Run("app definition scalar", func(t *testing.T) {
		t.Parallel()

		model := validAppDefinitionRequestModel()
		actual, diags := model.ToAppDefinitionData(AppDefinitionBaseModel{
			Src: types.StringValue("https://configured.example/app.js"),
		}, path.Empty())

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"src"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.AppDefinitionData{}, actual)
	})

	t.Run("extension nested scalar", func(t *testing.T) {
		t.Parallel()

		model := validExtensionRequestModel()
		actual, diags := model.ToExtensionData(ExtensionModel{
			Extension: &ExtensionModelExtension{Src: types.StringValue("https://configured.example/extension.js")},
		}, path.Empty())

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"extension.src"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.ExtensionData{}, actual)
	})

	t.Run("extension JSON", func(t *testing.T) {
		t.Parallel()

		model := validExtensionRequestModel()
		actual, diags := model.ToExtensionData(ExtensionModel{
			Parameters: NewNormalizedJSONTypesNormalizedValue([]byte(`{}`)),
		}, path.Empty())

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"parameters"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.ExtensionData{}, actual)
	})

	t.Run("webhook map", func(t *testing.T) {
		t.Parallel()

		model := validWebhookRequestModel()
		model.Headers = NewTypedMapUnknown[TypedObject[WebhookHeaderValue]]()
		actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{
			Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
		}, path.Empty())

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"headers"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.WebhookDefinitionData{}, actual)
	})

	t.Run("space enablement bool", func(t *testing.T) {
		t.Parallel()

		model := SpaceEnablementsModel{
			CrossSpaceLinks: types.BoolUnknown(),
			SpaceTemplates:  types.BoolValue(false),
		}
		actual, diags := model.ToSpaceEnablementData(t.Context(), SpaceEnablementsModel{
			CrossSpaceLinks: types.BoolValue(false),
			SpaceTemplates:  types.BoolValue(false),
		})

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"cross_space_links"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.SpaceEnablementData{}, actual)
	})

	t.Run("content type metadata object", func(t *testing.T) {
		t.Parallel()

		model := validContentTypeRequestModel()
		model.Metadata = NewTypedObjectUnknown[ContentTypeMetadataValue]()
		actual, diags := model.ToContentTypeRequestData(t.Context(), ContentTypeModel{
			Metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedNull(),
				Taxonomy:    NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{}),
			}),
		})

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"metadata"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.ContentTypeRequestData{}, actual)
	})

	t.Run("content type taxonomy list", func(t *testing.T) {
		t.Parallel()

		metadata := NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    NewTypedListUnknown[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		})
		model := validContentTypeRequestModel()
		model.Metadata = metadata
		actual, diags := model.ToContentTypeRequestData(t.Context(), ContentTypeModel{Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{}),
		})})

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"metadata.taxonomy"}, attributeDiagnosticPaths(t, diags))
		assert.Equal(t, cm.ContentTypeRequestData{}, actual)
	})
}

func TestConfigurationAwareConvertersPreserveKnownEmptyPlanValues(t *testing.T) {
	t.Parallel()

	t.Run("empty JSON object", func(t *testing.T) {
		t.Parallel()

		model := validExtensionRequestModel()
		model.Parameters = NewNormalizedJSONTypesNormalizedValue([]byte(`{}`))

		actual, diags := model.ToExtensionData(ExtensionModel{Parameters: jsontypes.NewNormalizedNull()}, path.Empty())

		require.False(t, diags.HasError(), diags.Errors())
		assert.JSONEq(t, `{}`, string(actual.Parameters))
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		emptyTaxonomy := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{})
		model := validContentTypeRequestModel()
		model.Metadata = NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    emptyTaxonomy,
		})

		actual, diags := model.ToContentTypeRequestData(t.Context(), ContentTypeModel{Metadata: NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    emptyTaxonomy,
		})})

		require.False(t, diags.HasError(), diags.Errors())

		metadata, ok := actual.Metadata.Get()
		require.True(t, ok)
		assert.NotNil(t, metadata.Taxonomy)
		assert.Empty(t, metadata.Taxonomy)
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		emptyHeaders := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{})
		model := validWebhookRequestModel()
		model.Headers = emptyHeaders

		actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{Headers: emptyHeaders}, path.Empty())

		require.False(t, diags.HasError(), diags.Errors())
		assert.NotNil(t, actual.Headers)
		assert.Empty(t, actual.Headers)
	})

	t.Run("false boolean", func(t *testing.T) {
		t.Parallel()

		model := SpaceEnablementsModel{
			CrossSpaceLinks: types.BoolValue(false),
			SpaceTemplates:  types.BoolValue(false),
		}
		actual, diags := model.ToSpaceEnablementData(t.Context(), model)

		require.False(t, diags.HasError(), diags.Errors())

		crossSpaceLinks, ok := actual.CrossSpaceLinks.Get()
		require.True(t, ok)
		assert.False(t, crossSpaceLinks.Enabled)
		spaceTemplates, ok := actual.SpaceTemplates.Get()
		require.True(t, ok)
		assert.False(t, spaceTemplates.Enabled)
	})
}
