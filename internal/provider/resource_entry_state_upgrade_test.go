//nolint:testpackage
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeEntryResourceStateV0ToV1LocalizesFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaV0 := EntryResourceSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := EntryModelV0{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("author"),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"name":  NewNormalizedJSONTypesNormalizedValue([]byte(`{"en-AU":"Name","en-US":"Name"}`)),
			"blurb": NewNormalizedJSONTypesNormalizedValue([]byte(`{"en-AU":{"nodeType":"document","data":{},"content":[]}}`)),
		}),
		Metadata: NewTypedObject[EntryMetadataValue](EntryMetadataValue{
			Concepts: NewTypedListFromStringSlice([]string{}),
			Tags:     NewTypedListFromStringSlice([]string{"blog"}),
		}),
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
	require.False(t, response.Diagnostics.HasError())

	var stateV1 EntryModel
	require.False(t, response.State.Get(ctx, &stateV1).HasError())

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

	upgraders := (&entryResource{}).UpgradeState(context.Background())

	require.Contains(t, upgraders, int64(0))
	assert.NotNil(t, upgraders[0].PriorSchema)
	assert.NotNil(t, upgraders[0].StateUpgrader)
}

func TestUpgradeEntryResourceStateV0ToV1PreservesNullFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaV0 := EntryResourceSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := EntryModelV0{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("author"),
		Fields:             NewTypedMapNull[jsontypes.Normalized](),
		Metadata:           NewTypedObjectNull[EntryMetadataValue](),
		Timeouts:           TimeoutsNull(),
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

	ctx := context.Background()
	schemaV0 := EntryResourceSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := EntryModelV0{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("author"),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"title":       NewNormalizedJSONTypesNormalizedValue([]byte(`[]`)),
			"name":        NewNormalizedJSONTypesNormalizedValue([]byte(`{"en-AU":"Name"}`)),
			"empty-value": jsontypes.NewNormalizedNull(),
		}),
		Metadata: NewTypedObjectNull[EntryMetadataValue](),
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

	var stateV1 EntryModel
	require.False(t, response.State.Get(ctx, &stateV1).HasError())
	assert.NotContains(t, stateV1.Fields.Elements(), "title")
	assert.NotContains(t, stateV1.Fields.Elements(), "empty-value")
	assert.Equal(t, `"Name"`, stateV1.Fields.Elements()["name"].Elements()["en-AU"].ValueString())
}

func TestUpgradeEntryResourceStateV0ToV1PreservesRawNullFieldValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaV0 := EntryResourceSchemaV0(ctx)
	schemaV1 := EntryResourceSchema(ctx)

	stateV0 := EntryModelV0{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("author"),
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"title": NewNormalizedJSONTypesNormalizedValue([]byte(`null`)),
		}),
		Metadata: NewTypedObjectNull[EntryMetadataValue](),
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
	require.False(t, response.Diagnostics.HasError())

	var stateV1 EntryModel
	require.False(t, response.State.Get(ctx, &stateV1).HasError())
	assert.True(t, stateV1.Fields.Elements()["title"].IsNull())
}
