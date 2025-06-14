package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorInterfaceModelReadFromResponse(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := map[string]struct {
		editorInterface cm.EditorInterface
		expectedModel   provider.EditorInterfaceModel
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
			expectedModel: provider.EditorInterfaceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
				EditorLayout:  provider.NewTypedListNull[provider.EditorInterfaceEditorLayoutItemValue](ctx),
				Controls:      provider.NewTypedListNull[provider.EditorInterfaceControlValue](ctx),
				GroupControls: provider.NewTypedListNull[provider.EditorInterfaceGroupControlValue](ctx),
				Sidebar:       provider.NewTypedListNull[provider.EditorInterfaceSidebarValue](ctx),
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
			expectedModel: provider.EditorInterfaceModel{
				ID:            types.StringValue("space/environment/content_type"),
				SpaceID:       types.StringValue("space"),
				EnvironmentID: types.StringValue("environment"),
				ContentTypeID: types.StringValue("content_type"),
				EditorLayout:  DiagsNoErrorsMust(provider.NewTypedList(ctx, []provider.EditorInterfaceEditorLayoutItemValue{})),
				Controls:      DiagsNoErrorsMust(provider.NewTypedList(ctx, []provider.EditorInterfaceControlValue{})),
				GroupControls: DiagsNoErrorsMust(provider.NewTypedList(ctx, []provider.EditorInterfaceGroupControlValue{})),
				Sidebar:       DiagsNoErrorsMust(provider.NewTypedList(ctx, []provider.EditorInterfaceSidebarValue{})),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, modelDiags := provider.NewEditorInterfaceResourceModelFromResponse(t.Context(), test.editorInterface)

			assert.Equal(t, test.expectedModel, model)
			assert.Empty(t, modelDiags)
		})
	}
}
