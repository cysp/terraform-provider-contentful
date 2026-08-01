//nolint:testpackage
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListResultConversionErrorLeavesResultUnpublished(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	implementation := contentTypeResource{}

	schemaResponse := resource.SchemaResponse{}
	implementation.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	identityResponse := resource.IdentitySchemaResponse{}
	implementation.IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identityResponse)

	req := list.ListRequest{
		IncludeResource:        true,
		ResourceSchema:         schemaResponse.Schema,
		ResourceIdentitySchema: identityResponse.IdentitySchema,
	}
	result := newListResultFromResponse(
		ctx,
		req,
		"malformed",
		ContentTypeIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			ContentTypeID: types.StringValue("content-type"),
		},
		func() (ContentTypeModel, diag.Diagnostics) {
			return ContentTypeModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Malformed Contentful response", "conversion failed")}
		},
	)

	require.True(t, result.Diagnostics.HasError())
	assert.True(t, result.Resource.Raw.IsNull())
	assert.True(t, result.Identity.Raw.IsNull())
}

func TestListResultSetterErrorLeavesResultUnpublished(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	implementation := contentTypeResource{}

	schemaResponse := resource.SchemaResponse{}
	implementation.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	identityResponse := resource.IdentitySchemaResponse{}
	implementation.IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identityResponse)

	req := list.ListRequest{
		IncludeResource:        true,
		ResourceSchema:         schemaResponse.Schema,
		ResourceIdentitySchema: identityResponse.IdentitySchema,
	}

	for name, test := range map[string]struct {
		identity any
		convert  func() (any, diag.Diagnostics)
	}{
		"identity": {
			identity: struct{}{},
			convert: func() (any, diag.Diagnostics) {
				return ContentTypeModel{}, nil
			},
		},
		"resource": {
			identity: ContentTypeIdentityModel{
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content-type"),
			},
			convert: func() (any, diag.Diagnostics) {
				return struct{}{}, nil
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := newListResultFromResponse(ctx, req, "malformed", test.identity, test.convert)

			require.True(t, result.Diagnostics.HasError())
			assert.True(t, result.Identity.Raw.IsNull())
			assert.True(t, result.Resource.Raw.IsNull())
		})
	}
}
