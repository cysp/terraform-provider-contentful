package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripToEditorInterfaceData(t *testing.T) {
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

	req, diags := model.ToEditorInterfaceData(ctx)
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
	assert.Equal(t, cm.EditorInterfaceDataControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Controls.Value[0])

	assert.True(t, req.GroupControls.Set)
	assert.Len(t, req.GroupControls.Value, 1)
	assert.Equal(t, cm.EditorInterfaceDataGroupControlsItem{
		GroupId:         "group_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.GroupControls.Value[0])

	assert.True(t, req.Sidebar.Set)
	assert.Len(t, req.Sidebar.Value, 1)
	assert.Equal(t, cm.EditorInterfaceDataSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
		Settings:        []byte(`{"foo":"bar"}`),
	}, req.Sidebar.Value[0])
}

func TestEditorInterfaceRequestRejectsNullAndUnknownObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]TypedObject[EditorInterfaceControlValue]{
		"null":    NewTypedObjectNull[EditorInterfaceControlValue](),
		"unknown": NewTypedObjectUnknown[EditorInterfaceControlValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := EditorInterfaceModel{
				Controls: NewTypedList([]TypedObject[EditorInterfaceControlValue]{value}),
			}
			request, diags := model.ToEditorInterfaceData(t.Context())
			require.True(t, diags.HasError())
			assert.False(t, request.Controls.Set)
			assert.Equal(t, []string{"controls[0]"}, diagnosticPaths(t, diags))
		})
	}
}

func TestEditorInterfaceLayoutItemRequiresExactlyOneAlternative(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("editor_layout").AtListIndex(0)
	field := NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemFieldValue{
		FieldID: types.StringValue("field"),
	})
	group := NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		GroupID: types.StringValue("group"),
		Name:    types.StringValue("Group"),
		Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{}),
	})

	tests := map[string]struct {
		value         EditorInterfaceEditorLayoutItemGroupItemValue
		expected      cm.EditorInterfaceEditorLayoutItem
		expectedPaths []string
	}{
		"field": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: field,
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			},
			expected: cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(
				cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "field"},
			),
		},
		"group": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: group,
			},
			expected: cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(
				cm.EditorInterfaceEditorLayoutGroupItem{
					GroupId: "group",
					Name:    "Group",
					Items:   []cm.EditorInterfaceEditorLayoutItem{},
				},
			),
		},
		"neither": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			},
			expectedPaths: []string{"editor_layout[0]"},
		},
		"both": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: field,
				Group: group,
			},
			expectedPaths: []string{"editor_layout[0]"},
		},
		"unknown": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			},
			expectedPaths: []string{"editor_layout[0].field"},
		},
		"invalid field": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemFieldValue{
					FieldID: types.StringUnknown(),
				}),
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			},
			expectedPaths: []string{"editor_layout[0].field.field_id"},
		},
		"invalid group ID": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringUnknown(),
					Name:    types.StringValue("Group"),
					Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{}),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.group_id"},
		},
		"invalid group name": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringValue("group"),
					Name:    types.StringNull(),
					Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{}),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.name"},
		},
		"unknown group items": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringValue("group"),
					Name:    types.StringValue("Group"),
					Items:   NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]](),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.items"},
		},
		"null group items": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringValue("group"),
					Name:    types.StringValue("Group"),
					Items:   NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]](),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.items"},
		},
		"invalid nested field": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringValue("group"),
					Name:    types.StringValue("Group"),
					Items: NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{
						NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
							Field: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
								FieldID: types.StringUnknown(),
							}),
						}),
					}),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.items[0].field.field_id"},
		},
		"null nested field": {
			value: EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
					GroupID: types.StringValue("group"),
					Name:    types.StringValue("Group"),
					Items: NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{
						NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
							Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
						}),
					}),
				}),
			},
			expectedPaths: []string{"editor_layout[0].group.items[0].field"},
		},
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := value.value.ToEditorInterfaceEditorLayoutItem(t.Context(), valuePath)

			assert.Equal(t, value.expected, actual)
			assert.ElementsMatch(t, value.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestEditorInterfaceTopLevelGroupRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("editor_layout").AtListIndex(0).AtName("group")

	tests := map[string]struct {
		value        EditorInterfaceEditorLayoutItemGroupValue
		expectedPath string
	}{
		"group ID": {
			value: EditorInterfaceEditorLayoutItemGroupValue{
				GroupID: types.StringUnknown(),
				Name:    types.StringValue("Group"),
				Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{}),
			},
			expectedPath: "editor_layout[0].group.group_id",
		},
		"name": {
			value: EditorInterfaceEditorLayoutItemGroupValue{
				GroupID: types.StringValue("group"),
				Name:    types.StringNull(),
				Items:   NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{}),
			},
			expectedPath: "editor_layout[0].group.name",
		},
		"items": {
			value: EditorInterfaceEditorLayoutItemGroupValue{
				GroupID: types.StringValue("group"),
				Name:    types.StringValue("Group"),
				Items:   NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]](),
			},
			expectedPath: "editor_layout[0].group.items",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.value.ToEditorInterfaceEditorLayoutItem(t.Context(), valuePath)

			assert.Zero(t, actual)
			assert.Equal(t, []string{test.expectedPath}, diagnosticPaths(t, diags))
		})
	}
}

