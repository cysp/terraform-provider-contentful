package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorInterfaceModelReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface cm.EditorInterface
		expectedModel   provider.EditorInterfaceResourceModel
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
			expectedModel: provider.EditorInterfaceResourceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
				EditorLayout:  types.ListNull(provider.EditorInterfaceEditorLayoutValue{}.Type(t.Context())),
				Controls:      types.ListNull(provider.EditorInterfaceControlValue{}.Type(t.Context())),
				GroupControls: types.ListNull(provider.EditorInterfaceGroupControlValue{}.Type(t.Context())),
				Sidebar:       types.ListNull(provider.EditorInterfaceSidebarValue{}.Type(t.Context())),
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
			expectedModel: provider.EditorInterfaceResourceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
				EditorLayout:  types.ListValueMust(provider.EditorInterfaceEditorLayoutValue{}.Type(t.Context()), []attr.Value{}),
				Controls:      types.ListValueMust(provider.EditorInterfaceControlValue{}.Type(t.Context()), []attr.Value{}),
				GroupControls: types.ListValueMust(provider.EditorInterfaceGroupControlValue{}.Type(t.Context()), []attr.Value{}),
				Sidebar:       types.ListValueMust(provider.EditorInterfaceSidebarValue{}.Type(t.Context()), []attr.Value{}),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.EditorInterfaceResourceModel{}

			diags := model.ReadFromResponse(t.Context(), &test.editorInterface)

			assert.EqualValues(t, test.expectedModel, model)
			assert.Empty(t, diags)
		})
	}
}
