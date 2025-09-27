package provider_test

import (
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContentTypeListResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestFields{
		Name:   "Author",
		Fields: []cm.ContentTypeRequestFieldsFieldsItem{},
	})

	testserver := httptest.NewServer(server)
	t.Cleanup(testserver.Close)

	p := New("test", ContentfulProviderOptionsWithHTTPTestServer(testserver)...)

	var providerSchemaResponse provider.SchemaResponse
	p.Schema(t.Context(), provider.SchemaRequest{}, &providerSchemaResponse)

	providerConfigValue := types.ObjectValueMust(map[string]attr.Type{
		"url":          types.StringType,
		"access_token": types.StringType,
	}, map[string]attr.Value{
		"url":          types.StringNull(),
		"access_token": types.StringNull(),
	})

	providerConfigValueTerraformValue, providerConfigValueTerraformValueErr := providerConfigValue.ToTerraformValue(t.Context())
	require.NoError(t, providerConfigValueTerraformValueErr)

	var providerConfigureResponse provider.ConfigureResponse
	p.Configure(t.Context(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    providerConfigValueTerraformValue,
			Schema: providerSchemaResponse.Schema,
		},
	}, &providerConfigureResponse)

	contentTypeResource := NewContentTypeResource()

	var contentTypeResourceConfigureResponse resource.ConfigureResponse
	contentTypeResource.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: providerConfigureResponse.ResourceData,
	}, &contentTypeResourceConfigureResponse)

	var contentTypeResourceIdentitySchemaResponse resource.IdentitySchemaResponse
	contentTypeResource.(resource.ResourceWithIdentity).IdentitySchema(t.Context(), resource.IdentitySchemaRequest{}, &contentTypeResourceIdentitySchemaResponse)

	var contentTypeResourceSchemaResponse resource.SchemaResponse
	contentTypeResource.Schema(t.Context(), resource.SchemaRequest{}, &contentTypeResourceSchemaResponse)

	contentTypeListResource := NewContentTypeListResource()

	var contentTypeListResourceConfigureResponse resource.ConfigureResponse
	contentTypeListResource.(list.ListResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: providerConfigureResponse.ListResourceData,
	}, &contentTypeListResourceConfigureResponse)

	var contentTypeListResourceSchemaResponse list.ListResourceSchemaResponse
	contentTypeListResource.ListResourceConfigSchema(t.Context(), list.ListResourceSchemaRequest{}, &contentTypeListResourceSchemaResponse)

	requestConfigValue, requestConfigValueDiags := types.ObjectValue(map[string]attr.Type{
		"space_id":       types.StringType,
		"environment_id": types.StringType,
	}, map[string]attr.Value{
		"space_id":       types.StringValue("0p38pssr0fi3"),
		"environment_id": types.StringValue("test"),
	})
	require.Empty(t, requestConfigValueDiags)

	requestConfigValueTerraformValue, requestConfigValueTerraformValueErr := requestConfigValue.ToTerraformValue(t.Context())
	require.NoError(t, requestConfigValueTerraformValueErr)

	request := list.ListRequest{
		Config: tfsdk.Config{
			Raw:    requestConfigValueTerraformValue,
			Schema: contentTypeListResourceSchemaResponse.Schema,
		},
		ResourceIdentitySchema: contentTypeResourceIdentitySchemaResponse.IdentitySchema,
		ResourceSchema:         contentTypeResourceSchemaResponse.Schema,
	}

	var listResultsStream list.ListResultsStream
	contentTypeListResource.List(t.Context(), request, &listResultsStream)

	for result := range listResultsStream.Results {
		require.NotNil(t, result.Diagnostics)
	}
}

func TestContentTypeListResourceNotFoundEnvironment(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	testserver := httptest.NewServer(server)
	t.Cleanup(testserver.Close)

	p := New("test", ContentfulProviderOptionsWithHTTPTestServer(testserver)...)

	var providerSchemaResponse provider.SchemaResponse
	p.Schema(t.Context(), provider.SchemaRequest{}, &providerSchemaResponse)

	providerConfigValue := types.ObjectValueMust(map[string]attr.Type{
		"url":          types.StringType,
		"access_token": types.StringType,
	}, map[string]attr.Value{
		"url":          types.StringNull(),
		"access_token": types.StringNull(),
	})

	providerConfigValueTerraformValue, providerConfigValueTerraformValueErr := providerConfigValue.ToTerraformValue(t.Context())
	require.NoError(t, providerConfigValueTerraformValueErr)

	var providerConfigureResponse provider.ConfigureResponse
	p.Configure(t.Context(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    providerConfigValueTerraformValue,
			Schema: providerSchemaResponse.Schema,
		},
	}, &providerConfigureResponse)

	contentTypeListResource := NewContentTypeListResource()

	var contentTypeListResourceConfigureResponse resource.ConfigureResponse
	contentTypeListResource.(list.ListResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: providerConfigureResponse.ListResourceData,
	}, &contentTypeListResourceConfigureResponse)

	var contentTypeListResourceSchemaResponse list.ListResourceSchemaResponse
	contentTypeListResource.ListResourceConfigSchema(t.Context(), list.ListResourceSchemaRequest{}, &contentTypeListResourceSchemaResponse)

	requestConfigValue, requestConfigValueDiags := types.ObjectValue(map[string]attr.Type{
		"space_id":       types.StringType,
		"environment_id": types.StringType,
	}, map[string]attr.Value{
		"space_id":       types.StringValue("0p38pssr0fi3"),
		"environment_id": types.StringValue("nonexistent"),
	})
	require.Empty(t, requestConfigValueDiags)

	requestConfigValueTerraformValue, requestConfigValueTerraformValueErr := requestConfigValue.ToTerraformValue(t.Context())
	require.NoError(t, requestConfigValueTerraformValueErr)

	request := list.ListRequest{
		Config: tfsdk.Config{
			Raw:    requestConfigValueTerraformValue,
			Schema: contentTypeListResourceSchemaResponse.Schema,
		},
	}

	var listResultsStream list.ListResultsStream
	contentTypeListResource.List(t.Context(), request, &listResultsStream)

	for result := range listResultsStream.Results {
		require.NotNil(t, result.Diagnostics)
	}
}
