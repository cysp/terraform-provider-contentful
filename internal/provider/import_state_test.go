package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func TestImportStatePassthroughMultipartID(t *testing.T) {
	t.Parallel()

	assertValues := func(t *testing.T, resp *resource.ImportStateResponse, expectIdentity bool) {
		t.Helper()

		var stateSpaceID, stateEntryID types.String
		resp.Diagnostics.Append(resp.State.GetAttribute(t.Context(), path.Root("space_id"), &stateSpaceID)...)
		resp.Diagnostics.Append(resp.State.GetAttribute(t.Context(), path.Root("entry_id"), &stateEntryID)...)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
		assert.Equal(t, types.StringValue("space"), stateSpaceID)
		assert.Equal(t, types.StringValue("entry"), stateEntryID)

		if !expectIdentity {
			assert.Nil(t, resp.Identity)

			return
		}

		var identitySpaceID, identityEntryID types.String
		resp.Diagnostics.Append(resp.Identity.GetAttribute(t.Context(), path.Root("space_id"), &identitySpaceID)...)
		resp.Diagnostics.Append(resp.Identity.GetAttribute(t.Context(), path.Root("entry_id"), &identityEntryID)...)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
		assert.Equal(t, types.StringValue("space"), identitySpaceID)
		assert.Equal(t, types.StringValue("entry"), identityEntryID)
	}

	tests := map[string]struct {
		request        resource.ImportStateRequest
		expectIdentity bool
	}{
		"identity": {
			request: resource.ImportStateRequest{
				Identity: importStatePassthroughIdentity(importStateRawValues("space", "entry")),
			},
			expectIdentity: true,
		},
		"identity without response identity": {
			request: resource.ImportStateRequest{
				Identity: importStatePassthroughIdentity(importStateRawValues("space", "entry")),
			},
		},
		"legacy ID": {
			request:        resource.ImportStateRequest{ID: "space/entry"},
			expectIdentity: true,
		},
		"legacy ID without response identity": {
			request: resource.ImportStateRequest{ID: "space/entry"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := importStatePassthroughResponse()
			if !test.expectIdentity {
				resp.Identity = nil
			}

			provider.ImportStatePassthroughMultipartID(t.Context(), importStatePassthroughPaths, test.request, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
			assertValues(t, resp, test.expectIdentity)
		})
	}
}

func TestImportStatePassthroughMultipartIDFromIdentityRejectsInvalidComponent(t *testing.T) {
	t.Parallel()

	tests := map[string]any{
		"null":    nil,
		"unknown": tftypes.UnknownValue,
		"empty":   "",
	}

	for name, component := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reqIdentity := importStatePassthroughIdentity(importStateRawValues("space", component))
			resp := importStatePassthroughResponse()

			provider.ImportStatePassthroughMultipartID(t.Context(), importStatePassthroughPaths, resource.ImportStateRequest{
				Identity: reqIdentity,
			}, resp)

			assert.True(t, resp.Diagnostics.HasError())
			assert.Equal(t, []string{"entry_id"}, importDiagnosticPaths(t, resp.Diagnostics))
			assertImportStatePassthroughUntouched(t, resp)
		})
	}
}

func TestImportStatePassthroughMultipartIDFromIDRejectsEmptyComponent(t *testing.T) {
	t.Parallel()

	resp := importStatePassthroughResponse()

	provider.ImportStatePassthroughMultipartID(
		t.Context(),
		importStatePassthroughPaths,
		resource.ImportStateRequest{ID: "space/"},
		resp,
	)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, []string{"entry_id"}, importDiagnosticPaths(t, resp.Diagnostics))
	assertImportStatePassthroughUntouched(t, resp)
}

func TestImportStatePassthroughMultipartIDFromIdentityRejectsDecodeError(t *testing.T) {
	t.Parallel()

	rawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
		"entry_id": tftypes.Bool,
	}}
	reqIdentity := &tfsdk.ResourceIdentity{
		Schema: importStatePassthroughIdentitySchema,
		Raw: tftypes.NewValue(rawType, map[string]tftypes.Value{
			"space_id": tftypes.NewValue(tftypes.String, "space"),
			"entry_id": tftypes.NewValue(tftypes.Bool, true),
		}),
	}
	resp := importStatePassthroughResponse()

	provider.ImportStatePassthroughMultipartID(t.Context(), importStatePassthroughPaths, resource.ImportStateRequest{
		Identity: reqIdentity,
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, []string{"entry_id"}, importDiagnosticPaths(t, resp.Diagnostics))
	assertImportStatePassthroughUntouched(t, resp)
}

func TestImportStatePassthroughMultipartIDFromIDIsAtomicAfterLateSetterError(t *testing.T) {
	t.Parallel()

	rawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
	}}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schema.Schema{Attributes: map[string]schema.Attribute{
				"space_id": schema.StringAttribute{Optional: true},
			}},
			Raw: tftypes.NewValue(rawType, map[string]tftypes.Value{
				"space_id": tftypes.NewValue(tftypes.String, nil),
			}),
		},
	}

	provider.ImportStatePassthroughMultipartID(
		t.Context(),
		[]path.Path{path.Root("space_id"), path.Root("missing")},
		resource.ImportStateRequest{ID: "space/entry"},
		resp,
	)

	require.True(t, resp.Diagnostics.HasError())

	var spaceID types.String
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("space_id"), &spaceID).HasError())
	assert.True(t, spaceID.IsNull())
}

