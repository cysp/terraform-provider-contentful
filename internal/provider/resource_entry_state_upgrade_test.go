//nolint:testpackage
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeEntryResourceStateV0ToV1LocalizesFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	schemaV0 := entryStateSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := entryStateV0{
		ID: types.StringValue("space/environment/entry"), SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment"), EntryID: types.StringValue("entry"),
		ContentTypeID:    types.StringValue("author"),
		PublishedVersion: types.Int64Value(9),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"omitted": jsontypes.NewNormalizedNull(),
			"name":    NewNormalizedJSONValue([]byte(`{"en-AU":"Name","en-US":"Name"}`)),
			"blurb":   NewNormalizedJSONValue([]byte(`{"en-AU":{"nodeType":"document","data":{},"content":[]}}`)),
		}),
		Metadata: NewTypedObject[entryMetadataV0](entryMetadataV0{
			Concepts: NewTypedListFromStringSlice([]string{}),
			Tags:     NewTypedListFromStringSlice([]string{"blog"}),
		}),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(TimeoutsNull().AttributeTypes(ctx), map[string]attr.Value{"create": types.StringNull(), "read": types.StringNull(), "update": types.StringValue("3m"), "delete": types.StringNull()})},
	}

	priorState := tfsdk.State{Schema: schemaV0}
	require.False(t, priorState.Set(ctx, &stateV0).HasError())

	response := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: schemaV1},
	}

	upgradeEntryResourceStateV0ToV1(ctx, resource.UpgradeStateRequest{
		State: &priorState,
	}, &response)
	require.False(t, response.Diagnostics.HasError())

	var stateV1 EntryModel
	require.False(t, response.State.Get(ctx, &stateV1).HasError())

	require.Equal(t, stateV0.PublishedVersion, stateV1.PublishedVersion)
	require.Equal(t, stateV0.Timeouts, stateV1.Timeouts)
	require.Contains(t, stateV1.Fields.Elements(), "omitted")
	require.True(t, stateV1.Fields.Elements()["omitted"].IsNull())

	name := stateV1.Fields.Elements()["name"].Elements()
	assert.Equal(t, `"Name"`, name["en-AU"].ValueString())
	assert.Equal(t, `"Name"`, name["en-US"].ValueString())

	blurb := stateV1.Fields.Elements()["blurb"].Elements()
	assert.JSONEq(t, `{"content":[],"data":{},"nodeType":"document"}`, blurb["en-AU"].ValueString())
	require.Len(t, stateV1.Metadata.Value().Tags.Elements(), 1)
	assert.Equal(t, "blog", stateV1.Metadata.Value().Tags.Elements()[0].ValueString())
}

func TestEntryResourceUpgradeStateRegistersV0Upgrader(t *testing.T) {
	t.Parallel()

	upgraders := (&entryResource{}).UpgradeState(t.Context())

	require.Contains(t, upgraders, int64(0))
	assert.NotNil(t, upgraders[0].PriorSchema)
	assert.NotNil(t, upgraders[0].StateUpgrader)
}

func TestUpgradeEntryResourceStateV0ToV1PreservesNullFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	schemaV0 := entryStateSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := entryStateV0{
		ID: types.StringValue("space/environment/entry"), SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment"), EntryID: types.StringValue("entry"),
		ContentTypeID:    types.StringValue("author"),
		PublishedVersion: types.Int64Null(),
		Fields:           NewTypedMapNull[jsontypes.Normalized](),
		Metadata:         NewTypedObjectNull[entryMetadataV0](),
		Timeouts:         TimeoutsNull(),
	}

	priorState := tfsdk.State{Schema: schemaV0}
	require.False(t, priorState.Set(ctx, &stateV0).HasError())

	response := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: schemaV1},
	}

	upgradeEntryResourceStateV0ToV1(ctx, resource.UpgradeStateRequest{
		State: &priorState,
	}, &response)
	require.False(t, response.Diagnostics.HasError())

	var stateV1 EntryModel
	require.False(t, response.State.Get(ctx, &stateV1).HasError())
	assert.True(t, stateV1.Fields.IsNull())
}

func TestUpgradeEntryResourceStateV0ToV1ReportsInvalidLocalizedField(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	schemaV0 := entryStateSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := entryStateV0{
		ID: types.StringValue("space/environment/entry"), SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment"), EntryID: types.StringValue("entry"),
		ContentTypeID:    types.StringValue("author"),
		PublishedVersion: types.Int64Null(),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"title":       NewNormalizedJSONValue([]byte(`[]`)),
			"name":        NewNormalizedJSONValue([]byte(`{"en-AU":"Name"}`)),
			"empty-value": jsontypes.NewNormalizedNull(),
		}),
		Metadata: NewTypedObjectNull[entryMetadataV0](),
		Timeouts: TimeoutsNull(),
	}

	priorState := tfsdk.State{Schema: schemaV0}
	require.False(t, priorState.Set(ctx, &stateV0).HasError())

	response := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: schemaV1},
	}

	upgradeEntryResourceStateV0ToV1(ctx, resource.UpgradeStateRequest{
		State: &priorState,
	}, &response)

	require.True(t, response.Diagnostics.HasError())

	assert.True(t, response.State.Raw.IsNull())
}

func TestUpgradeEntryResourceStateV0ToV1RejectsWholeFieldJSONNull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	schemaV0 := entryStateSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := entryStateV0{
		ID: types.StringValue("space/environment/entry"), SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment"), EntryID: types.StringValue("entry"),
		ContentTypeID:    types.StringValue("author"),
		PublishedVersion: types.Int64Null(),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"title": NewNormalizedJSONValue([]byte(`null`)),
		}),
		Metadata: NewTypedObjectNull[entryMetadataV0](),
		Timeouts: TimeoutsNull(),
	}

	priorState := tfsdk.State{Schema: schemaV0}
	require.False(t, priorState.Set(ctx, &stateV0).HasError())

	response := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: schemaV1},
	}

	upgradeEntryResourceStateV0ToV1(ctx, resource.UpgradeStateRequest{
		State: &priorState,
	}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.Equal(t, []string{`fields["title"]`}, attributeDiagnosticPaths(t, response.Diagnostics))
	assert.Contains(t, response.Diagnostics[0].Detail(), "whole-field JSON null")
	assert.True(t, response.State.Raw.IsNull(), "the original state must remain upgradeable")
}
