package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID string, editorInterfaceFields cm.EditorInterfaceData) cm.EditorInterface {
	editorInterface := cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys(spaceID, environmentID, contentTypeID, "default"),
	}

	UpdateEditorInterfaceFromFields(&editorInterface, editorInterfaceFields)

	return editorInterface
}

func UpdateEditorInterfaceFromFields(editorInterface *cm.EditorInterface, editorInterfaceFields cm.EditorInterfaceData) {
	editorInterface.Sys.Version++

	convertOptNil(&editorInterface.EditorLayout, &editorInterfaceFields.EditorLayout, func(editorLayout []cm.EditorInterfaceEditorLayoutItem) []cm.EditorInterfaceEditorLayoutItem {
		return editorLayout
	})

	convertOptNil(&editorInterface.Controls, &editorInterfaceFields.Controls, func(controls []cm.EditorInterfaceDataControlsItem) []cm.EditorInterfaceControlsItem {
		return convertSlice(controls, func(control cm.EditorInterfaceDataControlsItem) cm.EditorInterfaceControlsItem {
			return cm.EditorInterfaceControlsItem(control)
		})
	})

	convertOptNil(&editorInterface.GroupControls, &editorInterfaceFields.GroupControls, func(groupControl []cm.EditorInterfaceDataGroupControlsItem) []cm.EditorInterfaceGroupControlsItem {
		return convertSlice(groupControl, func(groupControlItem cm.EditorInterfaceDataGroupControlsItem) cm.EditorInterfaceGroupControlsItem {
			return cm.EditorInterfaceGroupControlsItem(groupControlItem)
		})
	})

	convertOptNil(&editorInterface.Sidebar, &editorInterfaceFields.Sidebar, func(sidebar []cm.EditorInterfaceDataSidebarItem) []cm.EditorInterfaceSidebarItem {
		return convertSlice(sidebar, func(sidebarItem cm.EditorInterfaceDataSidebarItem) cm.EditorInterfaceSidebarItem {
			return cm.EditorInterfaceSidebarItem(sidebarItem)
		})
	})
}

func NewDefaultEditorInterface(spaceID, environmentID, contentTypeID string, fields []cm.ContentTypeFieldsItem) cm.EditorInterface {
	controls := make([]cm.EditorInterfaceControlsItem, len(fields))
	for index, field := range fields {
		controls[index] = cm.EditorInterfaceControlsItem{FieldId: field.ID}
	}

	return cm.EditorInterface{
		Sys: cm.EditorInterfaceSys{
			Space:       cm.NewSpaceLink(spaceID),
			Environment: cm.NewEnvironmentLink(environmentID),
			Type:        cm.EditorInterfaceSysTypeEditorInterface,
			ID:          "default",
			ContentType: cm.NewContentTypeLink(contentTypeID),
			Version:     1,
		},
		Controls: cm.NewOptNilEditorInterfaceControlsItemArray(controls),
	}
}

func SyncEditorInterfaceWithContentType(editorInterface *cm.EditorInterface, previousFieldIDs []string, fields []cm.ContentTypeFieldsItem) {
	editorInterface.Sys.Version++

	existingControls, controlsSet := editorInterface.Controls.Get()
	if !controlsSet {
		return
	}

	currentFieldIDs := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		currentFieldIDs[field.ID] = struct{}{}
	}

	controls := make([]cm.EditorInterfaceControlsItem, 0, len(existingControls)+len(fields))

	controlledFieldIDs := make(map[string]struct{}, len(existingControls))
	for _, control := range existingControls {
		if _, fieldExists := currentFieldIDs[control.FieldId]; !fieldExists {
			continue
		}

		controls = append(controls, control)
		controlledFieldIDs[control.FieldId] = struct{}{}
	}

	previousFieldIDSet := make(map[string]struct{}, len(previousFieldIDs))
	for _, fieldID := range previousFieldIDs {
		previousFieldIDSet[fieldID] = struct{}{}
	}

	for _, field := range fields {
		if _, wasPublished := previousFieldIDSet[field.ID]; wasPublished {
			continue
		}

		if _, alreadyControlled := controlledFieldIDs[field.ID]; alreadyControlled {
			continue
		}

		controls = append(controls, cm.EditorInterfaceControlsItem{FieldId: field.ID})
	}

	editorInterface.Controls.SetTo(controls)
}

func contentTypeFieldIDs(fields []cm.ContentTypeFieldsItem) []string {
	fieldIDs := make([]string, len(fields))
	for index, field := range fields {
		fieldIDs[index] = field.ID
	}

	return fieldIDs
}
