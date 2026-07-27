package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryConfigurableResourceCollectionValidatesChildren(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	exemptions := map[string]string{
		"contentful_entry.fields": "entry field values are JSON and explicit JSON null is meaningful",
	}

	for _, resourceFactory := range New("test").Resources(ctx) {
		managedResource := resourceFactory()
		metadataResponse := resource.MetadataResponse{}
		managedResource.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "contentful"}, &metadataResponse)

		schemaResponse := resource.SchemaResponse{}
		managedResource.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
		require.False(t, schemaResponse.Diagnostics.HasError(), schemaResponse.Diagnostics)

		validateConfigurableCollectionAttributes(t, metadataResponse.TypeName, schemaResponse.Schema.Attributes, exemptions)
	}
}

//nolint:cyclop,gocognit
func validateConfigurableCollectionAttributes(
	t *testing.T,
	prefix string,
	attributes map[string]schema.Attribute,
	exemptions map[string]string,
) {
	t.Helper()

	for name, attribute := range attributes {
		attributePath := prefix + "." + name

		switch attribute := attribute.(type) {
		case schema.ListAttribute:
			if attribute.Required || attribute.Optional {
				if _, exempt := exemptions[attributePath]; !exempt {
					t.Run(attributePath, func(t *testing.T) {
						t.Parallel()
						validateListChildren(t, attribute)
					})
				}
			}
		case schema.ListNestedAttribute:
			if attribute.Required || attribute.Optional {
				if _, exempt := exemptions[attributePath]; !exempt {
					t.Run(attributePath, func(t *testing.T) {
						t.Parallel()
						validateListChildren(t, attribute)
					})
				}
			}

			validateConfigurableCollectionAttributes(t, attributePath, attribute.NestedObject.Attributes, exemptions)
		case schema.SetAttribute:
			if attribute.Required || attribute.Optional {
				if _, exempt := exemptions[attributePath]; !exempt {
					t.Run(attributePath, func(t *testing.T) {
						t.Parallel()
						validateSetChildren(t, attribute)
					})
				}
			}
		case schema.SetNestedAttribute:
			validateConfigurableCollectionAttributes(t, attributePath, attribute.NestedObject.Attributes, exemptions)
		case schema.MapAttribute:
			if attribute.Required || attribute.Optional {
				if _, exempt := exemptions[attributePath]; !exempt {
					t.Run(attributePath, func(t *testing.T) {
						t.Parallel()
						validateMapChildren(t, attribute)
						validateNestedMapCollectionChildren(t, attribute)
					})
				}
			}
		case schema.MapNestedAttribute:
			if attribute.Required || attribute.Optional {
				if _, exempt := exemptions[attributePath]; !exempt {
					t.Run(attributePath, func(t *testing.T) {
						t.Parallel()
						validateMapChildren(t, attribute)
					})
				}
			}

			validateConfigurableCollectionAttributes(t, attributePath, attribute.NestedObject.Attributes, exemptions)
		case schema.SingleNestedAttribute:
			validateConfigurableCollectionAttributes(t, attributePath, attribute.Attributes, exemptions)
		}
	}
}

func validateListChildren(t *testing.T, attribute schema.Attribute) {
	t.Helper()

	var elementType attr.Type

	var validators []validator.List

	switch attribute := attribute.(type) {
	case schema.ListAttribute:
		elementType = attribute.ElementType
		validators = attribute.Validators
	case schema.ListNestedAttribute:
		elementType = attribute.NestedObject.Type()
		validators = attribute.Validators
	default:
		require.FailNowf(t, "unexpected attribute type", "got %T, want a list attribute", attribute)
	}

	require.NotEmpty(t, validators)
	assertListDiagnostics(t, validators, types.ListValueMust(elementType, []attr.Value{valueForType(t, elementType, nil)}), true)
	assertListDiagnostics(t, validators, types.ListUnknown(elementType), false)
	assertListDiagnostics(t, validators, types.ListValueMust(elementType, []attr.Value{valueForType(t, elementType, tftypes.UnknownValue)}), false)
}

