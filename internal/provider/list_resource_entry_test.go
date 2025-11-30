package provider

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/jx"
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmtesting "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryListResource_List(t *testing.T) {
	ctx := context.Background()

	// Create test handler and server
	cmServer, err := cmtesting.NewContentfulManagementServer()
	require.NoError(t, err)
	
	server := httptest.NewServer(cmServer)
	defer server.Close()

	client, err := cm.NewClient(
		server.URL,
		cm.NewAccessTokenSecuritySource("test-token"),
	)
	require.NoError(t, err)

	// Create test data
	spaceID := "test-space"
	environmentID := "test-env"
	contentTypeID := "test-content-type"

	// Add test entries using the server helper
	entryReq1 := cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"title": jx.Raw(`"Entry 1"`),
		}),
	}
	entryReq2 := cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"title": jx.Raw(`"Entry 2"`),
		}),
	}
	entryReq3 := cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"title": jx.Raw(`"Entry 3"`),
		}),
	}

	cmServer.SetEntry(spaceID, environmentID, contentTypeID, "entry-1", entryReq1)
	cmServer.SetEntry(spaceID, environmentID, contentTypeID, "entry-2", entryReq2)
	cmServer.SetEntry(spaceID, environmentID, "other-content-type", "entry-3", entryReq3)

	// Create list resource
	listResource := NewEntryListResource()
	listRes, ok := listResource.(*entryListResource)
	require.True(t, ok)

	// Configure the resource with provider data
	var configResp resource.ConfigureResponse
	listRes.Configure(ctx, resource.ConfigureRequest{
		ProviderData: ContentfulProviderData{
			client: client,
		},
	}, &configResp)
	require.False(t, configResp.Diagnostics.HasError())

	// Get schema
	var schemaResp list.ListResourceSchemaResponse
	listRes.ListResourceConfigSchema(ctx, list.ListResourceSchemaRequest{}, &schemaResp)

	// Get resource schema for request
	var resourceSchemaResp resource.SchemaResponse
	entryResource := NewEntryResource()
	entryResource.Schema(ctx, resource.SchemaRequest{}, &resourceSchemaResp)

	// Get resource identity schema for request
	entryResourceWithIdentity, ok := entryResource.(resource.ResourceWithIdentity)
	require.True(t, ok)
	
	var identitySchemaResp resource.IdentitySchemaResponse
	entryResourceWithIdentity.IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identitySchemaResp)

	// Create request
	attrTypes := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object).AttributeTypes
	configValue := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, map[string]tftypes.Value{
		"space_id":       tftypes.NewValue(tftypes.String, spaceID),
		"environment_id": tftypes.NewValue(tftypes.String, environmentID),
		"content_type":   tftypes.NewValue(tftypes.String, contentTypeID),
		"select":         tftypes.NewValue(tftypes.String, nil),
		"limit":          tftypes.NewValue(tftypes.Number, nil),
		"skip":           tftypes.NewValue(tftypes.Number, nil),
		"order":          tftypes.NewValue(tftypes.String, nil),
		"query":          tftypes.NewValue(tftypes.String, nil),
	})

	req := list.ListRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
		IncludeResource:        true,
		Limit:                  10,
		ResourceSchema:         resourceSchemaResp.Schema,
		ResourceIdentitySchema: identitySchemaResp.IdentitySchema,
	}

	// Create results stream
	var stream list.ListResultsStream

	// Execute list
	listRes.List(ctx, req, &stream)

	// Collect results
	var results []list.ListResult
	for result := range stream.Results {
		results = append(results, result)
		if result.Diagnostics.HasError() {
			t.Fatalf("Got error diagnostics: %v", result.Diagnostics)
		}
	}

	// Verify results
	assert.Len(t, results, 2, "Expected 2 entries matching the content type")

	// Verify first result
	if len(results) > 0 {
		assert.NotNil(t, results[0].Identity)
		assert.NotNil(t, results[0].Resource)
		assert.NotEmpty(t, results[0].DisplayName)
	}
}
