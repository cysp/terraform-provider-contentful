package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorInterfaceModelReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface cm.EditorInterface
		expectedModel   EditorInterfaceModel
	}{
		"null": {
			editorInterface: cm.EditorInterface{
				Sys: cm.EditorInterfaceSys{
					Space: cm.SpaceLink{
						Sys: cm.SpaceLinkSys{
							ID: "space",
						},
					},
					Environment: cm.EnvironmentLink{
						Sys: cm.EnvironmentLinkSys{
							ID: "environment",
						},
					},
					ContentType: cm.ContentTypeLink{
						Sys: cm.ContentTypeLinkSys{
							ID: "content_type",
						},
					},
					ID: "null",
				},
			},
			expectedModel: EditorInterfaceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
				EditorLayout:  NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]](),
				Controls:      NewTypedListNull[TypedObject[EditorInterfaceControlValue]](),
				GroupControls: NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]](),
				Sidebar:       NewTypedListNull[TypedObject[EditorInterfaceSidebarValue]](),
			},
		},
		"empty": {
			editorInterface: cm.EditorInterface{
				Sys: cm.EditorInterfaceSys{
					Space: cm.SpaceLink{
						Sys: cm.SpaceLinkSys{
							ID: "space",
						},
					},
					Environment: cm.EnvironmentLink{
						Sys: cm.EnvironmentLinkSys{
							ID: "environment",
						},
					},
					ContentType: cm.ContentTypeLink{
						Sys: cm.ContentTypeLinkSys{
							ID: "content_type",
						},
					},
					ID: "empty",
				},
				EditorLayout:  cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{}),
				Controls:      cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{}),
				GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{}),
				Sidebar:       cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{}),
			},
			expectedModel: EditorInterfaceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
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
