package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface cm.EditorInterface
		expectedModel   provider.EditorInterfaceModel
	}{
		"null": {
			editorInterface: cm.EditorInterface{},
			expectedModel: provider.EditorInterfaceModel{
				EditorLayout:  types.ListNull(provider.EditorInterfaceEditorLayoutValue{}.Type(t.Context())),
				Controls:      types.ListNull(provider.EditorInterfaceControlValue{}.Type(t.Context())),
				GroupControls: types.ListNull(provider.EditorInterfaceGroupControlValue{}.Type(t.Context())),
				Sidebar:       types.ListNull(provider.EditorInterfaceSidebarValue{}.Type(t.Context())),
			},
		},
		"empty": {
			editorInterface: cm.EditorInterface{
				EditorLayout:  cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{}),
				Controls:      cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{}),
				GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{}),
				Sidebar:       cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{}),
			},
			expectedModel: provider.EditorInterfaceModel{
				EditorLayout:  provider.NewEmptyListMust(provider.EditorInterfaceEditorLayoutValue{}.Type(t.Context())),
				Controls:      provider.NewEmptyListMust(provider.EditorInterfaceControlValue{}.Type(t.Context())),
				GroupControls: provider.NewEmptyListMust(provider.EditorInterfaceGroupControlValue{}.Type(t.Context())),
				Sidebar:       provider.NewEmptyListMust(provider.EditorInterfaceSidebarValue{}.Type(t.Context())),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.EditorInterfaceModel{}

			diags := model.ReadFromResponse(t.Context(), &test.editorInterface)

			assert.EqualValues(t, test.expectedModel, model)
			assert.Empty(t, diags)
		})
	}
}
