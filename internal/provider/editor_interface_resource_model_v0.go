package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EditorInterfaceResourceModelV0 struct {
	ID            string                                      `json:"id" tfsdk:"id"`
	SpaceID       string                                      `json:"space_id" tfsdk:"space_id"`
	EnvironmentID string                                      `json:"environment_id" tfsdk:"environment_id"`
	ContentTypeID string                                      `json:"content_type_id" tfsdk:"content_type_id"`
	EditorLayout  []EditorInterfaceEditorLayoutElementValueV0 `json:"editor_layout" tfsdk:"editor_layout"`
	Controls      []EditorInterfaceControlValueV0             `json:"controls" tfsdk:"controls"`
	GroupControls []EditorInterfaceGroupControlValueV0        `json:"group_controls" tfsdk:"group_controls"`
	Sidebar       []EditorInterfaceSidebarValueV0             `json:"sidebar" tfsdk:"sidebar"`
}

type EditorInterfaceEditorLayoutElementValueV0 struct {
	GroupID string   `json:"group_id" tfsdk:"group_id"`
	Name    string   `json:"name" tfsdk:"name"`
	Items   []string `json:"items" tfsdk:"items"`
}

type EditorInterfaceEditorLayoutElementValueItemV0 struct {
	FieldID *string                                         `json:"field_id" tfsdk:"field_id"`
	GroupID *string                                         `json:"group_id" tfsdk:"group_id"`
	Name    *string                                         `json:"name" tfsdk:"name"`
	Items   []EditorInterfaceEditorLayoutElementValueItemV0 `json:"items" tfsdk:"items"`
}

type EditorInterfaceControlValueV0 struct {
	FieldID         string  `json:"field_id" tfsdk:"field_id"`
	WidgetNamespace *string `json:"widget_namespace" tfsdk:"widget_namespace"`
	WidgetID        *string `json:"widget_id" tfsdk:"widget_id"`
	Settings        *string `json:"settings" tfsdk:"settings"`
}

type EditorInterfaceGroupControlValueV0 struct {
	GroupID         string  `json:"group_id" tfsdk:"group_id"`
	WidgetNamespace *string `json:"widget_namespace" tfsdk:"widget_namespace"`
	WidgetID        *string `json:"widget_id" tfsdk:"widget_id"`
	Settings        *string `json:"settings" tfsdk:"settings"`
}

type EditorInterfaceSidebarValueV0 struct {
	WidgetNamespace *string `json:"widget_namespace" tfsdk:"widget_namespace"`
	WidgetID        *string `json:"widget_id" tfsdk:"widget_id"`
	Settings        *string `json:"settings" tfsdk:"settings"`
	Disabled        *bool   `json:"disabled" tfsdk:"disabled"`
}

func upgradeEditorInterfaceResourceModelV0ToV1(ctx context.Context, modelV0 EditorInterfaceResourceModelV0) (EditorInterfaceResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := EditorInterfaceResourceModel{
		ID:            types.StringValue(modelV0.ID),
		SpaceID:       types.StringValue(modelV0.SpaceID),
		EnvironmentID: types.StringValue(modelV0.EnvironmentID),
		ContentTypeID: types.StringValue(modelV0.ContentTypeID),
	}

	editorLayoutElements := make([]EditorInterfaceEditorLayoutElementValue, 0, len(modelV0.EditorLayout))
	for _, element := range modelV0.EditorLayout {
		editorLayoutElement := EditorInterfaceEditorLayoutElementValue{
			GroupID: types.StringValue(element.GroupID),
			Name:    types.StringValue(element.Name),
		}

		items := make([]EditorInterfaceEditorLayoutElementItemValue, 0, len(element.Items))

		for _, item := range element.Items {
			var itemObject EditorInterfaceEditorLayoutElementValueItemV0
			itemObjectUnmarshalErr := json.Unmarshal([]byte(item), &itemObject)
			if itemObjectUnmarshalErr != nil {
				diags.AddError(
					"Failed to unmarshal item",
					"Failed to unmarshal item: "+itemObjectUnmarshalErr.Error(),
				)
				continue
			}

			itemElement := EditorInterfaceEditorLayoutElementItemValue{}

			if itemObject.FieldID != nil {
				itemElement.Field = EditorInterfaceEditorLayoutElementItemFieldValue{
					FieldID: types.StringPointerValue(itemObject.FieldID),
				}
			}
			if itemObject.GroupID != nil {
				itemElement.Group = EditorInterfaceEditorLayoutElementItemGroupValue{
					GroupID: types.StringPointerValue(itemObject.GroupID),
					Name:    types.StringPointerValue(itemObject.Name),
				}
			}

			items = append(items, itemElement)
		}

		editorLayoutElementItems, editorLayoutElementItemsDiags := types.ListValueFrom(ctx, EditorInterfaceEditorLayoutElementItemValue{}.CustomType(ctx), items)
		diags.Append(editorLayoutElementItemsDiags...)

		editorLayoutElement.Items = editorLayoutElementItems

		editorLayoutElements = append(editorLayoutElements, editorLayoutElement)
	}

	editorLayoutList, editorLayoutListDiags := types.ListValueFrom(ctx, EditorInterfaceEditorLayoutElementValue{}.CustomType(ctx), editorLayoutElements)
	diags.Append(editorLayoutListDiags...)

	model.EditorLayout = editorLayoutList

	controlsElements := make([]EditorInterfaceControlValue, len(modelV0.Controls))
	for i, element := range modelV0.Controls {
		controlsElements[i] = EditorInterfaceControlValue{
			FieldID:         types.StringValue(element.FieldID),
			WidgetNamespace: types.StringPointerValue(element.WidgetNamespace),
			WidgetID:        types.StringPointerValue(element.WidgetID),
			Settings:        jsontypes.NewNormalizedPointerValue(element.Settings),
		}
	}

	controlsList, controlsListDiags := types.ListValueFrom(ctx, EditorInterfaceControlValue{}.CustomType(ctx), controlsElements)
	diags.Append(controlsListDiags...)
	model.Controls = controlsList

	groupControlsElements := make([]EditorInterfaceGroupControlValue, len(modelV0.GroupControls))
	for i, element := range modelV0.GroupControls {
		groupControlsElements[i] = EditorInterfaceGroupControlValue{
			GroupID:         types.StringValue(element.GroupID),
			WidgetNamespace: types.StringPointerValue(element.WidgetNamespace),
			WidgetID:        types.StringPointerValue(element.WidgetID),
			Settings:        jsontypes.NewNormalizedPointerValue(element.Settings),
		}
	}

	groupControlsList, groupControlsListDiags := types.ListValueFrom(ctx, EditorInterfaceGroupControlValue{}.CustomType(ctx), groupControlsElements)
	diags.Append(groupControlsListDiags...)

	model.GroupControls = groupControlsList

	sidebarElements := make([]EditorInterfaceSidebarValue, len(modelV0.Sidebar))
	for i, element := range modelV0.Sidebar {
		sidebarElements[i] = EditorInterfaceSidebarValue{
			WidgetNamespace: types.StringPointerValue(element.WidgetNamespace),
			WidgetID:        types.StringPointerValue(element.WidgetID),
			Settings:        jsontypes.NewNormalizedPointerValue(element.Settings),
			Disabled:        types.BoolPointerValue(element.Disabled),
		}
	}

	sidebarList, sidebarListDiags := types.ListValueFrom(ctx, EditorInterfaceSidebarValue{}.CustomType(ctx), sidebarElements)
	diags.Append(sidebarListDiags...)

	model.Sidebar = sidebarList

	return model, diags
}
