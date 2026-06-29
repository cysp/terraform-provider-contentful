package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	importStatePassthroughPaths          = []path.Path{path.Root("space_id"), path.Root("entry_id")}
	importStatePassthroughResourceSchema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{Optional: true},
			"entry_id": schema.StringAttribute{Optional: true},
		},
	}
	importStatePassthroughIdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id": identityschema.StringAttribute{RequiredForImport: true},
			"entry_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
	importStatePassthroughRawType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
		"entry_id": tftypes.String,
	}}
)

func TestImportStatePassthroughMultipartIDFromIdentity(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	reqIdentity := importStatePassthroughIdentity(importStateRawValues("space", "entry"))
	resp := importStatePassthroughResponse()

	provider.ImportStatePassthroughMultipartID(ctx, importStatePassthroughPaths, resource.ImportStateRequest{
		Identity: reqIdentity,
	}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())

	var stateSpaceID, stateEntryID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &stateSpaceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entry_id"), &stateEntryID)...)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
	assert.Equal(t, types.StringValue("space"), stateSpaceID)
	assert.Equal(t, types.StringValue("entry"), stateEntryID)

	var identitySpaceID, identityEntryID types.String
	resp.Diagnostics.Append(resp.Identity.GetAttribute(ctx, path.Root("space_id"), &identitySpaceID)...)
	resp.Diagnostics.Append(resp.Identity.GetAttribute(ctx, path.Root("entry_id"), &identityEntryID)...)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
	assert.Equal(t, types.StringValue("space"), identitySpaceID)
	assert.Equal(t, types.StringValue("entry"), identityEntryID)
}

func TestImportStatePassthroughMultipartIDFromIdentityRejectsNullComponent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	reqIdentity := importStatePassthroughIdentity(importStateRawValues("space", nil))
	resp := importStatePassthroughResponse()

	provider.ImportStatePassthroughMultipartID(ctx, importStatePassthroughPaths, resource.ImportStateRequest{
		Identity: reqIdentity,
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func importStatePassthroughIdentity(rawValues map[string]tftypes.Value) *tfsdk.ResourceIdentity {
	return &tfsdk.ResourceIdentity{
		Schema: importStatePassthroughIdentitySchema,
		Raw:    tftypes.NewValue(importStatePassthroughRawType, rawValues),
	}
}

func importStatePassthroughResponse() *resource.ImportStateResponse {
	return &resource.ImportStateResponse{
		State:    tfsdk.State{Schema: importStatePassthroughResourceSchema, Raw: tftypes.NewValue(importStatePassthroughRawType, importStateRawValues(nil, nil))},
		Identity: importStatePassthroughIdentity(importStateRawValues(nil, nil)),
	}
}

func importStateRawValues(spaceID, entryID any) map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"space_id": tftypes.NewValue(tftypes.String, spaceID),
		"entry_id": tftypes.NewValue(tftypes.String, entryID),
	}
}
