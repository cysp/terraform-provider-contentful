package provider_test

import (
	"fmt"
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestEditorInterfaceMutationReconciliationProperties(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rapid.Check(t, func(t *rapid.T) {
		count := rapid.IntRange(2, 5).Draw(t, "element count")
		plannedControls := make([]TypedObject[EditorInterfaceControlValue], count)
		plannedGroups := make([]TypedObject[EditorInterfaceGroupControlValue], count)
		plannedSidebar := make([]TypedObject[EditorInterfaceSidebarValue], count)
		layout := make([]TypedObject[EditorInterfaceEditorLayoutItemValue], count)
		returnedControls := make([]TypedObject[EditorInterfaceControlValue], count)
		returnedGroups := make([]TypedObject[EditorInterfaceGroupControlValue], count)
		returnedSidebar := make([]TypedObject[EditorInterfaceSidebarValue], count)
		controls := make([]cm.EditorInterfaceControlsItem, count)
		groups := make([]cm.EditorInterfaceGroupControlsItem, count)
		sidebar := make([]cm.EditorInterfaceSidebarItem, count)
		responseLayout := make([]cm.EditorInterfaceEditorLayoutItem, count)

		reversedSettings := make([]string, count)
		for index := range count {
			plannedJSON, returnedJSON, reversedJSON := drawReconciliationJSON(t)
			reversedSettings[index] = reversedJSON
			fieldID := fmt.Sprintf("field-%d", index)
			disabled := rapid.Bool().Draw(t, "sidebar disabled")
			widgetID, wireWidgetID := types.StringNull(), cm.OptString{}

			switch rapid.IntRange(0, 2).Draw(t, "optional widget ID") {
			case 1:
				widgetID, wireWidgetID = types.StringValue(""), cm.NewOptString("")
			case 2:
				widgetID, wireWidgetID = types.StringValue("singleLine"), cm.NewOptString("singleLine")
			}

			plannedControls[index] = NewTypedObject(EditorInterfaceControlValue{
				FieldID: types.StringValue(fieldID), WidgetNamespace: types.StringNull(), WidgetID: widgetID, Settings: jsontypes.NewNormalizedValue(plannedJSON),
			})
			plannedGroups[index] = NewTypedObject(EditorInterfaceGroupControlValue{
				GroupID: types.StringValue(fieldID), WidgetNamespace: types.StringNull(), WidgetID: widgetID, Settings: jsontypes.NewNormalizedValue(plannedJSON),
			})
			plannedSidebar[index] = NewTypedObject(EditorInterfaceSidebarValue{
				WidgetNamespace: types.StringValue("sidebar-builtin"), WidgetID: types.StringValue(fieldID),
				Settings: jsontypes.NewNormalizedValue(plannedJSON), Disabled: types.BoolValue(disabled),
			})
			layout[index] = editorLayoutPlanGroup(fieldID, fieldID, NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{}))

			returnedControls[index] = NewTypedObject(EditorInterfaceControlValue{
				FieldID: types.StringValue(fieldID), WidgetNamespace: types.StringNull(), WidgetID: widgetID, Settings: jsontypes.NewNormalizedValue(returnedJSON),
			})
			returnedGroups[index] = NewTypedObject(EditorInterfaceGroupControlValue{
				GroupID: types.StringValue(fieldID), WidgetNamespace: types.StringNull(), WidgetID: widgetID, Settings: jsontypes.NewNormalizedValue(returnedJSON),
			})
			returnedSidebar[index] = NewTypedObject(EditorInterfaceSidebarValue{
				WidgetNamespace: types.StringValue("sidebar-builtin"), WidgetID: types.StringValue(fieldID),
				Settings: jsontypes.NewNormalizedValue(returnedJSON), Disabled: types.BoolValue(disabled),
			})
			controls[index] = cm.EditorInterfaceControlsItem{FieldId: fieldID, WidgetId: wireWidgetID, Settings: []byte(returnedJSON)}
			groups[index] = cm.EditorInterfaceGroupControlsItem{GroupId: fieldID, WidgetId: wireWidgetID, Settings: []byte(returnedJSON)}
			sidebar[index] = cm.EditorInterfaceSidebarItem{
				WidgetNamespace: "sidebar-builtin", WidgetId: fieldID, Settings: []byte(returnedJSON), Disabled: cm.NewOptBool(disabled),
			}
			responseLayout[index] = editorLayoutResponseGroup(fieldID, fieldID)
		}

		plan := completeEditorInterfaceMutationPlan()
		plan.EditorLayout = NewTypedList(layout)
		plan.Controls, plan.GroupControls, plan.Sidebar = NewTypedList(plannedControls), NewTypedList(plannedGroups), NewTypedList(plannedSidebar)

		expectedEquivalent := EditorInterfaceModel{
			IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "master", "article"),
			SpaceID:         types.StringValue("space"), EnvironmentID: types.StringValue("master"), ContentTypeID: types.StringValue("article"),
			EditorLayout:  plan.EditorLayout,
			Controls:      plan.Controls,
			GroupControls: plan.GroupControls,
			Sidebar:       plan.Sidebar,
		}
		if rapid.Bool().Draw(t, "response-owned controls") {
			plan.Controls = NewTypedListUnknown[TypedObject[EditorInterfaceControlValue]]()
			expectedEquivalent.Controls = NewTypedList(returnedControls)
		}

		if rapid.Bool().Draw(t, "response-owned group controls") {
			plan.GroupControls = NewTypedListUnknown[TypedObject[EditorInterfaceGroupControlValue]]()
			expectedEquivalent.GroupControls = NewTypedList(returnedGroups)
		}

		if rapid.Bool().Draw(t, "response-owned sidebar") {
			plan.Sidebar = NewTypedListUnknown[TypedObject[EditorInterfaceSidebarValue]]()
			expectedEquivalent.Sidebar = NewTypedList(returnedSidebar)
		}

		response := cm.EditorInterface{
			Sys:          cm.NewEditorInterfaceSys("space", "master", "article", "default"),
			EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray(responseLayout),
			Controls:     cm.NewOptNilEditorInterfaceControlsItemArray(controls), GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray(groups),
			Sidebar: cm.NewOptNilEditorInterfaceSidebarItemArray(sidebar),
		}
		state, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags)
		assert.Equal(t, expectedEquivalent, state, "known collections restore exact Plan; unknown collections take response")

		expected := EditorInterfaceModel{
			IDIdentityModel:              expectedEquivalent.IDIdentityModel,
			EditorInterfaceIdentityModel: expectedEquivalent.EditorInterfaceIdentityModel,
			EditorLayout:                 NewTypedList(slices.Clone(layout)),
			Controls:                     NewTypedList(returnedControls),
			GroupControls:                NewTypedList(returnedGroups),
			Sidebar:                      NewTypedList(returnedSidebar),
		}
		index := rapid.IntRange(0, count-1).Draw(t, "changed index")
		next := (index + 1) % count

		var expectedPath string

		switch rapid.IntRange(0, 7).Draw(t, "contradiction") {
		case 0:
			plan.Controls = NewTypedList(plannedControls)
			controls[index], controls[next] = controls[next], controls[index]
			returnedControls[index], returnedControls[next] = returnedControls[next], returnedControls[index]
			expectedPath = "controls"
		case 1:
			plan.GroupControls = NewTypedList(plannedGroups)
			groups[index], groups[next] = groups[next], groups[index]
			returnedGroups[index], returnedGroups[next] = returnedGroups[next], returnedGroups[index]
			expectedPath = "group_controls"
		case 2:
			plan.Sidebar = NewTypedList(plannedSidebar)
			sidebar[index], sidebar[next] = sidebar[next], sidebar[index]
			returnedSidebar[index], returnedSidebar[next] = returnedSidebar[next], returnedSidebar[index]
			expectedPath = "sidebar"
		case 3:
			responseLayout[index], responseLayout[next] = responseLayout[next], responseLayout[index]
			expected.EditorLayout.Elements()[index], expected.EditorLayout.Elements()[next] = layout[next], layout[index]
			expectedPath = "editor_layout"
		case 4:
			plan.Controls = NewTypedList(plannedControls)
			controls[index].Settings = []byte(reversedSettings[index])
			changed := returnedControls[index].Value()
			changed.Settings = jsontypes.NewNormalizedValue(reversedSettings[index])
			returnedControls[index] = NewTypedObject(changed)
			expectedPath = "controls"
		case 5:
			plan.Sidebar = NewTypedList(plannedSidebar)
			changed := returnedSidebar[index].Value()
			changed.Disabled = types.BoolValue(!changed.Disabled.ValueBool())
			sidebar[index].Disabled = cm.NewOptBool(changed.Disabled.ValueBool())
			returnedSidebar[index] = NewTypedObject(changed)
			expectedPath = "sidebar"
		case 6:
			response.Sys.Environment.Sys.ID = "other-environment"
			expectedPath = "environment_id"
		case 7:
			plan.GroupControls = NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]]()
			response.GroupControls.SetTo([]cm.EditorInterfaceGroupControlsItem{})
			expected.GroupControls = NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{})
			expectedPath = "group_controls"
		}

		state, responseDiags, consistencyDiags = ReconcileEditorInterfaceMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags, expectedPath)
		assert.Equal(t, expected, state, "one ordered or nested contradiction must suppress every JSON overlay")
	})
}
