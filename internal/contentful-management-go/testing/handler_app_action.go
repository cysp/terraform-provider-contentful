package cmtesting

import (
	"cmp"
	"context"
	"encoding/json"
	"net/url"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateAppAction(_ context.Context, req *cm.AppActionData, params cm.CreateAppActionParams) (cm.CreateAppActionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	action, validation := newAppAction(params.OrganizationID, params.AppDefinitionID, generateResourceID(), *req)
	if validation != nil {
		return validation, nil
	}

	ts.appActions[[3]string{params.OrganizationID, params.AppDefinitionID, action.Sys.ID}] = action

	return action, nil
}

//nolint:ireturn
func (ts *Handler) UpdateAppAction(_ context.Context, req *cm.AppActionData, params cm.UpdateAppActionParams) (cm.UpdateAppActionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	key := [3]string{params.OrganizationID, params.AppDefinitionID, params.AppActionID}
	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) || ts.appActions[key] == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppAction not found"), nil), nil
	}

	action, validation := newAppAction(params.OrganizationID, params.AppDefinitionID, params.AppActionID, *req)
	if validation != nil {
		return validation, nil
	}

	ts.appActions[key] = action

	return action, nil
}

//nolint:ireturn
func (ts *Handler) GetAppAction(_ context.Context, params cm.GetAppActionParams) (cm.GetAppActionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	action := ts.appActions[[3]string{params.OrganizationID, params.AppDefinitionID, params.AppActionID}]
	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) || action == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppAction not found"), nil), nil
	}

	return action, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppAction(_ context.Context, params cm.DeleteAppActionParams) (cm.DeleteAppActionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	key := [3]string{params.OrganizationID, params.AppDefinitionID, params.AppActionID}
	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) || ts.appActions[key] == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppAction not found"), nil), nil
	}

	delete(ts.appActions, key)

	return &cm.NoContent{}, nil
}

const defaultAppActionCollectionLimit = 100

//nolint:ireturn
func (ts *Handler) GetAppActions(_ context.Context, params cm.GetAppActionsParams) (cm.GetAppActionsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	skip, limit := params.Skip.Or(0), params.Limit.Or(defaultAppActionCollectionLimit)
	if skip < 0 || limit < 1 || limit > 1000 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination: skip must be nonnegative and limit must be between 1 and 1000"), nil), nil
	}

	items := make([]cm.AppAction, 0)

	for key, action := range ts.appActions {
		if key[0] == params.OrganizationID && key[1] == params.AppDefinitionID {
			items = append(items, *action)
		}
	}

	slices.SortFunc(items, func(a, b cm.AppAction) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })
	start := min(skip, int64(len(items)))
	end := min(start+limit, int64(len(items)))

	return &cm.AppActionCollection{
		Sys:   cm.AppActionCollectionSys{Type: cm.AppActionCollectionSysTypeArray},
		Total: cm.NewOptInt(len(items)), Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Items: items[start:end],
	}, nil
}

// newAppAction models the definition forms verified in docs/research/app-framework/actions.md.
// It checks their input shape, not every JSON Schema keyword or invocation outcome.
func newAppAction(organizationID, appDefinitionID, actionID string, req cm.AppActionData) (*cm.AppAction, *cm.ErrorStatusCode) {
	invalid := func(message string) (*cm.AppAction, *cm.ErrorStatusCode) {
		return nil, NewContentfulManagementErrorStatusCodeValidationFailed(&message, nil)
	}
	if req.Name == "" {
		return invalid("Name must not be empty")
	}

	if !validAppActionTarget(req) {
		return invalid("An action requires either an endpoint with an HTTPS URL or a function-invocation with a Function link")
	}

	if !validAppActionSchema(req.ParametersSchema) || !validAppActionSchema(req.ResultSchema) {
		return invalid("A schema must be an object with a type")
	}

	if len(req.Parameters) != 0 {
		var definitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if json.Unmarshal(req.Parameters, &definitions) != nil || definitions == nil {
			return invalid("Parameters must be an array")
		}

		for _, definition := range definitions {
			if definition.ID == "" || definition.Name == "" || !slices.Contains([]string{"Symbol", "Number", "Boolean", "Enum"}, definition.Type) {
				return invalid("Each parameter requires id, name and a supported type")
			}
		}
	}

	parameters := req.Parameters
	switch req.Category {
	case "Custom":
		if (len(parameters) != 0) == (len(req.ParametersSchema) != 0) {
			return invalid("Custom actions require either parameters or parametersSchema")
		}

	case "Entries.v1.0", "Notification.v1.0":
		if len(parameters) != 0 {
			return invalid("Built-in category parameters are read-only")
		}
		// These definitions are literal responses from the live category probes.
		if req.Category == "Entries.v1.0" {
			parameters = []byte(`[{"id":"entryIds","name":"Entry Ids","description":"Ids of the entries you want to trigger the action for","type":"Symbol","required":true}]`)
		} else {
			parameters = []byte(`[{"id":"message","name":"Message","description":"The message being sent to external messaging service","type":"Symbol","required":true},{"id":"recipient","name":"Recipient","description":"","type":"Symbol","required":true}]`)
		}
	default:
		return invalid("Unsupported App Action category")
	}

	return &cm.AppAction{
		Sys:  cm.AppActionSys{ID: actionID, Type: cm.AppActionSysTypeAppAction, Organization: cm.NewOrganizationLink(organizationID), AppDefinition: cm.NewAppDefinitionLink(appDefinitionID)},
		Name: req.Name, Category: req.Category, Type: req.Type, Description: req.Description, URL: req.URL, Function: req.Function,
		Parameters: parameters, ParametersSchema: req.ParametersSchema, ResultSchema: req.ResultSchema,
	}, nil
}

func validAppActionSchema(raw []byte) bool {
	if len(raw) == 0 {
		return true
	}

	var schema map[string]json.RawMessage

	return json.Unmarshal(raw, &schema) == nil && len(schema["type"]) != 0 && string(schema["type"]) != "null"
}

func validAppActionTarget(req cm.AppActionData) bool {
	switch req.Type {
	case "endpoint":
		target, err := url.Parse(req.URL.Value)

		return req.URL.Set && err == nil && target.Scheme == "https" && target.Host != "" && !req.Function.Set
	case "function-invocation":
		return req.Function.Set && req.Function.Value.Sys.ID != "" && !req.URL.Set
	default:
		return false
	}
}
