package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The v0 decoder uses the historical field, metadata and timeout shapes. It
// needs no planning defaults, validators or modifiers.
func entryStateSchemaV0(ctx context.Context) schema.Schema {
	stringList := schema.ListAttribute{Optional: true, ElementType: types.StringType, CustomType: NewTypedListNull[types.String]().CustomType(ctx)}

	return schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                schema.StringAttribute{Computed: true},
		"space_id":          schema.StringAttribute{Required: true},
		"environment_id":    schema.StringAttribute{Required: true},
		"entry_id":          schema.StringAttribute{Optional: true, Computed: true},
		"content_type_id":   schema.StringAttribute{Required: true},
		"published_version": schema.Int64Attribute{Computed: true},
		"fields": schema.MapAttribute{
			Required:    true,
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedMapNull[jsontypes.Normalized]().CustomType(ctx),
		},
		"metadata": schema.SingleNestedAttribute{
			Optional: true, Computed: true,
			CustomType: NewTypedObjectNull[entryMetadataV0]().CustomType(ctx),
			Attributes: map[string]schema.Attribute{"concepts": stringList, "tags": stringList},
		},
		"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Update: true, Delete: true}),
	}}
}

type entryMetadataV0 struct {
	Concepts TypedList[types.String] `tfsdk:"concepts"`
	Tags     TypedList[types.String] `tfsdk:"tags"`
}

type entryStateV0 struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	EntryID       types.String `tfsdk:"entry_id"`

	ContentTypeID    types.String `tfsdk:"content_type_id"`
	PublishedVersion types.Int64  `tfsdk:"published_version"`

	Fields   TypedMap[jsontypes.Normalized] `tfsdk:"fields"`
	Metadata TypedObject[entryMetadataV0]   `tfsdk:"metadata"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *entryResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := entryStateSchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeEntryResourceStateV0ToV1,
		},
	}
}

func upgradeEntryResourceStateV0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateV0 entryStateV0
	resp.Diagnostics.Append(req.State.Get(ctx, &stateV0)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fields := NewTypedMapNull[TypedMap[jsontypes.Normalized]]()

	if stateV0.Fields.IsUnknown() {
		fields = NewTypedMapUnknown[TypedMap[jsontypes.Normalized]]()
	} else if !stateV0.Fields.IsNull() {
		elements := make(map[string]TypedMap[jsontypes.Normalized], len(stateV0.Fields.Elements()))

		for fieldID, fieldValue := range stateV0.Fields.Elements() {
			if fieldValue.IsNull() {
				elements[fieldID] = NewTypedMapNull[jsontypes.Normalized]()

				continue
			}

			if fieldValue.IsUnknown() {
				elements[fieldID] = NewTypedMapUnknown[jsontypes.Normalized]()

				continue
			}

			localizedValues, localizedValuesDiags := entryLocalizedFieldFromJSON(path.Root("fields").AtMapKey(fieldID), []byte(fieldValue.ValueString()))
			resp.Diagnostics.Append(localizedValuesDiags...)

			if localizedValuesDiags.HasError() {
				continue
			}

			elements[fieldID] = localizedValues
		}

		fields = NewTypedMap(elements)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	metadata := NewTypedObjectNull[EntryMetadataValue]()
	if stateV0.Metadata.IsUnknown() {
		metadata = NewTypedObjectUnknown[EntryMetadataValue]()
	} else if !stateV0.Metadata.IsNull() {
		prior := stateV0.Metadata.Value()
		metadata = NewTypedObject(EntryMetadataValue(prior))
	}

	stateV1 := EntryModel{
		ID:      stateV0.ID,
		SpaceID: stateV0.SpaceID, EnvironmentID: stateV0.EnvironmentID, EntryID: stateV0.EntryID,
		ContentTypeID:    stateV0.ContentTypeID,
		PublishedVersion: stateV0.PublishedVersion,
		Fields:           fields,
		Metadata:         metadata,
		Timeouts:         stateV0.Timeouts,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateV1)...)
}
