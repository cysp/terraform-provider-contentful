package provider_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteMutationPlanFixturesEncodeAsFullyKnownFrameworkValues(t *testing.T) {
	t.Parallel()

	for _, fixture := range []struct {
		name   string
		plan   any
		schema schema.Schema
	}{
		{name: "role", plan: completeRoleMutationPlan(), schema: RoleResourceSchema(t.Context())},
		{name: "editor interface", plan: completeEditorInterfaceMutationPlan(), schema: EditorInterfaceResourceSchema(t.Context())},
		{name: "webhook", plan: completeWebhookMutationPlan(), schema: WebhookResourceSchema(t.Context())},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			plan := tfsdk.Plan{Schema: fixture.schema}
			planDiags := plan.Set(t.Context(), fixture.plan)
			require.False(t, planDiags.HasError(), "%v", planDiags)
			assert.False(t, plan.Raw.IsNull())
			assert.True(t, plan.Raw.IsFullyKnown())
		})
	}
}

func TestAccCompleteMutationConfigurationsPlan(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]string{
		"role":             completeRoleMutationConfig,
		"editor interface": completeEditorInterfaceMutationConfig,
		"webhook":          completeWebhookMutationConfig,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ContentfulProviderMockedResourceTest(t, http.NotFoundHandler(), resource.TestCase{Steps: []resource.TestStep{{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}}})
		})
	}
}

const completeRoleMutationConfig = `
resource "contentful_role" "test" {
  space_id    = "space"
  name        = "Planned role"
  description = "Planned description"

  permissions = {
    Entry = ["read", "create"]
  }

  policies = [{
    actions    = ["read", "create"]
    constraint = jsonencode({ a = 1, b = 2 })
    effect     = "allow"
  }]

  timeouts = {
    create = "1m"
    read   = "1m"
    update = "1m"
    delete = "1m"
  }
}
`

const completeEditorInterfaceMutationConfig = `
resource "contentful_editor_interface" "test" {
  space_id        = "space"
  environment_id  = "master"
  content_type_id = "article"

  editor_layout = [{
    group = {
      group_id = "main"
      name     = "Main"
      items = [{
        group = {
          group_id = "nested"
          name     = "Nested"
          items    = [{ field = { field_id = "title" } }]
        }
      }]
    }
  }]

  controls = [{
    field_id         = "title"
    widget_namespace = "builtin"
    widget_id        = "singleLine"
    settings         = jsonencode({ help = "text" })
  }]

  group_controls = [{
    group_id         = "main"
    widget_namespace = "builtin"
    widget_id        = "fieldset"
    settings         = jsonencode({ collapsed = false, help = "text" })
  }]

  sidebar = [{
    widget_namespace = "sidebar-builtin"
    widget_id        = "publication-widget"
    settings         = jsonencode({ collapsed = false, help = "text" })
    disabled         = false
  }]

  timeouts = {
    create = "1m"
    read   = "1m"
    update = "1m"
  }
}
`

const completeWebhookMutationConfig = `
resource "contentful_webhook" "test" {
  space_id = "space"
  active   = true
  name     = "Planned webhook"
  url      = "https://example.com/planned"
  topics   = ["Entry.create", "Entry.save"]

  filters = [{
    not = {
      equals = {
        doc   = "sys.type"
        value = "Entry"
      }
    }
  }]

  http_basic_username = "username"
  http_basic_password = "password"

  headers = {
    X-Test = {
      value  = "header-value"
      secret = false
    }
  }

  transformation = {
    method                 = "POST"
    content_type           = "application/json"
    include_content_length = true
    body                   = jsonencode({ a = 1, b = 2 })
  }

  timeouts = {
    create = "1m"
    read   = "1m"
    update = "1m"
    delete = "1m"
  }
}
`

