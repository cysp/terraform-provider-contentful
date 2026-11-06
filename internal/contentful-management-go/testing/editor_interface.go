package cmtesting

import (
	"slices"

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

	// Both omitted and null collections clear the stored value. Present arrays,
	// including empty arrays, retain their request representation.
	if editorLayout, ok := editorInterfaceFields.EditorLayout.Get(); ok {
		editorInterface.EditorLayout.SetTo(editorLayout)
	} else {
		editorInterface.EditorLayout.Reset()
	}

	if controls, ok := editorInterfaceFields.Controls.Get(); ok {
		editorInterface.Controls.SetTo(slices.Clone(controls))
	} else {
		editorInterface.Controls.Reset()
	}

	if groupControls, ok := editorInterfaceFields.GroupControls.Get(); ok {
		editorInterface.GroupControls.SetTo(slices.Clone(groupControls))
	} else {
		editorInterface.GroupControls.Reset()
	}

	if sidebar, ok := editorInterfaceFields.Sidebar.Get(); ok {
		editorInterface.Sidebar.SetTo(slices.Clone(sidebar))
	} else {
		editorInterface.Sidebar.Reset()
	}
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

func SyncEditorInterfaceWithContentType(editorInterface *cm.EditorInterface, previousFields *contentTypePublicationSnapshot, fields []cm.ContentTypeFieldsItem) {
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

	for _, field := range fields {
		if previousFields != nil {
			if _, wasPublished := (*previousFields)[field.ID]; wasPublished {
				continue
			}
		}

		if _, alreadyControlled := controlledFieldIDs[field.ID]; alreadyControlled {
			continue
		}

		controls = append(controls, cm.EditorInterfaceControlsItem{FieldId: field.ID})
	}

	editorInterface.Controls.SetTo(controls)
}
