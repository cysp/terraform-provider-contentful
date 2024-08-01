package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripToPutEditorInterfaceReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	editorInterface := contentfulManagement.EditorInterface{
		EditorLayout: contentfulManagement.NewOptNilEditorInterfaceEditorLayoutItemArray([]contentfulManagement.EditorInterfaceEditorLayoutItem{
			{
				GroupId: "group_id",
				Name:    "name",
				Items:   []jx.Raw{jx.Raw(`{"foo":"bar"}`)},
			},
		}),
		Controls: contentfulManagement.NewOptNilEditorInterfaceControlsItemArray([]contentfulManagement.EditorInterfaceControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		GroupControls: contentfulManagement.NewOptNilEditorInterfaceGroupControlsItemArray([]contentfulManagement.EditorInterfaceGroupControlsItem{
			{
				GroupId:         "group_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: contentfulManagement.NewOptNilEditorInterfaceSidebarItemArray([]contentfulManagement.EditorInterfaceSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
	}

	model := provider.EditorInterfaceModel{}
	assert.Empty(t, model.ReadFromResponse(ctx, &editorInterface))

	req, diags := model.ToPutEditorInterfaceReq(ctx)
	assert.Empty(t, diags)

	assert.True(t, req.EditorLayout.Set)
	assert.Len(t, req.EditorLayout.Value, 1)
	assert.Equal(t, contentfulManagement.PutEditorInterfaceReqEditorLayoutItem{
		GroupId: "group_id",
		Name:    "name",
		Items:   []jx.Raw{jx.Raw(`{"foo":"bar"}`)},
	}, req.EditorLayout.Value[0])

	assert.True(t, req.Controls.Set)
	assert.Len(t, req.Controls.Value, 1)
	assert.Equal(t, contentfulManagement.PutEditorInterfaceReqControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
		WidgetId:        contentfulManagement.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Controls.Value[0])

	assert.True(t, req.GroupControls.Set)
	assert.Len(t, req.GroupControls.Value, 1)
	assert.Equal(t, contentfulManagement.PutEditorInterfaceReqGroupControlsItem{
		GroupId:         "group_id",
		WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
		WidgetId:        contentfulManagement.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.GroupControls.Value[0])

	assert.True(t, req.Sidebar.Set)
	assert.Len(t, req.Sidebar.Value, 1)
	assert.Equal(t, contentfulManagement.PutEditorInterfaceReqSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Sidebar.Value[0])
}

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
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: contentfulManagement.NewOptNilPutEditorInterfaceReqSidebarItemArray([]contentfulManagement.PutEditorInterfaceReqSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
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
				Settings:        []byte("invalid json"),
			},
			{
				FieldId:         "field_id",
				WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
				WidgetId:        contentfulManagement.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
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
				Settings:        []byte("invalid json"),
			},
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
	}, req)

	assert.Empty(t, diags)
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
				EditorLayout:  types.ListNull(provider.EditorLayoutValue{}.Type(context.Background())),
				Controls:      types.ListNull(provider.ControlsValue{}.Type(context.Background())),
				GroupControls: types.ListNull(provider.GroupControlsValue{}.Type(context.Background())),
				Sidebar:       types.ListNull(provider.SidebarValue{}.Type(context.Background())),
			},
		},
		"empty": {
			editorInterface: contentfulManagement.EditorInterface{
				EditorLayout:  contentfulManagement.NewOptNilEditorInterfaceEditorLayoutItemArray([]contentfulManagement.EditorInterfaceEditorLayoutItem{}),
				Controls:      contentfulManagement.NewOptNilEditorInterfaceControlsItemArray([]contentfulManagement.EditorInterfaceControlsItem{}),
				GroupControls: contentfulManagement.NewOptNilEditorInterfaceGroupControlsItemArray([]contentfulManagement.EditorInterfaceGroupControlsItem{}),
				Sidebar:       contentfulManagement.NewOptNilEditorInterfaceSidebarItemArray([]contentfulManagement.EditorInterfaceSidebarItem{}),
			},
			expectedModel: provider.EditorInterfaceModel{
				EditorLayout:  util.NewEmptyListMust(provider.EditorLayoutValue{}.Type(context.Background())),
				Controls:      util.NewEmptyListMust(provider.ControlsValue{}.Type(context.Background())),
				GroupControls: util.NewEmptyListMust(provider.GroupControlsValue{}.Type(context.Background())),
				Sidebar:       util.NewEmptyListMust(provider.SidebarValue{}.Type(context.Background())),
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