func TestToEditorInterfaceData(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	controlValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`{"foo":"bar"}`)),
	}))

	controls := NewTypedList([]TypedObject[EditorInterfaceControlValue]{
		controlValue1,
	})

	sidebarValue1 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`{"foo":"bar"}`)),
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

	req, diags := model.ToEditorInterfaceData(ctx)

	assert.Empty(t, diags)

	assert.Equal(t, cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{
			{
				FieldId:         "field_id",
				WidgetNamespace: cm.NewOptString("widget_namespace"),
				WidgetId:        cm.NewOptString("widget_id"),
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
		Sidebar: cm.NewOptNilEditorInterfaceDataSidebarItemArray([]cm.EditorInterfaceDataSidebarItem{
			{
				WidgetNamespace: "widget_namespace",
				WidgetId:        "widget_id",
				Settings:        []byte(`{"foo":"bar"}`),
			},
		}),
	}, req)
}

func TestToEditorInterfaceDataErrorHandling(t *testing.T) {
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
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`invalid json`)),
	}))

	controlValue3 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`{"foo":"bar"}`)),
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
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`invalid json`)),
		"disabled":         types.BoolNull(),
	}))

	sidebarValue3 := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`{"foo":"bar"}`)),
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

	req, diags := model.ToEditorInterfaceData(ctx)

	assert.Equal(t, cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{
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
		Sidebar: cm.NewOptNilEditorInterfaceDataSidebarItemArray([]cm.EditorInterfaceDataSidebarItem{
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

func TestToEditorInterfaceDataOptionalListStates(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		setValue     func(*EditorInterfaceModel, bool)
		expectedPath string
		assertSet    func(*testing.T, cm.EditorInterfaceData)
	}{
		"editor layout": {
			setValue: func(model *EditorInterfaceModel, unknown bool) {
				if unknown {
					model.EditorLayout = NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemValue]]()
				} else {
					model.EditorLayout = NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{})
				}
			},
			expectedPath: "editor_layout",
			assertSet: func(t *testing.T, request cm.EditorInterfaceData) {
				t.Helper()
				assert.True(t, request.EditorLayout.Set)
				assert.NotNil(t, request.EditorLayout.Value)
				assert.Empty(t, request.EditorLayout.Value)
			},
		},
		"controls": {
			setValue: func(model *EditorInterfaceModel, unknown bool) {
				if unknown {
					model.Controls = NewTypedListUnknown[TypedObject[EditorInterfaceControlValue]]()
				} else {
					model.Controls = NewTypedList([]TypedObject[EditorInterfaceControlValue]{})
				}
			},
			expectedPath: "controls",
			assertSet: func(t *testing.T, request cm.EditorInterfaceData) {
				t.Helper()
				assert.True(t, request.Controls.Set)
				assert.NotNil(t, request.Controls.Value)
				assert.Empty(t, request.Controls.Value)
			},
		},
		"group controls": {
			setValue: func(model *EditorInterfaceModel, unknown bool) {
				if unknown {
					model.GroupControls = NewTypedListUnknown[TypedObject[EditorInterfaceGroupControlValue]]()
				} else {
					model.GroupControls = NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{})
				}
			},
			expectedPath: "group_controls",
			assertSet: func(t *testing.T, request cm.EditorInterfaceData) {
				t.Helper()
				assert.True(t, request.GroupControls.Set)
				assert.NotNil(t, request.GroupControls.Value)
				assert.Empty(t, request.GroupControls.Value)
			},
		},
		"sidebar": {
			setValue: func(model *EditorInterfaceModel, unknown bool) {
				if unknown {
					model.Sidebar = NewTypedListUnknown[TypedObject[EditorInterfaceSidebarValue]]()
				} else {
					model.Sidebar = NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{})
				}
			},
			expectedPath: "sidebar",
			assertSet: func(t *testing.T, request cm.EditorInterfaceData) {
				t.Helper()
				assert.True(t, request.Sidebar.Set)
				assert.NotNil(t, request.Sidebar.Value)
				assert.Empty(t, request.Sidebar.Value)
			},
		},
	} {
		t.Run(name+" null is omitted", func(t *testing.T) {
			t.Parallel()

			model := newNullEditorInterfaceModel()
			request, diags := model.ToEditorInterfaceData(t.Context())

			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, cm.EditorInterfaceData{}, request)
		})

		t.Run(name+" known empty is preserved", func(t *testing.T) {
			t.Parallel()

			model := newNullEditorInterfaceModel()
			test.setValue(&model, false)

			request, diags := model.ToEditorInterfaceData(t.Context())

			require.False(t, diags.HasError(), diags.Errors())
			test.assertSet(t, request)
		})

		t.Run(name+" unknown is rejected", func(t *testing.T) {
			t.Parallel()

			model := newNullEditorInterfaceModel()
			test.setValue(&model, true)

			request, diags := model.ToEditorInterfaceData(t.Context())

			assert.Equal(t, cm.EditorInterfaceData{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

//nolint:dupl // Parallel table tests intentionally exercise matching request model shapes.
func TestEditorInterfaceControlRequestValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		mutate       func(*EditorInterfaceControlValue)
		expectedPath string
	}{
		"field id":         {mutate: func(value *EditorInterfaceControlValue) { value.FieldID = types.StringUnknown() }, expectedPath: "controls[2].field_id"},
		"widget namespace": {mutate: func(value *EditorInterfaceControlValue) { value.WidgetNamespace = types.StringUnknown() }, expectedPath: "controls[2].widget_namespace"},
		"widget id":        {mutate: func(value *EditorInterfaceControlValue) { value.WidgetID = types.StringUnknown() }, expectedPath: "controls[2].widget_id"},
		"settings":         {mutate: func(value *EditorInterfaceControlValue) { value.Settings = jsontypes.NewNormalizedUnknown() }, expectedPath: "controls[2].settings"},
	} {
		t.Run(name+" unknown", func(t *testing.T) {
			t.Parallel()

			value := validEditorInterfaceControlValue()
			test.mutate(&value)

			item, diags := value.ToEditorInterfaceDataControlsItem(path.Root("controls").AtListIndex(2))

			assert.Equal(t, cm.EditorInterfaceDataControlsItem{}, item)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}

	value := validEditorInterfaceControlValue()
	value.WidgetNamespace = types.StringNull()
	value.WidgetID = types.StringNull()
	value.Settings = jsontypes.NewNormalizedNull()

	item, diags := value.ToEditorInterfaceDataControlsItem(path.Root("controls").AtListIndex(0))

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, cm.EditorInterfaceDataControlsItem{FieldId: "field"}, item)
}

//nolint:dupl // Parallel table tests intentionally exercise matching request model shapes.
func TestEditorInterfaceGroupControlRequestValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		mutate       func(*EditorInterfaceGroupControlValue)
		expectedPath string
	}{
		"group id":         {mutate: func(value *EditorInterfaceGroupControlValue) { value.GroupID = types.StringUnknown() }, expectedPath: "group_controls[1].group_id"},
		"widget namespace": {mutate: func(value *EditorInterfaceGroupControlValue) { value.WidgetNamespace = types.StringUnknown() }, expectedPath: "group_controls[1].widget_namespace"},
		"widget id":        {mutate: func(value *EditorInterfaceGroupControlValue) { value.WidgetID = types.StringUnknown() }, expectedPath: "group_controls[1].widget_id"},
		"settings":         {mutate: func(value *EditorInterfaceGroupControlValue) { value.Settings = jsontypes.NewNormalizedUnknown() }, expectedPath: "group_controls[1].settings"},
	} {
		t.Run(name+" unknown", func(t *testing.T) {
			t.Parallel()

			value := validEditorInterfaceGroupControlValue()
			test.mutate(&value)

			item, diags := value.ToEditorInterfaceDataGroupControlsItem(path.Root("group_controls").AtListIndex(1))

			assert.Equal(t, cm.EditorInterfaceDataGroupControlsItem{}, item)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}

	value := validEditorInterfaceGroupControlValue()
	value.WidgetNamespace = types.StringNull()
	value.WidgetID = types.StringNull()
	value.Settings = jsontypes.NewNormalizedNull()

	item, diags := value.ToEditorInterfaceDataGroupControlsItem(path.Root("group_controls").AtListIndex(0))

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, cm.EditorInterfaceDataGroupControlsItem{GroupId: "group"}, item)
}

//nolint:dupl // Parallel table tests intentionally exercise matching request model shapes.
func TestEditorInterfaceSidebarRequestValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		mutate       func(*EditorInterfaceSidebarValue)
		expectedPath string
	}{
		"widget namespace": {mutate: func(value *EditorInterfaceSidebarValue) { value.WidgetNamespace = types.StringUnknown() }, expectedPath: "sidebar[3].widget_namespace"},
		"widget id":        {mutate: func(value *EditorInterfaceSidebarValue) { value.WidgetID = types.StringUnknown() }, expectedPath: "sidebar[3].widget_id"},
		"settings":         {mutate: func(value *EditorInterfaceSidebarValue) { value.Settings = jsontypes.NewNormalizedUnknown() }, expectedPath: "sidebar[3].settings"},
		"disabled":         {mutate: func(value *EditorInterfaceSidebarValue) { value.Disabled = types.BoolUnknown() }, expectedPath: "sidebar[3].disabled"},
	} {
		t.Run(name+" unknown", func(t *testing.T) {
			t.Parallel()

			value := validEditorInterfaceSidebarValue()
			test.mutate(&value)

			item, diags := value.ToEditorInterfaceDataSidebarItem(path.Root("sidebar").AtListIndex(3))

			assert.Equal(t, cm.EditorInterfaceDataSidebarItem{}, item)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}

	value := validEditorInterfaceSidebarValue()
	value.Settings = jsontypes.NewNormalizedNull()
	value.Disabled = types.BoolNull()

	item, diags := value.ToEditorInterfaceDataSidebarItem(path.Root("sidebar").AtListIndex(0))

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, cm.EditorInterfaceDataSidebarItem{WidgetNamespace: "builtin", WidgetId: "widget"}, item)
}

func TestEditorInterfaceEditorLayoutRejectsUnresolvedNestedValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		item         EditorInterfaceEditorLayoutItemValue
		expectedPath string
	}{
		"required top-level group": {
			item:         EditorInterfaceEditorLayoutItemValue{Group: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupValue]()},
			expectedPath: "editor_layout[0].group",
		},
		"null top-level group": {
			item:         EditorInterfaceEditorLayoutItemValue{Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue]()},
			expectedPath: "editor_layout[0].group",
		},
		"group id": {
			item: editorLayoutGroup(types.StringUnknown(), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutFieldUnion(types.StringValue("field"))),
			})),
			expectedPath: "editor_layout[0].group.group_id",
		},
		"group name": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringNull(), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutFieldUnion(types.StringValue("field"))),
			})),
			expectedPath: "editor_layout[0].group.name",
		},
		"group items": {
			item:         editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]]()),
			expectedPath: "editor_layout[0].group.items",
		},
		"null group items": {
			item:         editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]]()),
			expectedPath: "editor_layout[0].group.items",
		},
		"no union alternative": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
					Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
					Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
				}),
			})),
			expectedPath: "editor_layout[0].group.items[0]",
		},
		"unknown union alternative": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
					Field: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
					Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
				}),
			})),
			expectedPath: "editor_layout[0].group.items[0].field",
		},
		"field id": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutFieldUnion(types.StringUnknown())),
			})),
			expectedPath: "editor_layout[0].group.items[0].field.field_id",
		},
		"nested field object": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutNestedGroupUnion(NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{
					NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
						Field: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
					}),
				}))),
			})),
			expectedPath: "editor_layout[0].group.items[0].group.items[0].field",
		},
		"null nested field object": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutNestedGroupUnion(NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{
					NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
						Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
					}),
				}))),
			})),
			expectedPath: "editor_layout[0].group.items[0].group.items[0].field",
		},
		"nested field id": {
			item: editorLayoutGroup(types.StringValue("group"), types.StringValue("name"), NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{
				NewTypedObject(editorLayoutNestedGroupUnion(NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]{
					NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
						Field: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{FieldID: types.StringUnknown()}),
					}),
				}))),
			})),
			expectedPath: "editor_layout[0].group.items[0].group.items[0].field.field_id",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := newNullEditorInterfaceModel()
			model.EditorLayout = NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{NewTypedObject(test.item)})

			request, diags := model.ToEditorInterfaceData(t.Context())

			assert.Equal(t, cm.EditorInterfaceData{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestToEditorInterfaceDataFailsAtomically(t *testing.T) {
	t.Parallel()

	model := newNullEditorInterfaceModel()
	model.Controls = NewTypedList([]TypedObject[EditorInterfaceControlValue]{NewTypedObject(validEditorInterfaceControlValue())})
	invalidSidebar := validEditorInterfaceSidebarValue()
	invalidSidebar.WidgetID = types.StringUnknown()
	model.Sidebar = NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{NewTypedObject(invalidSidebar)})

	request, diags := model.ToEditorInterfaceData(t.Context())

	assert.Equal(t, cm.EditorInterfaceData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"sidebar[0].widget_id"}, attributeDiagnosticPaths(t, diags))
}

