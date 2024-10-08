package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSidebarValueKnown() SidebarValue {
	return SidebarValue{
		state: attr.ValueStateKnown,
	}
}

func (m *SidebarValue) ToPutEditorInterfaceReqSidebarItem(_ context.Context, path path.Path) (contentfulManagement.PutEditorInterfaceReqSidebarItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqSidebarItem{
		WidgetNamespace: m.WidgetNamespace.ValueString(),
		WidgetId:        m.WidgetId.ValueString(),
	}

	modelDisabled := m.Disabled.ValueBoolPointer()
	if modelDisabled != nil {
		item.Disabled.SetTo(*modelDisabled)
	}

	if !m.Settings.IsNull() && !m.Settings.IsUnknown() {
		modelSettings := m.Settings.ValueString()

		path := path.AtName("settings")

		if modelSettings != "" {
			decoder := jx.DecodeStr(modelSettings)

			err := item.Settings.Decode(decoder)
			if err != nil {
				diags.AddAttributeError(path, "Failed to decode settings", err.Error())
			}
		}
	}

	return item, diags
}

func NewSidebarValueFromResponse(_ path.Path, item contentfulManagement.EditorInterfaceSidebarItem) (SidebarValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := SidebarValue{
		WidgetNamespace: types.StringValue(item.WidgetNamespace),
		WidgetId:        types.StringValue(item.WidgetId),
		Disabled:        types.BoolNull(),
		Settings:        types.StringNull(),
		state:           attr.ValueStateKnown,
	}

	if disabled, ok := item.Disabled.Get(); ok {
		value.Disabled = types.BoolValue(disabled)
	}

	if settings, ok := item.Settings.Get(); ok {
		encoder := jx.Encoder{}
		util.EncodeJxRawMapOrdered(&encoder, settings)
		value.Settings = types.StringValue(encoder.String())
	}

	return value, diags
}

func NewSidebarListValueNull(ctx context.Context) types.List {
	return types.ListNull(SidebarValue{}.Type(ctx))
}

func NewSidebarListValueFromResponse(ctx context.Context, path path.Path, sidebarItems []contentfulManagement.EditorInterfaceSidebarItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(sidebarItems))

	for index, item := range sidebarItems {
		sidebarValue, sidebarValueDiags := NewSidebarValueFromResponse(path.AtListIndex(index), item)
		diags.Append(sidebarValueDiags...)

		listElementValues[index] = sidebarValue
	}

	list, listDiags := types.ListValue(SidebarValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