func completeRoleMutationPlan() RoleModel {
	return RoleModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "role"),
		RoleIdentityModel: RoleIdentityModel{
			SpaceID: types.StringValue("space"),
			RoleID:  types.StringValue("role"),
		},
		Name:        types.StringValue("Planned role"),
		Description: types.StringValue("Planned description"),
		Permissions: NewTypedMap(map[string]TypedList[types.String]{
			"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringValue("create")}),
		}),
		Policies: NewTypedList([]TypedObject[RolePolicyValue]{
			rolePolicyValue("allow", []string{"read", "create"}, jsontypes.NewNormalizedValue(`{"b":2,"a":1}`)),
		}),
		Timeouts: TimeoutsNull(),
	}
}

func completeRoleMutationResponse() cm.Role {
	return cm.Role{
		Sys:         cm.NewRoleSys("space", "role"),
		Name:        "Planned role",
		Description: cm.NewOptNilString("Planned description"),
		Permissions: cm.RolePermissions{
			"Entry": cm.NewStringArrayRolePermissionsItem([]string{"create", "read"}),
		},
		Policies: []cm.RolePoliciesItem{
			rolePolicyResponse("allow", []string{"create", "read"}, `{"a":1,"b":2}`),
		},
	}
}

func completeEditorInterfaceMutationPlan() EditorInterfaceModel {
	return EditorInterfaceModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "master", "article"),
		EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("master"),
			ContentTypeID: types.StringValue("article"),
		},
		EditorLayout: NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemValue]{
			editorLayoutPlanGroup("main", "Main", NewTypedList([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]{})),
		}),
		Controls: NewTypedList([]TypedObject[EditorInterfaceControlValue]{
			NewTypedObject(EditorInterfaceControlValue{
				FieldID:         types.StringValue("title"),
				WidgetNamespace: types.StringValue("builtin"),
				WidgetID:        types.StringValue("singleLine"),
				Settings:        jsontypes.NewNormalizedValue(`{"b":2,"a":1}`),
			}),
		}),
		GroupControls: NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{
			NewTypedObject(EditorInterfaceGroupControlValue{
				GroupID:         types.StringValue("main"),
				WidgetNamespace: types.StringValue("builtin"),
				WidgetID:        types.StringValue("fieldset"),
				Settings:        jsontypes.NewNormalizedValue(`{"help":"text","collapsed":false}`),
			}),
		}),
		Sidebar: NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{
			NewTypedObject(EditorInterfaceSidebarValue{
				WidgetNamespace: types.StringValue("sidebar-builtin"),
				WidgetID:        types.StringValue("publication-widget"),
				Settings:        jsontypes.NewNormalizedValue(`{"help":"text","collapsed":false}`),
				Disabled:        types.BoolValue(false),
			}),
		}),
		Timeouts: updateOnlyTimeoutsNull(),
	}
}

func completeEditorInterfaceMutationResponse() cm.EditorInterface {
	return cm.EditorInterface{
		Sys: cm.NewEditorInterfaceSys("space", "master", "article", "default"),
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			editorLayoutResponseGroup("main", "Main"),
		}),
		Controls: cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{{
			FieldId:         "title",
			WidgetNamespace: cm.NewOptString("builtin"),
			WidgetId:        cm.NewOptString("singleLine"),
			Settings:        []byte(`{"a":1,"b":2}`),
		}}),
		GroupControls: cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{{
			GroupId:         "main",
			WidgetNamespace: cm.NewOptString("builtin"),
			WidgetId:        cm.NewOptString("fieldset"),
			Settings:        []byte(`{"collapsed":false,"help":"text"}`),
		}}),
		Sidebar: cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{{
			WidgetNamespace: "sidebar-builtin",
			WidgetId:        "publication-widget",
			Settings:        []byte(`{"collapsed":false,"help":"text"}`),
			Disabled:        cm.NewOptBool(false),
		}}),
	}
}

