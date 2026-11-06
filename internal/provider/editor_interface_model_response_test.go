package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditorInterfaceResponsePreservesListPresence(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface cm.EditorInterface
		expectedModel   EditorInterfaceModel
	}{
		"null": {
			editorInterface: cm.EditorInterface{
				Sys: cm.NewEditorInterfaceSys("space", "environment", "content_type", "null"),
			},
			expectedModel: EditorInterfaceModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "environment", "content_type"),
				EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					ContentTypeID: types.StringValue("content_type"),
				},
				EditorLayout:  NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]](),
				Controls:      NewTypedListNull[TypedObject[EditorInterfaceControlValue]](),
				GroupControls: NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]](),
				Sidebar:       NewTypedListNull[TypedObject[EditorInterfaceSidebarValue]](),
			},
		},
		"empty": {
			editorInterface: cm.EditorInterface{
				Sys:           cm.NewEditorInterfaceSys("space", "environment", "content_type", "empty"),
				EditorLayout:  cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{}),
				Controls:      cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{}),
				GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{}),
				Sidebar:       cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{}),
			},
			expectedModel: EditorInterfaceModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "environment", "content_type"),
				EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					ContentTypeID: types.StringValue("content_type"),
				},
				EditorLayout:  NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{}),
				Controls:      NewTypedList([]TypedObject[EditorInterfaceControlValue]{}),
				GroupControls: NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{}),
				Sidebar:       NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{}),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, modelDiags := NewEditorInterfaceResourceModelFromResponse(t.Context(), test.editorInterface)

			assert.Equal(t, test.expectedModel, model)
			assert.Empty(t, modelDiags)
		})
	}
}

func TestEditorInterfaceResponseProjectsWidgetValues(t *testing.T) {
	t.Parallel()

	response := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "environment", "content-type", "default"),
		Controls: cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{{
			FieldId: "title", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("singleLine"),
			Settings: []byte(testJSON(map[string]any{"helpText": "Title"})),
		}}),
		GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{{
			GroupId: "details", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("fieldset"),
			Settings: []byte(testJSON(map[string]any{"collapsed": true})),
		}}),
		Sidebar: cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{{
			WidgetNamespace: "app", WidgetId: "custom", Disabled: cm.NewOptBool(true),
			Settings: []byte(testJSON(map[string]any{"label": "Details"})),
		}}),
	}

	actual, diags := NewEditorInterfaceResourceModelFromResponse(t.Context(), response)

	require.Empty(t, diags)
	assert.Equal(t, EditorInterfaceModel{
		ID:            types.StringValue("space/environment/content-type"),
		SpaceID:       types.StringValue("space"),
		EnvironmentID: types.StringValue("environment"),
		ContentTypeID: types.StringValue("content-type"),
		Controls: NewTypedList([]TypedObject[EditorInterfaceControlValue]{NewTypedObject(EditorInterfaceControlValue{
			FieldID: types.StringValue("title"), WidgetNamespace: types.StringValue("builtin"), WidgetID: types.StringValue("singleLine"),
			Settings: jsontypes.NewNormalizedValue(testJSON(map[string]any{"helpText": "Title"})),
		})}),
		GroupControls: NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{NewTypedObject(EditorInterfaceGroupControlValue{
			GroupID: types.StringValue("details"), WidgetNamespace: types.StringValue("builtin"), WidgetID: types.StringValue("fieldset"),
			Settings: jsontypes.NewNormalizedValue(testJSON(map[string]any{"collapsed": true})),
		})}),
		Sidebar: NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{NewTypedObject(EditorInterfaceSidebarValue{
			WidgetNamespace: types.StringValue("app"), WidgetID: types.StringValue("custom"), Disabled: types.BoolValue(true),
			Settings: jsontypes.NewNormalizedValue(testJSON(map[string]any{"label": "Details"})),
		})}),
	}, actual)
}
