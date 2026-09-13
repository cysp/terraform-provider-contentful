package cmtesting

import (
	"context"
	"net/http"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppEventSubscription(_ context.Context, params cm.GetAppEventSubscriptionParams) (cm.GetAppEventSubscriptionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	parent := ts.appDefinitions[params.AppDefinitionID]

	value := ts.appEventSubscriptions[[2]string{params.OrganizationID, params.AppDefinitionID}]
	if parent == nil || parent.Sys.Organization.Sys.ID != params.OrganizationID || value == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppEventSubscription not found"), nil), nil
	}

	return value, nil
}

// PutAppEventSubscription models complete document replacement. Function links
// come from the documented SDK shape; their clearing/switching behavior here is
// a fixture assumption, not live Function conformance or execution evidence.
//
//nolint:ireturn
func (ts *Handler) PutAppEventSubscription(_ context.Context, req *cm.AppEventSubscriptionData, params cm.PutAppEventSubscriptionParams) (cm.PutAppEventSubscriptionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	parent := ts.appDefinitions[params.AppDefinitionID]
	if parent == nil || parent.Sys.Organization.Sys.ID != params.OrganizationID {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil), nil
	}

	if len(req.Topics) == 0 {
		return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "ValidationFailed", new("At least one topic is required"), nil), nil
	}

	seen := make(map[string]bool, len(req.Topics))
	for _, topic := range req.Topics {
		if topic == "" || seen[topic] {
			return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "ValidationFailed", new("Topics must be nonempty and unique"), nil), nil
		}

		seen[topic] = true
	}

	if target, ok := req.TargetUrl.Get(); ok {
		if !strings.HasPrefix(target, "https://") || len(target) == len("https://") {
			return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "ValidationFailed", new("An HTTPS target is required"), nil), nil
		}
	} else if !req.Functions.IsSet() {
		return NewContentfulManagementErrorStatusCode(http.StatusBadRequest, "BadRequest", new("A target is required for HTTP subscriptions"), nil), nil
	}

	key := [2]string{params.OrganizationID, params.AppDefinitionID}
	prior := ts.appEventSubscriptions[key]
	value := cm.AppEventSubscription{
		Sys:    cm.AppEventSubscriptionSys{Type: cm.AppEventSubscriptionSysTypeAppEventSubscription, Organization: cm.NewOrganizationLink(params.OrganizationID), AppDefinition: cm.NewAppDefinitionLink(params.AppDefinitionID)},
		Topics: req.Topics, TargetUrl: req.TargetUrl, Functions: req.Functions,
	}

	ts.appEventSubscriptions[key] = &value
	if prior == nil {
		result := cm.PutAppEventSubscriptionCreated(value)

		return &result, nil
	}

	result := cm.PutAppEventSubscriptionOK(value)

	return &result, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppEventSubscription(_ context.Context, params cm.DeleteAppEventSubscriptionParams) (cm.DeleteAppEventSubscriptionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	key := [2]string{params.OrganizationID, params.AppDefinitionID}

	parent := ts.appDefinitions[params.AppDefinitionID]
	if parent == nil || parent.Sys.Organization.Sys.ID != params.OrganizationID || ts.appEventSubscriptions[key] == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppEventSubscription not found"), nil), nil
	}

	delete(ts.appEventSubscriptions, key)

	return &cm.NoContent{}, nil
}