func completeWebhookMutationPlan() WebhookModel {
	header := NewTypedObject(WebhookHeaderValue{Value: types.StringValue("header-value"), Secret: types.BoolValue(false)})
	transformation := NewTypedObject(WebhookTransformationValue{
		Method:               types.StringValue("POST"),
		ContentType:          types.StringValue("application/json"),
		IncludeContentLength: types.BoolValue(true),
		Body:                 jsontypes.NewNormalizedValue(`{"b":2,"a":1}`),
	})
	filters := NewTypedList([]TypedObject[WebhookFilterValue]{
		NewTypedObject(WebhookFilterValue{
			Not: NewTypedObjectNull[WebhookFilterNotValue](),
			Equals: NewTypedObject(WebhookFilterEqualsValue{
				Doc:   types.StringValue("sys.type"),
				Value: types.StringValue("Entry"),
			}),
			In:     NewTypedObjectNull[WebhookFilterInValue](),
			Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
		}),
	})

	return WebhookModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "webhook"),
		WebhookIdentityModel: WebhookIdentityModel{
			SpaceID:   types.StringValue("space"),
			WebhookID: types.StringValue("webhook"),
		},
		Name:              types.StringValue("Planned webhook"),
		URL:               types.StringValue("https://example.com/planned"),
		Topics:            NewTypedList([]types.String{types.StringValue("Entry.create"), types.StringValue("Entry.save")}),
		Filters:           filters,
		HTTPBasicPassword: types.StringValue("password"),
		HTTPBasicUsername: types.StringValue("username"),
		Headers:           NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"X-Test": header}),
		Transformation:    transformation,
		Active:            types.BoolValue(true),
		Timeouts:          TimeoutsNull(),
	}
}

func completeWebhookMutationResponse() cm.WebhookDefinition {
	return cm.WebhookDefinition{
		Sys:               cm.NewWebhookDefinitionSys("space", "webhook"),
		Name:              "Planned webhook",
		URL:               "https://example.com/planned",
		Topics:            []string{"Entry.create", "Entry.save"},
		Filters:           cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{cmWebhookEqualsFilter("Entry")}),
		HttpBasicUsername: cm.NewOptNilString("username"),
		Headers: cm.WebhookDefinitionHeaders{{
			Key:    "X-Test",
			Value:  cm.NewOptString("header-value"),
			Secret: cm.NewOptBool(false),
		}},
		Transformation: cm.NewOptNilWebhookDefinitionTransformation(cm.WebhookDefinitionTransformation{
			Method:               cm.NewOptString("POST"),
			ContentType:          cm.NewOptString("application/json"),
			IncludeContentLength: cm.NewOptBool(true),
			Body:                 []byte(`{"a":1,"b":2}`),
		}),
		Active: cm.NewOptBool(true),
	}
}

func cmWebhookEqualsFilter(value string) cm.WebhookDefinitionFilter {
	return cm.WebhookDefinitionFilter{
		Equals: cm.WebhookDefinitionFilterEquals{
			[]byte(`{"doc":"sys.type"}`),
			[]byte(`"` + value + `"`),
		},
	}
}

func TestCompleteRolePlanRejectsScalarContradictionWithoutPartialOverlay(t *testing.T) {
	t.Parallel()

	plan := completeRoleMutationPlan()
	response := completeRoleMutationResponse()
	response.Sys = cm.NewRoleSys("other-space", "other-role")
	response.Name = "Response role"

	state, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"space_id", "role_id", "name"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "space/role", state.ID.ValueString())
	assert.Equal(t, "Response role", state.Name.ValueString())
	assert.Equal(t, []types.String{types.StringValue("create"), types.StringValue("read")}, state.Permissions.Elements()["Entry"].Elements())
}

func TestCompleteRolePlanRestoresCanonicalisedEquivalentResponse(t *testing.T) {
	t.Parallel()

	plan := completeRoleMutationPlan()
	state, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(t.Context(), completeRoleMutationResponse(), plan)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, state.Permissions.Equal(plan.Permissions))
	assert.True(t, state.Policies.Equal(plan.Policies))
}

