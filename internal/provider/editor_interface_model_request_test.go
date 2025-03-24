package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripToEditorInterfaceFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	editorInterface := cm.EditorInterface{
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			{
				GroupId: "group_id",
				Name:    "name",
				Items:   []jx.Raw{jx.Raw(`{"foo":"bar"}`)},
			},
		}),
		Controls: cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{
			{
				GroupId:         "group_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
	}

	model := provider.EditorInterfaceResourceModel{}
	assert.Empty(t, model.ReadFromResponse(ctx, &editorInterface))

	req, diags := model.ToEditorInterfaceFields(ctx)
	assert.Empty(t, diags)

	assert.True(t, req.EditorLayout.Set)
	assert.Len(t, req.EditorLayout.Value, 1)
	assert.Equal(t, cm.EditorInterfaceFieldsEditorLayoutItem{
		GroupId: "group_id",
		Name:    "name",
		Items:   []jx.Raw{jx.Raw(`{"foo":"bar"}`)},
	}, req.EditorLayout.Value[0])

	assert.True(t, req.Controls.Set)
	assert.Len(t, req.Controls.Value, 1)
	assert.Equal(t, cm.EditorInterfaceFieldsControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Controls.Value[0])

	assert.True(t, req.GroupControls.Set)
	assert.Len(t, req.GroupControls.Value, 1)
	assert.Equal(t, cm.EditorInterfaceFieldsGroupControlsItem{
		GroupId:         "group_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.GroupControls.Value[0])

	assert.True(t, req.Sidebar.Set)
	assert.Len(t, req.Sidebar.Value, 1)
	assert.Equal(t, cm.EditorInterfaceFieldsSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Sidebar.Value[0])
}

func TestToEditorInterfaceFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	controlValue1 := provider.NewEditorInterfaceControlValueKnown()
	controlValue1.FieldID = types.StringValue("field_id")
	controlValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlValue1.WidgetID = types.StringValue("widget_id")
	controlValue1.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(provider.EditorInterfaceControlValue{}.Type(ctx), []attr.Value{
		controlValue1,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := provider.NewEditorInterfaceSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetID = types.StringValue("widget_id")
	sidebarValue1.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(provider.EditorInterfaceSidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
	})

	require.Empty(t, sidebarDiags)

	model := provider.EditorInterfaceResourceModel{
		SpaceID:       types.StringValue("space_id"),
		EnvironmentID: types.StringValue("environment_id"),
		ContentTypeID: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	req, diags := model.ToEditorInterfaceFields(ctx)

	assert.Empty(t, diags)

	assert.Equal(t, cm.EditorInterfaceFields{
		Controls: cm.NewOptNilEditorInterfaceFieldsControlsItemArray([]cm.EditorInterfaceFieldsControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: cm.NewOptNilEditorInterfaceFieldsSidebarItemArray([]cm.EditorInterfaceFieldsSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
	}, req)
}

func TestToEditorInterfaceFieldsErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	controlValue1 := provider.NewEditorInterfaceControlValueKnown()
	controlValue1.FieldID = types.StringValue("field_id")
	controlValue1.WidgetNamespace = types.StringValue("widget_namespace")
	controlValue1.WidgetID = types.StringValue("widget_id")
	controlValue1.Settings = jsontypes.NewNormalizedNull()

	controlValue2 := provider.NewEditorInterfaceControlValueKnown()
	controlValue2.FieldID = types.StringValue("field_id")
	controlValue2.WidgetNamespace = types.StringValue("widget_namespace")
	controlValue2.WidgetID = types.StringValue("widget_id")
	controlValue2.Settings = jsontypes.NewNormalizedValue(`invalid json`)

	controlValue3 := provider.NewEditorInterfaceControlValueKnown()
	controlValue3.FieldID = types.StringValue("field_id")
	controlValue3.WidgetNamespace = types.StringValue("widget_namespace")
	controlValue3.WidgetID = types.StringValue("widget_id")
	controlValue3.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	controls, controlsDiags := types.ListValue(provider.EditorInterfaceControlValue{}.Type(ctx), []attr.Value{
		controlValue1,
		controlValue2,
		controlValue3,
	})

	require.Empty(t, controlsDiags)

	sidebarValue1 := provider.NewEditorInterfaceSidebarValueKnown()
	sidebarValue1.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue1.WidgetID = types.StringValue("widget_id")
	sidebarValue1.Settings = jsontypes.NewNormalizedNull()

	sidebarValue2 := provider.NewEditorInterfaceSidebarValueKnown()
	sidebarValue2.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue2.WidgetID = types.StringValue("widget_id")
	sidebarValue2.Settings = jsontypes.NewNormalizedValue(`invalid json`)

	sidebarValue3 := provider.NewEditorInterfaceSidebarValueKnown()
	sidebarValue3.WidgetNamespace = types.StringValue("widget_namespace")
	sidebarValue3.WidgetID = types.StringValue("widget_id")
	sidebarValue3.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	sidebar, sidebarDiags := types.ListValue(provider.EditorInterfaceSidebarValue{}.Type(ctx), []attr.Value{
		sidebarValue1,
		sidebarValue2,
		sidebarValue3,
	})

	require.Empty(t, sidebarDiags)

	model := provider.EditorInterfaceResourceModel{
		SpaceID:       types.StringValue("space_id"),
		EnvironmentID: types.StringValue("environment_id"),
		ContentTypeID: types.StringValue("content_type_id"),

		Controls: controls,
		Sidebar:  sidebar,
	}

	req, diags := model.ToEditorInterfaceFields(ctx)

	assert.Equal(t, cm.EditorInterfaceFields{
		Controls: cm.NewOptNilEditorInterfaceFieldsControlsItemArray([]cm.EditorInterfaceFieldsControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
			},
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte("invalid json"),
			},
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: cm.NewOptNilEditorInterfaceFieldsSidebarItemArray([]cm.EditorInterfaceFieldsSidebarItem{
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