func validateMapChildren(t *testing.T, attribute schema.Attribute) {
	t.Helper()

	var elementType attr.Type

	var validators []validator.Map

	switch attribute := attribute.(type) {
	case schema.MapAttribute:
		elementType = attribute.ElementType
		validators = attribute.Validators
	case schema.MapNestedAttribute:
		elementType = attribute.NestedObject.Type()
		validators = attribute.Validators
	default:
		require.FailNowf(t, "unexpected attribute type", "got %T, want a map attribute", attribute)
	}

	require.NotEmpty(t, validators)
	assertMapDiagnostics(t, validators, types.MapValueMust(elementType, map[string]attr.Value{"key": valueForType(t, elementType, nil)}), true)
	assertMapDiagnostics(t, validators, types.MapUnknown(elementType), false)
	assertMapDiagnostics(t, validators, types.MapValueMust(elementType, map[string]attr.Value{"key": valueForType(t, elementType, tftypes.UnknownValue)}), false)
}

func validateNestedMapCollectionChildren(t *testing.T, attribute schema.MapAttribute) {
	t.Helper()

	if _, ok := attribute.ElementType.TerraformType(t.Context()).(tftypes.List); !ok {
		return
	}

	nullChild := types.MapValueMust(attribute.ElementType, map[string]attr.Value{
		"key": listValueWithElement(t, attribute.ElementType, nil),
	})
	unknownChild := types.MapValueMust(attribute.ElementType, map[string]attr.Value{
		"key": listValueWithElement(t, attribute.ElementType, tftypes.UnknownValue),
	})

	assertMapDiagnostics(t, attribute.Validators, nullChild, true)
	assertMapDiagnostics(t, attribute.Validators, unknownChild, false)
}

func validateSetChildren(t *testing.T, attribute schema.Attribute) {
	t.Helper()

	setAttribute, ok := attribute.(schema.SetAttribute)
	require.True(t, ok)
	require.NotEmpty(t, setAttribute.Validators)

	assertSetDiagnostics(t, setAttribute.Validators, types.SetValueMust(setAttribute.ElementType, []attr.Value{valueForType(t, setAttribute.ElementType, nil)}), true)
	assertSetDiagnostics(t, setAttribute.Validators, types.SetUnknown(setAttribute.ElementType), false)
	assertSetDiagnostics(t, setAttribute.Validators, types.SetValueMust(setAttribute.ElementType, []attr.Value{valueForType(t, setAttribute.ElementType, tftypes.UnknownValue)}), false)
}

//nolint:ireturn
func valueForType(t *testing.T, valueType attr.Type, value any) attr.Value {
	t.Helper()

	result, err := valueType.ValueFromTerraform(t.Context(), tftypes.NewValue(valueType.TerraformType(t.Context()), value))
	require.NoError(t, err)

	return result
}

//nolint:ireturn
func listValueWithElement(t *testing.T, listType attr.Type, element any) attr.Value {
	t.Helper()

	terraformType, ok := listType.TerraformType(t.Context()).(tftypes.List)
	require.True(t, ok)

	result, err := listType.ValueFromTerraform(t.Context(), tftypes.NewValue(terraformType, []tftypes.Value{
		tftypes.NewValue(terraformType.ElementType, element),
	}))
	require.NoError(t, err)

	return result
}

func assertListDiagnostics(t *testing.T, validators []validator.List, value types.List, expectedError bool) {
	t.Helper()

	response := validator.ListResponse{}
	for _, valueValidator := range validators {
		valueValidator.ValidateList(t.Context(), validator.ListRequest{Path: path.Root("value"), ConfigValue: value}, &response)
	}

	assert.Equal(t, expectedError, response.Diagnostics.HasError(), response.Diagnostics)
}

func assertMapDiagnostics(t *testing.T, validators []validator.Map, value types.Map, expectedError bool) {
	t.Helper()

	response := validator.MapResponse{}
	for _, valueValidator := range validators {
		valueValidator.ValidateMap(t.Context(), validator.MapRequest{Path: path.Root("value"), ConfigValue: value}, &response)
	}

	assert.Equal(t, expectedError, response.Diagnostics.HasError(), response.Diagnostics)
}

func assertSetDiagnostics(t *testing.T, validators []validator.Set, value types.Set, expectedError bool) {
	t.Helper()

	response := validator.SetResponse{}
	for _, valueValidator := range validators {
		valueValidator.ValidateSet(t.Context(), validator.SetRequest{Path: path.Root("value"), ConfigValue: value}, &response)
	}

	assert.Equal(t, expectedError, response.Diagnostics.HasError(), response.Diagnostics)
}