func TestCompleteRolePlanRejectsIndependentAttributeContradictions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate             func(*testing.T, *RoleModel, *cm.Role)
		expectedPath       string
		stateValue         func(RoleModel) string
		expectedStateValue string
	}{
		"description": {
			mutate: func(_ *testing.T, _ *RoleModel, response *cm.Role) {
				response.Description = cm.NewOptNilString("Response description")
			},
			expectedPath:       "description",
			stateValue:         func(state RoleModel) string { return state.Description.ValueString() },
			expectedStateValue: "Response description",
		},
		"legacy ID": {
			mutate: func(_ *testing.T, plan *RoleModel, _ *cm.Role) {
				plan.ID = types.StringValue("different/legacy/id")
			},
			expectedPath:       "id",
			stateValue:         func(state RoleModel) string { return state.ID.ValueString() },
			expectedStateValue: "space/role",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := completeRoleMutationPlan()
			response := completeRoleMutationResponse()
			test.mutate(t, &plan, &response)

			state, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(t.Context(), response, plan)

			assert.Empty(t, responseDiags)
			require.True(t, consistencyDiags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, consistencyDiags))
			assert.Equal(t, test.expectedStateValue, test.stateValue(state))
		})
	}
}

func TestCompleteEditorInterfacePlanRejectsNestedContradictionAndResponseRetarget(t *testing.T) {
	t.Parallel()

	plan := completeEditorInterfaceMutationPlan()
	response := completeEditorInterfaceMutationResponse()
	response.Sys = cm.NewEditorInterfaceSys("other-space", "master", "article", "default")
	controls, ok := response.Controls.Get()
	require.True(t, ok)

	controls[0].WidgetId = cm.NewOptString("response-widget")
	response.Controls.SetTo(controls)
	sidebar, ok := response.Sidebar.Get()
	require.True(t, ok)

	sidebar[0].Disabled = cm.NewOptBool(true)
	response.Sidebar.SetTo(sidebar)

	state, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"space_id", "controls", "sidebar"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "space", state.SpaceID.ValueString())
	assert.Equal(t, "space/master/article", state.ID.ValueString())
	assert.Equal(t, "response-widget", state.Controls.Elements()[0].Value().WidgetID.ValueString())
	assert.True(t, state.Sidebar.Elements()[0].Value().Disabled.ValueBool())
}

func TestCompleteEditorInterfacePlanRestoresCanonicalisedEquivalentResponse(t *testing.T) {
	t.Parallel()

	plan := completeEditorInterfaceMutationPlan()
	state, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), completeEditorInterfaceMutationResponse(), plan)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, state.EditorLayout.Equal(plan.EditorLayout))
	assert.True(t, state.Controls.Equal(plan.Controls))
	assert.True(t, state.GroupControls.Equal(plan.GroupControls))
	assert.True(t, state.Sidebar.Equal(plan.Sidebar))
}