func TestToEditorInterfaceDataRejectsUnknownListElement(t *testing.T) {
	t.Parallel()

	model := newNullEditorInterfaceModel()
	model.Controls = NewTypedList([]TypedObject[EditorInterfaceControlValue]{
		NewTypedObject(validEditorInterfaceControlValue()),
		NewTypedObjectUnknown[EditorInterfaceControlValue](),
	})

	request, diags := model.ToEditorInterfaceData(t.Context())

	assert.Equal(t, cm.EditorInterfaceData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"controls[1]"}, attributeDiagnosticPaths(t, diags))
}

func newNullEditorInterfaceModel() EditorInterfaceModel {
	return EditorInterfaceModel{
		EditorLayout:  NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]](),
		Controls:      NewTypedListNull[TypedObject[EditorInterfaceControlValue]](),
		GroupControls: NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]](),
		Sidebar:       NewTypedListNull[TypedObject[EditorInterfaceSidebarValue]](),
	}
}

func validEditorInterfaceControlValue() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		FieldID:         types.StringValue("field"),
		WidgetNamespace: types.StringValue("builtin"),
		WidgetID:        types.StringValue("widget"),
		Settings:        NewNormalizedJSONTypesNormalizedValue([]byte(`{"key":"value"}`)),
	}
}

