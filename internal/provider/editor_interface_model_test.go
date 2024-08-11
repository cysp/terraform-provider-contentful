package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPutEditorInterfaceReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	controlsValue1 := provider.NewControlsValueKnown()
	controlsValue1.FieldId = types.StringValue("field_id")
	controlsValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue1.WidgetId = types.StringValue("widget_id")
	controlsValue1.Settings = types.StringValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(provider.ControlsValue{}.Type(ctx), []attr.Value{
		controlsValue1,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := provider.NewSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetId = types.StringValue("widget_id")
	sidebarValue1.Settings = types.StringValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(provider.SidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
	})

	require.Empty(t, sidebarDiags)

	model := provider.EditorInterfaceModel{
		SpaceId:       types.StringValue("space_id"),
		EnvironmentId: types.StringValue("environment_id"),
		ContentTypeId: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	req, diags := model.ToPutEditorInterfaceReq(ctx)

	assert.Empty(t, diags)

	assert.EqualValues(t, contentfulManagement.PutEditorInterfaceReq{
		Controls: contentfulManagement.NewOptNilPutEditorInterfaceReqControlsItemArray([]contentfulManagement.PutEditorInterfaceReqControlsItem{
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
		}),
		Sidebar: contentfulManagement.NewOptNilPutEditorInterfaceReqSidebarItemArray([]contentfulManagement.PutEditorInterfaceReqSidebarItem{
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
		}),
	}, req)
}

func TestToPutEditorInterfaceReqErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	controlsValue1 := provider.NewControlsValueKnown()
	controlsValue1.FieldId = types.StringValue("field_id")
	controlsValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue1.WidgetId = types.StringValue("widget_id")
	controlsValue1.Settings = types.StringNull()

	controlsValue2 := provider.NewControlsValueKnown()
	controlsValue2.FieldId = types.StringValue("field_id")
	controlsValue2.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue2.WidgetId = types.StringValue("widget_id")
	controlsValue2.Settings = types.StringValue(`invalid json`)

	controlsValue3 := provider.NewControlsValueKnown()
	controlsValue3.FieldId = types.StringValue("field_id")
	controlsValue3.WidgetNamespace = types.StringValue("widget_namespace")
	controlsValue3.WidgetId = types.StringValue("widget_id")
	controlsValue3.Settings = types.StringValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(provider.ControlsValue{}.Type(ctx), []attr.Value{
		controlsValue1,
		controlsValue2,
		controlsValue3,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := provider.NewSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetId = types.StringValue("widget_id")
	sidebarValue1.Settings = types.StringNull()

	sidebarValue2 := provider.NewSidebarValueKnown()
	sidebarValue2.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue2.WidgetId = types.StringValue("widget_id")
	sidebarValue2.Settings = types.StringValue(`invalid json`)

	sidebarValue3 := provider.NewSidebarValueKnown()
	sidebarValue3.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue3.WidgetId = types.StringValue("widget_id")
	sidebarValue3.Settings = types.StringValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(provider.SidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
		sidebarValue2,
		sidebarValue3,
	})

	require.Empty(t, sidebarDiags)

	model := provider.EditorInterfaceModel{
		SpaceId:       types.StringValue("space_id"),
		EnvironmentId: types.StringValue("environment_id"),
		ContentTypeId: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	req, diags := model.ToPutEditorInterfaceReq(ctx)

	assert.EqualValues(t, contentfulManagement.PutEditorInterfaceReq{
		Controls: contentfulManagement.NewOptNilPutEditorInterfaceReqControlsItemArray([]contentfulManagement.PutEditorInterfaceReqControlsItem{
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
		}),
		Sidebar: contentfulManagement.NewOptNilPutEditorInterfaceReqSidebarItemArray([]contentfulManagement.PutEditorInterfaceReqSidebarItem{
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
		}),
	}, req)

	assert.Len(t, diags, 2)

	//nolint:forcetypeassert
	assert.EqualValues(t, "controls[1].settings", diags[0].(diag.DiagnosticWithPath).Path().String())
	assert.EqualValues(t, "Failed to decode settings", diags[0].Summary())

	//nolint:forcetypeassert
	assert.EqualValues(t, "sidebar[1].settings", diags[1].(diag.DiagnosticWithPath).Path().String())
	assert.EqualValues(t, "Failed to decode settings", diags[1].Summary())
}

func TestReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface contentfulManagement.EditorInterface
		expectedModel   provider.EditorInterfaceModel
	}{
		"null": {
			editorInterface: contentfulManagement.EditorInterface{},
			expectedModel: provider.EditorInterfaceModel{
				Controls: types.ListNull(provider.ControlsValue{}.Type(context.Background())),
				Sidebar:  types.ListNull(provider.SidebarValue{}.Type(context.Background())),
			},
		},
		"empty": {
			editorInterface: contentfulManagement.EditorInterface{
				Controls: contentfulManagement.NewOptNilEditorInterfaceControlsItemArray([]contentfulManagement.EditorInterfaceControlsItem{}),
				Sidebar:  contentfulManagement.NewOptNilEditorInterfaceSidebarItemArray([]contentfulManagement.EditorInterfaceSidebarItem{}),
			},
			expectedModel: provider.EditorInterfaceModel{
				Controls: util.NewEmptyListMust(provider.ControlsValue{}.Type(context.Background())),
				Sidebar:  util.NewEmptyListMust(provider.SidebarValue{}.Type(context.Background())),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.EditorInterfaceModel{}

			diags := model.ReadFromResponse(context.Background(), &test.editorInterface)

			assert.EqualValues(t, test.expectedModel, model)
			assert.Empty(t, diags)
		})
	}
}