func TestCompleteEditorInterfacePlanDistinguishesNullFromEmptyCollections(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		makePlanNull        func(*EditorInterfaceModel)
		makeResponseEmpty   func(*cm.EditorInterface)
		assertResponseState func(*testing.T, EditorInterfaceModel)
	}{
		"controls": {
			makePlanNull: func(plan *EditorInterfaceModel) {
				plan.Controls = NewTypedListNull[TypedObject[EditorInterfaceControlValue]]()
			},
			makeResponseEmpty: func(response *cm.EditorInterface) {
				response.Controls = cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{})
			},
			assertResponseState: func(t *testing.T, state EditorInterfaceModel) {
				t.Helper()
				assert.Equal(t, NewTypedList([]TypedObject[EditorInterfaceControlValue]{}), state.Controls)
			},
		},
		"group_controls": {
			makePlanNull: func(plan *EditorInterfaceModel) {
				plan.GroupControls = NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]]()
			},
			makeResponseEmpty: func(response *cm.EditorInterface) {
				response.GroupControls = cm.NewOptNilEditorInterfaceGroupControlsItemArray([]cm.EditorInterfaceGroupControlsItem{})
			},
			assertResponseState: func(t *testing.T, state EditorInterfaceModel) {
				t.Helper()
				assert.Equal(t, NewTypedList([]TypedObject[EditorInterfaceGroupControlValue]{}), state.GroupControls)
			},
		},
		"sidebar": {
			makePlanNull: func(plan *EditorInterfaceModel) {
				plan.Sidebar = NewTypedListNull[TypedObject[EditorInterfaceSidebarValue]]()
			},
			makeResponseEmpty: func(response *cm.EditorInterface) {
				response.Sidebar = cm.NewOptNilEditorInterfaceSidebarItemArray([]cm.EditorInterfaceSidebarItem{})
			},
			assertResponseState: func(t *testing.T, state EditorInterfaceModel) {
				t.Helper()
				assert.Equal(t, NewTypedList([]TypedObject[EditorInterfaceSidebarValue]{}), state.Sidebar)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := completeEditorInterfaceMutationPlan()
			response := completeEditorInterfaceMutationResponse()

			test.makePlanNull(&plan)
			test.makeResponseEmpty(&response)

			state, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, plan)

			assert.Empty(t, responseDiags)
			require.True(t, consistencyDiags.HasError())
			assert.Equal(t, []string{name}, attributeDiagnosticPaths(t, consistencyDiags))
			test.assertResponseState(t, state)
		})
	}
}