func validEditorInterfaceGroupControlValue() EditorInterfaceGroupControlValue {
	return EditorInterfaceGroupControlValue{
		GroupID:         types.StringValue("group"),
		WidgetNamespace: types.StringValue("builtin"),
		WidgetID:        types.StringValue("widget"),
		Settings:        NewNormalizedJSONTypesNormalizedValue([]byte(`{"key":"value"}`)),
	}
}

func validEditorInterfaceSidebarValue() EditorInterfaceSidebarValue {
	return EditorInterfaceSidebarValue{
		WidgetNamespace: types.StringValue("builtin"),
		WidgetID:        types.StringValue("widget"),
		Settings:        NewNormalizedJSONTypesNormalizedValue([]byte(`{"key":"value"}`)),
		Disabled:        types.BoolValue(false),
	}
}

func editorLayoutGroup(
	groupID types.String,
	name types.String,
	items TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]],
) EditorInterfaceEditorLayoutItemValue {
	return EditorInterfaceEditorLayoutItemValue{
		Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupValue{
			GroupID: groupID,
			Name:    name,
			Items:   items,
		}),
	}
}

func editorLayoutFieldUnion(fieldID types.String) EditorInterfaceEditorLayoutItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemValue{
		Field: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemFieldValue{FieldID: fieldID}),
		Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
	}
}

func editorLayoutNestedGroupUnion(
	items TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]],
) EditorInterfaceEditorLayoutItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemValue{
		Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
		Group: NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
			GroupID: types.StringValue("nested-group"),
			Name:    types.StringValue("nested name"),
			Items:   items,
		}),
	}
}
