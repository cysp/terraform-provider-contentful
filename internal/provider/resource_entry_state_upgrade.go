package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EntryModelV0 struct {
	IDIdentityModel
	EntryIdentityModel

	ContentTypeID types.String `tfsdk:"content_type_id"`

	Fields   TypedMap[jsontypes.Normalized]  `tfsdk:"fields"`
	Metadata TypedObject[EntryMetadataValue] `tfsdk:"metadata"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *entryResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := EntryResourceSchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeEntryResourceStateV0ToV1,
		},
	}
}

func upgradeEntryResourceStateV0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateV0 EntryModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &stateV0)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fields := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{})

	if !stateV0.Fields.IsNull() && !stateV0.Fields.IsUnknown() {
		elements := make(map[string]TypedMap[jsontypes.Normalized], len(stateV0.Fields.Elements()))

		for fieldID, fieldValue := range stateV0.Fields.Elements() {
			if fieldValue.IsNull() || fieldValue.IsUnknown() {
				continue
			}

			localizedValues, localizedValuesDiags := NewEntryLocalizedFieldFromRaw(path.Root("fields").AtMapKey(fieldID), []byte(fieldValue.ValueString()))
			resp.Diagnostics.Append(localizedValuesDiags...)

			if localizedValuesDiags.HasError() {
				continue
			}

			elements[fieldID] = localizedValues
		}

		fields = NewTypedMap(elements)
	}

	stateV1 := EntryModel{
		IDIdentityModel:    stateV0.IDIdentityModel,
		EntryIdentityModel: stateV0.EntryIdentityModel,
		ContentTypeID:      stateV0.ContentTypeID,
		Fields:             fields,
		Metadata:           stateV0.Metadata,
		Timeouts:           stateV0.Timeouts,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateV1)...)
}
