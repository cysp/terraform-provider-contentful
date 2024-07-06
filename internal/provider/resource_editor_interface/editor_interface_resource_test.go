package resource_editor_interface_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_editor_interface"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPutEditorInterfaceReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	controlsValue1 := resource_editor_interface.NewControlsValueKnown()
	controlsValue1.FieldId = types.StringValue("field_id")
	controlsValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue1.WidgetId = types.StringValue("widget_id")
	controlsValue1.Settings = types.StringValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(resource_editor_interface.ControlsValue{}.Type(ctx), []attr.Value{
		controlsValue1,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := resource_editor_interface.NewSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetId = types.StringValue("widget_id")
	sidebarValue1.Settings = types.StringValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(resource_editor_interface.SidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
	})

	require.Empty(t, sidebarDiags)

	model := resource_editor_interface.EditorInterfaceModel{
		SpaceId:       types.StringValue("space_id"),
		EnvironmentId: types.StringValue("environment_id"),
		ContentTypeId: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	diags := diag.Diagnostics{}
	req := model.ToPutEditorInterfaceReq(ctx, &diags)

	assert.Empty(t, diags)

	assert.EqualValues(t, contentfulManagement.PutEditorInterfaceReq{
		Controls: []contentfulManagement.PutEditorInterfaceReqControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings: contentfulManagement.OptPutEditorInterfaceReqControlsItemSettings{
					Set: true,
					Value: map[string]jx.Raw{
						"foo": jx.Raw(`"bar"`),
					},
				},
			},
		},
		Sidebar: []contentfulManagement.PutEditorInterfaceReqSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings: contentfulManagement.OptPutEditorInterfaceReqSidebarItemSettings{
					Set: true,
					Value: map[string]jx.Raw{
						"foo": jx.Raw(`"bar"`),
					},
				},
			},
		},
	}, req)
}

func TestToPutEditorInterfaceReqErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	controlsValue1 := resource_editor_interface.NewControlsValueKnown()
	controlsValue1.FieldId = types.StringValue("field_id")
	controlsValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue1.WidgetId = types.StringValue("widget_id")
	controlsValue1.Settings = types.StringNull()

	controlsValue2 := resource_editor_interface.NewControlsValueKnown()
	controlsValue2.FieldId = types.StringValue("field_id")
	controlsValue2.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue2.WidgetId = types.StringValue("widget_id")
	controlsValue2.Settings = types.StringValue(`invalid json`)

	controlsValue3 := resource_editor_interface.NewControlsValueKnown()
	controlsValue3.FieldId = types.StringValue("field_id")
	controlsValue3.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue3.WidgetId = types.StringValue("widget_id")
	controlsValue3.Settings = types.StringValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(resource_editor_interface.ControlsValue{}.Type(ctx), []attr.Value{
		controlsValue1,
		controlsValue2,
		controlsValue3,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := resource_editor_interface.NewSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetId = types.StringValue("widget_id")
	sidebarValue1.Settings = types.StringNull()

	sidebarValue2 := resource_editor_interface.NewSidebarValueKnown()
	sidebarValue2.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue2.WidgetId = types.StringValue("widget_id")
	sidebarValue2.Settings = types.StringValue(`invalid json`)

	sidebarValue3 := resource_editor_interface.NewSidebarValueKnown()
	sidebarValue3.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue3.WidgetId = types.StringValue("widget_id")
	sidebarValue3.Settings = types.StringValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(resource_editor_interface.SidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
		sidebarValue2,
		sidebarValue3,
	})

	require.Empty(t, sidebarDiags)

	model := resource_editor_interface.EditorInterfaceModel{
		SpaceId:       types.StringValue("space_id"),
		EnvironmentId: types.StringValue("environment_id"),
		ContentTypeId: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	diags := diag.Diagnostics{}
	req := model.ToPutEditorInterfaceReq(ctx, &diags)

	assert.EqualValues(t, contentfulManagement.PutEditorInterfaceReq{
		Controls: []contentfulManagement.PutEditorInterfaceReqControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
			},
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings: contentfulManagement.OptPutEditorInterfaceReqControlsItemSettings{
					Set:   true,
					Value: map[string]jx.Raw{},
				},
			},
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings: contentfulManagement.OptPutEditorInterfaceReqControlsItemSettings{
					Set: true,
					Value: map[string]jx.Raw{
						"foo": jx.Raw(`"bar"`),
					},
				},
			},
		},
		Sidebar: []contentfulManagement.PutEditorInterfaceReqSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
			},
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings: contentfulManagement.OptPutEditorInterfaceReqSidebarItemSettings{
					Set:   true,
					Value: map[string]jx.Raw{},
				},
			},
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings: contentfulManagement.OptPutEditorInterfaceReqSidebarItemSettings{
					Set: true,
					Value: map[string]jx.Raw{
						"foo": jx.Raw(`"bar"`),
					},
				},
			},
		},
	}, req)

	assert.Len(t, diags, 2)

	//nolint:forcetypeassert
	assert.EqualValues(t, "controls[1].settings", diags[0].(diag.DiagnosticWithPath).Path().String())
	assert.EqualValues(t, "Failed to decode settings", diags[0].Summary())

	//nolint:forcetypeassert
	assert.EqualValues(t, "sidebar[1].settings", diags[1].(diag.DiagnosticWithPath).Path().String())
	assert.EqualValues(t, "Failed to decode settings", diags[1].Summary())
}

func TestControlsValueToPutEditorInterfaceReqControlsItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("controls")

	model := resource_editor_interface.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`{"foo":"bar"}`)

	diags := diag.Diagnostics{}
	item := model.ToPutEditorInterfaceReqControlsItem(ctx, path, &diags)

	assert.EqualValues(t, "field_id", item.FieldId)
	assert.EqualValues(t, contentfulManagement.NewOptString("widget_namespace"), item.WidgetNamespace)
	assert.EqualValues(t, contentfulManagement.NewOptString("widget_id"), item.WidgetId)
	assert.True(t, item.Settings.Set)

	assert.Empty(t, diags)
}

func TestControlsValueToPutEditorInterfaceReqControlsItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("controls")

	model := resource_editor_interface.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`invalid json`)

	diags := diag.Diagnostics{}
	model.ToPutEditorInterfaceReqControlsItem(ctx, path, &diags)

	assert.NotEmpty(t, diags)
	assert.Len(t, diags, 1)
}

func TestSidebarValueToPutEditorInterfaceReqSidebarItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("sidebar")

	model := resource_editor_interface.NewSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = types.StringValue(`{"foo":"bar"}`)

	diags := diag.Diagnostics{}
	item := model.ToPutEditorInterfaceReqSidebarItem(ctx, path, &diags)

	assert.EqualValues(t, "widget_namespace", item.WidgetNamespace)
	assert.EqualValues(t, "widget_id", item.WidgetId)
	assert.False(t, item.Disabled.Set)
	assert.True(t, item.Settings.Set)

	assert.Empty(t, diags)
}

func TestSidebarValueToPutEditorInterfaceReqSidebarItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("sidebar")

	model := resource_editor_interface.NewSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = types.StringValue(`invalid json`)

	diags := diag.Diagnostics{}
	model.ToPutEditorInterfaceReqSidebarItem(ctx, path, &diags)

	assert.NotEmpty(t, diags)
	assert.Len(t, diags, 1)
}
