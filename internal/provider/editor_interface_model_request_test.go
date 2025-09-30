package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestRoundTripToEditorInterfaceFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	editorInterface := cm.EditorInterface{
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutGroupItem{
			GroupId: "group_id",
			Name:    "name",
			Items:   []cm.EditorInterfaceEditorLayoutItem{cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "foo"})},
		})}),
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

	model, modelDiags := NewEditorInterfaceResourceModelFromResponse(ctx, editorInterface)
	assert.Empty(t, modelDiags)

	req, diags := model.ToEditorInterfaceFields(ctx)
	assert.Empty(t, diags)

	assert.True(t, req.EditorLayout.Set)
	assert.Len(t, req.EditorLayout.Value, 1)
	assert.Equal(t, cm.EditorInterfaceEditorLayoutItem{
		Type: cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem,
		EditorInterfaceEditorLayoutGroupItem: cm.EditorInterfaceEditorLayoutGroupItem{
			GroupId: "group_id",
			Name:    "name",
			Items: []cm.EditorInterfaceEditorLayoutItem{
				{
					Type:                                 cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem,
					EditorInterfaceEditorLayoutFieldItem: cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "foo"},
				},
			},
		},
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

	controlValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`{"foo":"bar"}`),
	}))

	controls := NewTypedList([]TypedObject[EditorInterfaceControlValue]{
		controlValue1,
	})

	sidebarValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`{"foo":"bar"}`),
		"disabled":         types.BoolNull(),
	}))

	sidebar := NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{
		sidebarValue1,
	})

	model := EditorInterfaceModel{
		EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
			SpaceID:       types.StringValue("space_id"),
			EnvironmentID: types.StringValue("environment_id"),
			ContentTypeID: types.StringValue("content_type_id"),
		},

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

	controlValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedNull(),
	}))

	controlValue2 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`invalid json`),
	}))

	controlValue3 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`{"foo":"bar"}`),
	}))

	controls := NewTypedList([]TypedObject[EditorInterfaceControlValue]{
		controlValue1,
		controlValue2,
		controlValue3,
	})

	sidebarValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedNull(),
		"disabled":         types.BoolNull(),
	}))

	sidebarValue2 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`invalid json`),
		"disabled":         types.BoolNull(),
	}))

	sidebarValue3 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`{"foo":"bar"}`),
		"disabled":         types.BoolNull(),
	}))

	sidebar := NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{
		sidebarValue1,
		sidebarValue2,
		sidebarValue3,
	})

	model := EditorInterfaceModel{
		EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
			SpaceID:       types.StringValue("space_id"),
			EnvironmentID: types.StringValue("environment_id"),
			ContentTypeID: types.StringValue("content_type_id"),
		},

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
