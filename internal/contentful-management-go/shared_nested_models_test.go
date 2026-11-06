package contentfulmanagement_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppDefinitionSharedLocationsWireValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		locations []cm.AppDefinitionLocationsItem
		wire      map[string]any
	}{
		"omitted": {wire: map[string]any{"name": "App"}},
		"empty":   {locations: []cm.AppDefinitionLocationsItem{}, wire: map[string]any{"name": "App", "locations": []any{}}},
		"nested": {
			locations: []cm.AppDefinitionLocationsItem{
				{Location: "entry-field", FieldTypes: []cm.AppDefinitionLocationsItemFieldTypesItem{
					{Type: "Symbol"},
					{Type: "Link", LinkType: cm.NewOptString("")},
					{Type: "Array", Items: cm.NewOptAppDefinitionLocationsItemFieldTypesItemItems(cm.AppDefinitionLocationsItemFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Entry")})},
				}},
				{Location: "page", NavigationItem: cm.NewOptAppDefinitionLocationsItemNavigationItem(cm.AppDefinitionLocationsItemNavigationItem{Name: "", Path: "/app"})},
				{Location: "entry-sidebar", FieldTypes: []cm.AppDefinitionLocationsItemFieldTypesItem{}},
			},
			wire: map[string]any{"name": "App", "locations": []any{
				map[string]any{"location": "entry-field", "fieldTypes": []any{
					map[string]any{"type": "Symbol"},
					map[string]any{"type": "Link", "linkType": ""},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Entry"}},
				}},
				map[string]any{"location": "page", "navigationItem": map[string]any{"name": "", "path": "/app"}},
				map[string]any{"location": "entry-sidebar", "fieldTypes": []any{}},
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := cm.AppDefinitionData{Name: "App", Locations: test.locations}
			encoded, err := request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, testJSON(test.wire), string(encoded))

			test.wire["sys"] = map[string]any{
				"type": "AppDefinition", "id": "app",
				"organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}},
			}

			var response cm.AppDefinition
			require.NoError(t, json.Unmarshal([]byte(testJSON(test.wire)), &response))
			assert.Equal(t, test.locations, response.Locations)
		})
	}
}

func TestExtensionSharedConfigurationWireValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		extension cm.ExtensionExtension
		wire      map[string]any
	}{
		"empty field types": {
			extension: cm.ExtensionExtension{Name: "Extension", FieldTypes: []cm.ExtensionExtensionFieldTypesItem{}},
			wire:      map[string]any{"name": "Extension", "fieldTypes": []any{}},
		},
		"nested and zero values": {
			extension: cm.ExtensionExtension{
				Name: "Extension", Srcdoc: cm.NewOptString(""), Sidebar: cm.NewOptBool(false),
				FieldTypes: []cm.ExtensionExtensionFieldTypesItem{
					{Type: "Symbol"},
					{Type: "Link", LinkType: cm.NewOptString("")},
					{Type: "Array", Items: cm.NewOptExtensionExtensionFieldTypesItemItems(cm.ExtensionExtensionFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Asset")})},
				},
			},
			wire: map[string]any{
				"name": "Extension", "srcdoc": "", "sidebar": false,
				"fieldTypes": []any{
					map[string]any{"type": "Symbol"},
					map[string]any{"type": "Link", "linkType": ""},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Asset"}},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := cm.ExtensionData{Extension: test.extension}
			encoded, err := request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, testJSON(map[string]any{"extension": test.wire}), string(encoded))

			var response cm.Extension
			require.NoError(t, json.Unmarshal([]byte(testJSON(map[string]any{
				"sys": map[string]any{
					"type": "Extension", "id": "extension", "version": 3,
					"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
					"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "environment"}},
				},
				"extension": test.wire,
			})), &response))
			assert.Equal(t, test.extension, response.Extension)
		})
	}
}

func TestEditorInterfaceSharedControlsWireValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		request cm.EditorInterfaceData
		wire    map[string]any
	}{
		"omitted": {wire: map[string]any{}},
		"null": {
			request: cm.EditorInterfaceData{
				Controls:      cm.OptNilEditorInterfaceControlsItemArray{Set: true, Null: true},
				GroupControls: cm.OptNilEditorInterfaceGroupControlsItemArray{Set: true, Null: true},
				Sidebar:       cm.OptNilEditorInterfaceSidebarItemArray{Set: true, Null: true},
			},
			wire: map[string]any{"controls": nil, "groupControls": nil, "sidebar": nil},
		},
		"empty": {
			request: cm.EditorInterfaceData{
				Controls:      cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{}),
				GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{}),
				Sidebar:       cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{}),
			},
			wire: map[string]any{"controls": []any{}, "groupControls": []any{}, "sidebar": []any{}},
		},
		"nested and zero values": {
			request: cm.EditorInterfaceData{
				Controls: cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{
					{FieldId: "title"},
					{FieldId: "body", WidgetNamespace: cm.NewOptString(""), WidgetId: cm.NewOptString("text"), Settings: []byte(`null`)},
				}),
				GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{
					{GroupId: "main", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("topLevelTab"), Settings: []byte(testJSON(map[string]any{}))},
				}),
				Sidebar: cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{
					{WidgetNamespace: "builtin", WidgetId: "info-panel", Disabled: cm.NewOptBool(false), Settings: []byte(testJSON(map[string]any{"nested": []any{nil, "", false}}))},
					{WidgetNamespace: "extension", WidgetId: "custom"},
				}),
			},
			wire: map[string]any{
				"controls": []any{
					map[string]any{"fieldId": "title"},
					map[string]any{"fieldId": "body", "widgetNamespace": "", "widgetId": "text", "settings": nil},
				},
				"groupControls": []any{map[string]any{"groupId": "main", "widgetNamespace": "builtin", "widgetId": "topLevelTab", "settings": map[string]any{}}},
				"sidebar": []any{
					map[string]any{"widgetNamespace": "builtin", "widgetId": "info-panel", "disabled": false, "settings": map[string]any{"nested": []any{nil, "", false}}},
					map[string]any{"widgetNamespace": "extension", "widgetId": "custom"},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoded, err := test.request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, testJSON(test.wire), string(encoded))

			test.wire["sys"] = map[string]any{
				"type": "EditorInterface", "id": "default", "version": 3,
				"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "environment"}},
				"contentType": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "ContentType", "id": "content-type"}},
			}

			var response cm.EditorInterface
			require.NoError(t, json.Unmarshal([]byte(testJSON(test.wire)), &response))
			assert.Equal(t, test.request.Controls, response.Controls)
			assert.Equal(t, test.request.GroupControls, response.GroupControls)
			assert.Equal(t, test.request.Sidebar, response.Sidebar)
		})
	}
}
