package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID string, editorInterfaceFields cm.EditorInterfaceFields) cm.EditorInterface {
	editorInterface := cm.EditorInterface{
		Sys: NewEditorInterfaceSys(spaceID, environmentID, contentTypeID),
	}

	UpdateEditorInterfaceFromFields(&editorInterface, editorInterfaceFields)

	return editorInterface
}

func NewEditorInterfaceSys(spaceID, environmentID, contentTypeID string) cm.EditorInterfaceSys {
	return cm.EditorInterfaceSys{
		Type: cm.EditorInterfaceSysTypeEditorInterface,
		ID:   "default",
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		Environment: cm.EnvironmentLink{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       environmentID,
			},
		},
		ContentType: cm.ContentTypeLink{
			Sys: cm.ContentTypeLinkSys{
				Type:     cm.ContentTypeLinkSysTypeLink,
				LinkType: cm.ContentTypeLinkSysLinkTypeContentType,
				ID:       contentTypeID,
			},
		},
	}
}

func UpdateEditorInterfaceFromFields(editorInterface *cm.EditorInterface, editorInterfaceFields cm.EditorInterfaceFields) {
	convertOptNil(&editorInterface.EditorLayout, &editorInterfaceFields.EditorLayout, func(editorLayout []cm.EditorInterfaceEditorLayoutItem) []cm.EditorInterfaceEditorLayoutItem {
		return editorLayout
	})

	convertOptNil(&editorInterface.Controls, &editorInterfaceFields.Controls, func(controls []cm.EditorInterfaceFieldsControlsItem) []cm.EditorInterfaceControlsItem {
		return convertSlice(controls, func(control cm.EditorInterfaceFieldsControlsItem) cm.EditorInterfaceControlsItem {
			return cm.EditorInterfaceControlsItem(control)
		})
	})

	convertOptNil(&editorInterface.GroupControls, &editorInterfaceFields.GroupControls, func(groupControl []cm.EditorInterfaceFieldsGroupControlsItem) []cm.EditorInterfaceGroupControlsItem {
		return convertSlice(groupControl, func(groupControlItem cm.EditorInterfaceFieldsGroupControlsItem) cm.EditorInterfaceGroupControlsItem {
			return cm.EditorInterfaceGroupControlsItem(groupControlItem)
		})
	})

	convertOptNil(&editorInterface.Sidebar, &editorInterfaceFields.Sidebar, func(sidebar []cm.EditorInterfaceFieldsSidebarItem) []cm.EditorInterfaceSidebarItem {
		return convertSlice(sidebar, func(sidebarItem cm.EditorInterfaceFieldsSidebarItem) cm.EditorInterfaceSidebarItem {
			return cm.EditorInterfaceSidebarItem(sidebarItem)
		})
	})
}