func TestImportStatePassthroughMultipartIDFromIdentityIsAtomicAfterLateSetterError(t *testing.T) {
	t.Parallel()

	identityRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
		"missing":  tftypes.String,
	}}
	identitySchema := identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"space_id": identityschema.StringAttribute{RequiredForImport: true},
		"missing":  identityschema.StringAttribute{RequiredForImport: true},
	}}
	reqIdentity := &tfsdk.ResourceIdentity{
		Schema: identitySchema,
		Raw: tftypes.NewValue(identityRawType, map[string]tftypes.Value{
			"space_id": tftypes.NewValue(tftypes.String, "space"),
			"missing":  tftypes.NewValue(tftypes.String, "entry"),
		}),
	}

	stateRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
	}}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schema.Schema{Attributes: map[string]schema.Attribute{
				"space_id": schema.StringAttribute{Optional: true},
			}},
			Raw: tftypes.NewValue(stateRawType, map[string]tftypes.Value{
				"space_id": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		Identity: &tfsdk.ResourceIdentity{
			Schema: identitySchema,
			Raw: tftypes.NewValue(identityRawType, map[string]tftypes.Value{
				"space_id": tftypes.NewValue(tftypes.String, nil),
				"missing":  tftypes.NewValue(tftypes.String, nil),
			}),
		},
	}

	provider.ImportStatePassthroughMultipartID(
		t.Context(),
		[]path.Path{path.Root("space_id"), path.Root("missing")},
		resource.ImportStateRequest{Identity: reqIdentity},
		resp,
	)

	require.True(t, resp.Diagnostics.HasError())

	var stateSpaceID types.String
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("space_id"), &stateSpaceID).HasError())
	assert.True(t, stateSpaceID.IsNull())

	var identitySpaceID, identityMissing types.String
	require.False(t, resp.Identity.GetAttribute(t.Context(), path.Root("space_id"), &identitySpaceID).HasError())
	require.False(t, resp.Identity.GetAttribute(t.Context(), path.Root("missing"), &identityMissing).HasError())
	assert.True(t, identitySpaceID.IsNull())
	assert.True(t, identityMissing.IsNull())
}

func TestImportStatePassthroughMultipartIDIsAtomicAfterIdentitySetterError(t *testing.T) {
	t.Parallel()

	stateRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
		"entry_id": tftypes.String,
	}}
	identityRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"space_id": tftypes.String,
	}}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: importStatePassthroughResourceSchema,
			Raw: tftypes.NewValue(stateRawType, map[string]tftypes.Value{
				"space_id": tftypes.NewValue(tftypes.String, nil),
				"entry_id": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		Identity: &tfsdk.ResourceIdentity{
			Schema: identityschema.Schema{Attributes: map[string]identityschema.Attribute{
				"space_id": identityschema.StringAttribute{RequiredForImport: true},
			}},
			Raw: tftypes.NewValue(identityRawType, map[string]tftypes.Value{
				"space_id": tftypes.NewValue(tftypes.String, nil),
			}),
		},
	}

	provider.ImportStatePassthroughMultipartID(
		t.Context(),
		importStatePassthroughPaths,
		resource.ImportStateRequest{ID: "space/entry"},
		resp,
	)

	require.True(t, resp.Diagnostics.HasError())

	var stateSpaceID, stateEntryID types.String
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("space_id"), &stateSpaceID).HasError())
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("entry_id"), &stateEntryID).HasError())
	assert.True(t, stateSpaceID.IsNull())
	assert.True(t, stateEntryID.IsNull())

	var identitySpaceID types.String
	require.False(t, resp.Identity.GetAttribute(t.Context(), path.Root("space_id"), &identitySpaceID).HasError())
	assert.True(t, identitySpaceID.IsNull())
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

func assertImportStatePassthroughUntouched(t *testing.T, resp *resource.ImportStateResponse) {
	t.Helper()

	var stateSpaceID, stateEntryID types.String
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("space_id"), &stateSpaceID).HasError())
	require.False(t, resp.State.GetAttribute(t.Context(), path.Root("entry_id"), &stateEntryID).HasError())
	assert.True(t, stateSpaceID.IsNull())
	assert.True(t, stateEntryID.IsNull())

	var identitySpaceID, identityEntryID types.String
	require.False(t, resp.Identity.GetAttribute(t.Context(), path.Root("space_id"), &identitySpaceID).HasError())
	require.False(t, resp.Identity.GetAttribute(t.Context(), path.Root("entry_id"), &identityEntryID).HasError())
	assert.True(t, identitySpaceID.IsNull())
	assert.True(t, identityEntryID.IsNull())
}
func importDiagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
