package cmtesting_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
)

// These helpers historically detach the slices that required element conversion.
// Other request storage, including raw settings and parameters, may remain shared.
func TestAppDefinitionFieldsDetachLocationSlices(t *testing.T) {
	t.Parallel()

	fields := cm.AppDefinitionData{
		Name: "App",
		Locations: []cm.AppDefinitionLocationsItem{{
			Location: "entry-field", FieldTypes: []cm.AppDefinitionLocationsItemFieldTypesItem{{Type: "Symbol"}},
		}},
	}
	definition := cmt.NewAppDefinitionFromFields("organization", "app", fields)

	fields.Locations[0].Location = "page"
	fields.Locations[0].FieldTypes[0].Type = "Text"

	assert.Equal(t, []cm.AppDefinitionLocationsItem{{
		Location: "entry-field", FieldTypes: []cm.AppDefinitionLocationsItemFieldTypesItem{{Type: "Symbol"}},
	}}, definition.Locations)
}

func TestExtensionFieldsDetachFieldTypes(t *testing.T) {
	t.Parallel()

	fields := cm.ExtensionData{Extension: cm.ExtensionExtension{
		Name: "Extension", FieldTypes: []cm.ExtensionExtensionFieldTypesItem{{Type: "Symbol"}},
	}}
	extension := cmt.NewExtensionFromFields("space", "environment", "extension", fields)

	fields.Extension.FieldTypes[0].Type = "Text"

	assert.Equal(t, []cm.ExtensionExtensionFieldTypesItem{{Type: "Symbol"}}, extension.Extension.FieldTypes)
}

func TestEditorInterfaceFieldsDetachControlSlices(t *testing.T) {
	t.Parallel()

	fields := cm.EditorInterfaceData{
		Controls:      cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{{FieldId: "title"}}),
		GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{{GroupId: "main"}}),
		Sidebar:       cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{{WidgetNamespace: "builtin", WidgetId: "info-panel"}}),
	}
	editorInterface := cmt.NewEditorInterfaceFromFields("space", "environment", "content-type", fields)

	fields.Controls.Value[0].FieldId = "body"
	fields.GroupControls.Value[0].GroupId = "other"
	fields.Sidebar.Value[0].WidgetId = "other"

	assert.Equal(t, []cm.EditorInterfaceControlsItem{{FieldId: "title"}}, editorInterface.Controls.Value)
	assert.Equal(t, []cm.EditorInterfaceGroupControlsItem{{GroupId: "main"}}, editorInterface.GroupControls.Value)
	assert.Equal(t, []cm.EditorInterfaceSidebarItem{{WidgetNamespace: "builtin", WidgetId: "info-panel"}}, editorInterface.Sidebar.Value)
}
