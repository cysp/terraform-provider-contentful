package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_editor_interface"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestReadEditorInterfaceModel(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		editorInterface contentfulManagement.EditorInterface
		expectedModel   resource_editor_interface.EditorInterfaceModel
	}{
		"null": {
			editorInterface: contentfulManagement.EditorInterface{},
			expectedModel: resource_editor_interface.EditorInterfaceModel{
				Controls: types.ListNull(resource_editor_interface.ControlsValue{}.Type(context.Background())),
				Sidebar:  types.ListNull(resource_editor_interface.SidebarValue{}.Type(context.Background())),
			},
		},
		"empty": {
			editorInterface: contentfulManagement.EditorInterface{
				Controls: contentfulManagement.NewOptNilEditorInterfaceControlsItemArray([]contentfulManagement.EditorInterfaceControlsItem{}),
				Sidebar:  contentfulManagement.NewOptNilEditorInterfaceSidebarItemArray([]contentfulManagement.EditorInterfaceSidebarItem{}),
			},
			expectedModel: resource_editor_interface.EditorInterfaceModel{
				Controls: util.NewEmptyListMust(resource_editor_interface.ControlsValue{}.Type(context.Background())),
				Sidebar:  util.NewEmptyListMust(resource_editor_interface.SidebarValue{}.Type(context.Background())),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := resource_editor_interface.EditorInterfaceModel{}

			diags := provider.ReadEditorInterfaceModel(context.Background(), &model, test.editorInterface)

			assert.EqualValues(t, test.expectedModel, model)
			assert.Empty(t, diags)
		})
	}
}