func TestCompleteEditorInterfacePlanRejectsIndependentAttributeContradictions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate             func(*testing.T, *EditorInterfaceModel, *cm.EditorInterface)
		expectedPath       string
		stateValue         func(EditorInterfaceModel) string
		expectedStateValue string
	}{
		"environment ID": {
			mutate: func(_ *testing.T, _ *EditorInterfaceModel, response *cm.EditorInterface) {
				response.Sys = cm.NewEditorInterfaceSys("space", "other-environment", "article", "default")
			},
			expectedPath:       "environment_id",
			stateValue:         func(state EditorInterfaceModel) string { return state.EnvironmentID.ValueString() },
			expectedStateValue: "master",
		},
		"content type ID": {
			mutate: func(_ *testing.T, _ *EditorInterfaceModel, response *cm.EditorInterface) {
				response.Sys = cm.NewEditorInterfaceSys("space", "master", "other-content-type", "default")
			},
			expectedPath:       "content_type_id",
			stateValue:         func(state EditorInterfaceModel) string { return state.ContentTypeID.ValueString() },
			expectedStateValue: "article",
		},
		"legacy ID": {
			mutate: func(_ *testing.T, plan *EditorInterfaceModel, _ *cm.EditorInterface) {
				plan.ID = types.StringValue("different/legacy/id")
			},
			expectedPath:       "id",
			stateValue:         func(state EditorInterfaceModel) string { return state.ID.ValueString() },
			expectedStateValue: "space/master/article",
		},
		"group controls": {
			mutate: func(t *testing.T, _ *EditorInterfaceModel, response *cm.EditorInterface) {
				t.Helper()

				groupControls, ok := response.GroupControls.Get()
				require.True(t, ok)

				groupControls[0].WidgetId = cm.NewOptString("response-widget")
				response.GroupControls.SetTo(groupControls)
			},
			expectedPath: "group_controls",
			stateValue: func(state EditorInterfaceModel) string {
				return state.GroupControls.Elements()[0].Value().WidgetID.ValueString()
			},
			expectedStateValue: "response-widget",
		},
		"group control settings": {
			mutate: func(t *testing.T, _ *EditorInterfaceModel, response *cm.EditorInterface) {
				t.Helper()

				groupControls, ok := response.GroupControls.Get()
				require.True(t, ok)

				groupControls[0].Settings = []byte(`{"different":true}`)
				response.GroupControls.SetTo(groupControls)
			},
			expectedPath: "group_controls",
			stateValue: func(state EditorInterfaceModel) string {
				return state.GroupControls.Elements()[0].Value().Settings.ValueString()
			},
			expectedStateValue: `{"different":true}`,
		},
		"sidebar settings": {
			mutate: func(t *testing.T, _ *EditorInterfaceModel, response *cm.EditorInterface) {
				t.Helper()

				sidebar, ok := response.Sidebar.Get()
				require.True(t, ok)

				sidebar[0].Settings = []byte(`{"different":true}`)
				response.Sidebar.SetTo(sidebar)
			},
			expectedPath: "sidebar",
			stateValue: func(state EditorInterfaceModel) string {
				return state.Sidebar.Elements()[0].Value().Settings.ValueString()
			},
			expectedStateValue: `{"different":true}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := completeEditorInterfaceMutationPlan()
			response := completeEditorInterfaceMutationResponse()
			test.mutate(t, &plan, &response)

			state, responseDiags, consistencyDiags := ReconcileEditorInterfaceMutationResponse(t.Context(), response, plan)

			assert.Empty(t, responseDiags)
			require.True(t, consistencyDiags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, consistencyDiags))
			assert.Equal(t, test.expectedStateValue, test.stateValue(state))
		})
	}
}

func TestCompleteWebhookPlanRejectsScalarContradictionWithoutPartialOverlay(t *testing.T) {
	t.Parallel()

	plan := completeWebhookMutationPlan()
	response := completeWebhookMutationResponse()
	response.Sys = cm.NewWebhookDefinitionSys("other-space", "other-webhook")
	response.URL = "https://example.com/response"
	response.Filters = cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
		cmWebhookEqualsFilter("Asset"),
		cmWebhookEqualsFilter("Entry"),
	})
	plan.Filters = NewTypedList([]TypedObject[WebhookFilterValue]{
		NewTypedObject(WebhookFilterValue{
			Not:    NewTypedObjectNull[WebhookFilterNotValue](),
			Equals: NewTypedObject(WebhookFilterEqualsValue{Doc: types.StringValue("sys.type"), Value: types.StringValue("Entry")}),
			In:     NewTypedObjectNull[WebhookFilterInValue](), Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
		}),
		NewTypedObject(WebhookFilterValue{
			Not:    NewTypedObjectNull[WebhookFilterNotValue](),
			Equals: NewTypedObject(WebhookFilterEqualsValue{Doc: types.StringValue("sys.type"), Value: types.StringValue("Asset")}),
			In:     NewTypedObjectNull[WebhookFilterInValue](), Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
		}),
	})

	state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"space_id", "webhook_id", "url"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.Equal(t, "space/webhook", state.ID.ValueString())
	assert.Equal(t, "https://example.com/response", state.URL.ValueString())
	assert.Equal(t, "Asset", state.Filters.Elements()[0].Value().Equals.Value().Value.ValueString())
	assert.Equal(t, plan.HTTPBasicPassword, state.HTTPBasicPassword)
}

func TestCompleteWebhookPlanRestoresCanonicalisedEquivalentResponse(t *testing.T) {
	t.Parallel()

	plan := completeWebhookMutationPlan()
	state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), completeWebhookMutationResponse(), plan)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.True(t, state.Filters.Equal(plan.Filters))
	assert.True(t, state.Transformation.Equal(plan.Transformation))
	assert.Equal(t, plan.HTTPBasicPassword, state.HTTPBasicPassword)
}

func TestCompleteWebhookPlanRejectsIndependentAttributeContradictions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate             func(*testing.T, *WebhookModel, *cm.WebhookDefinition)
		expectedPath       string
		stateValue         func(WebhookModel) any
		expectedStateValue any
	}{
		"legacy ID": {
			mutate: func(_ *testing.T, plan *WebhookModel, _ *cm.WebhookDefinition) {
				plan.ID = types.StringValue("different/legacy/id")
			},
			expectedPath:       "id",
			stateValue:         func(state WebhookModel) any { return state.ID.ValueString() },
			expectedStateValue: "space/webhook",
		},
		"name": {
			mutate: func(_ *testing.T, _ *WebhookModel, response *cm.WebhookDefinition) {
				response.Name = "Response webhook"
			},
			expectedPath:       "name",
			stateValue:         func(state WebhookModel) any { return state.Name.ValueString() },
			expectedStateValue: "Response webhook",
		},
		"topics": {
			mutate: func(_ *testing.T, _ *WebhookModel, response *cm.WebhookDefinition) {
				response.Topics = []string{"Entry.save", "Entry.create"}
			},
			expectedPath:       "topics",
			stateValue:         func(state WebhookModel) any { return state.Topics.Elements() },
			expectedStateValue: []types.String{types.StringValue("Entry.save"), types.StringValue("Entry.create")},
		},
		"Basic username": {
			mutate: func(_ *testing.T, _ *WebhookModel, response *cm.WebhookDefinition) {
				response.HttpBasicUsername = cm.NewOptNilString("response-username")
			},
			expectedPath:       "http_basic_username",
			stateValue:         func(state WebhookModel) any { return state.HTTPBasicUsername.ValueString() },
			expectedStateValue: "response-username",
		},
		"transformation": {
			mutate: func(t *testing.T, _ *WebhookModel, response *cm.WebhookDefinition) {
				t.Helper()

				transformation, ok := response.Transformation.Get()
				require.True(t, ok)

				transformation.Method = cm.NewOptString("PUT")
				response.Transformation.SetTo(transformation)
			},
			expectedPath:       "transformation",
			stateValue:         func(state WebhookModel) any { return state.Transformation.Value().Method.ValueString() },
			expectedStateValue: "PUT",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := completeWebhookMutationPlan()
			response := completeWebhookMutationResponse()
			test.mutate(t, &plan, &response)

			state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

			assert.Empty(t, responseDiags)
			require.True(t, consistencyDiags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, consistencyDiags))
			assert.Equal(t, test.expectedStateValue, test.stateValue(state))
		})
	}
}

func TestWebhookUnknownOptionalComputedPlanUsesResponse(t *testing.T) {
	t.Parallel()

	plan := completeWebhookMutationPlan()
	plan.Headers = NewTypedMapUnknown[TypedObject[WebhookHeaderValue]]()
	response := completeWebhookMutationResponse()
	response.Headers = cm.WebhookDefinitionHeaders{{Key: "Response", Value: cm.NewOptString("response"), Secret: cm.NewOptBool(true)}}

	state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	assert.Empty(t, consistencyDiags)
	assert.Contains(t, state.Headers.Elements(), "Response")
	assert.Equal(t, "response", state.Headers.Elements()["Response"].Value().Value.ValueString())
	assert.True(t, state.Headers.Elements()["Response"].Value().Secret.ValueBool())
}

func TestWebhookKnownOptionalComputedPlansConstrainMutationResponse(t *testing.T) {
	t.Parallel()

	plan := completeWebhookMutationPlan()
	response := completeWebhookMutationResponse()
	response.Active = cm.NewOptBool(false)
	response.Headers = cm.WebhookDefinitionHeaders{{Key: "Response", Value: cm.NewOptString("response"), Secret: cm.NewOptBool(true)}}

	state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

	assert.Empty(t, responseDiags)
	require.True(t, consistencyDiags.HasError())
	assert.Equal(t, []string{"headers", "active"}, attributeDiagnosticPaths(t, consistencyDiags))
	assert.False(t, state.Active.ValueBool())
	assert.Contains(t, state.Headers.Elements(), "Response")
	assert.NotContains(t, state.Headers.Elements(), "X-Test")
}
