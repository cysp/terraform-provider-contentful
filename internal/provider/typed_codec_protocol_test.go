package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocol6ResourceValidationRejectsMalformedCollectionJSON(t *testing.T) {
	t.Parallel()

	for resourceType, config := range map[string]map[string]any{
		"contentful_entry": {
			"space_id": "space", "environment_id": "master", "content_type_id": "article",
			"fields": map[string]string{"title": "{"},
		},
		"contentful_content_type": {
			"space_id": "space", "environment_id": "master", "content_type_id": "article",
			"name": "Article", "description": "", "display_field": "title",
			"fields": []any{map[string]any{
				"id": "title", "name": "Title", "type": "Symbol", "localized": false, "required": false,
				"validations": []string{"{"},
				"link_type":   nil, "items": nil, "default_value": nil,
				"disabled": nil, "omitted": nil, "allowed_resources": nil,
			}},
		},
	} {
		t.Run(resourceType, func(t *testing.T) {
			t.Parallel()

			server, err := testAccProtoV6ProviderFactories["contentful"]()
			require.NoError(t, err)

			schema, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
			require.NoError(t, err)
			require.Empty(t, schema.Diagnostics)

			resourceSchema, ok := schema.ResourceSchemas[resourceType]
			require.True(t, ok)

			objectType, ok := resourceSchema.ValueType().(tftypes.Object)
			require.True(t, ok)

			// Protocol configuration includes nulls for attributes omitted in HCL.
			for name := range objectType.AttributeTypes {
				if _, present := config[name]; !present {
					config[name] = nil
				}
			}

			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{
				TypeName: resourceType,
				Config:   &tfprotov6.DynamicValue{JSON: []byte(testJSON(config))},
			})
			require.NoError(t, err)
			require.Len(t, response.Diagnostics, 1)
			assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
			assert.Equal(t, "Invalid JSON String Value", response.Diagnostics[0].Summary)
			assert.Contains(t, response.Diagnostics[0].Detail, "not valid JSON string format")
		})
	}
}
