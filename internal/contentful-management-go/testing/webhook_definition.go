package cmtesting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
)

func NewWebhookDefinitionFromFields(spaceID, webhookDefinitionID string, webhookDefinitionFields cm.WebhookDefinitionData) cm.WebhookDefinition {
	webhookDefinition := cm.WebhookDefinition{
		Sys:               cm.NewWebhookDefinitionSys(spaceID, webhookDefinitionID),
		HttpBasicUsername: cm.NewOptNilStringNull(),
		HttpBasicPassword: cm.NewOptNilStringNull(),
	}

	setWebhookDefinitionFields(&webhookDefinition, webhookDefinitionFields)

	return webhookDefinition
}

// applyWebhookDefinitionUpdate resolves retained values after request validation
// and the version check. Commit only after every retained name was found.
func applyWebhookDefinitionUpdate(webhook *cm.WebhookDefinition, fields cm.WebhookDefinitionData) *cm.ErrorStatusCode {
	updated := *webhook
	setWebhookDefinitionFields(&updated, fields)

	byName := make(map[string]cm.WebhookDefinitionHeader, len(webhook.Headers))
	for _, header := range webhook.Headers {
		byName[header.Key] = header
	}

	for i, header := range updated.Headers {
		stored, exists := byName[header.Key]
		if header.Value.IsSet() {
			if exists {
				updated.Headers[i].Secret = stored.Secret
			}

			continue
		}

		if !exists {
			return missingWebhookHeaderValueError(header.Key)
		}

		updated.Headers[i] = stored
	}

	if !reflect.DeepEqual(*webhook, updated) {
		updated.Sys.Version++
		*webhook = updated
	}

	return nil
}

type webhookValidationIssue struct {
	Name    string `json:"name"`
	Path    []any  `json:"path"`
	Details string `json:"details"`
}

// Header shape errors precede definition validation. Both precede a PUT's
// version check, but retaining an omitted value needs a matching version.
func validateWebhookDefinition(stored *cm.WebhookDefinition, fields *cm.WebhookDefinitionData) (*cm.ErrorStatusCode, error) {
	var shape, issues []webhookValidationIssue

	seen := make(map[string]bool, len(fields.Headers))
	missingValue := false

	for index, header := range fields.Headers {
		if _, supplied := header.Value.Get(); !supplied {
			missingValue = true

			if !header.Secret.IsSet() {
				shape = append(shape, webhookValidationIssue{Name: "invalid_union", Path: []any{"headers", index}, Details: fmt.Sprintf(`Invalid input at "headers.%d"`, index)})
			}
		}

		if seen[header.Key] {
			issues = append(issues, webhookValidationIssue{Name: "headers", Path: []any{"key"}, Details: "Header key " + header.Key + " is duplicated"})
		}

		seen[header.Key] = true
	}

	if len(shape) > 0 {
		issues = shape
	} else {
		issues = append(issues, webhookBasicAuthIssues(stored, fields)...)
	}

	if len(issues) > 0 {
		return webhookValidationError(issues)
	}

	if stored == nil && missingValue {
		return NewContentfulManagementErrorStatusCode(http.StatusInternalServerError, "ServerError", new("Failed to create webhook definition/s"), nil), nil
	}

	return nil, nil //nolint:nilnil // The definition passed validation.
}

func webhookValidationError(issues []webhookValidationIssue) (*cm.ErrorStatusCode, error) {
	details, err := json.Marshal(map[string][]webhookValidationIssue{"errors": issues})
	if err != nil {
		return nil, fmt.Errorf("encode webhook validation details: %w", err)
	}

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), details), nil
}

func webhookBasicAuthIssues(stored *cm.WebhookDefinition, fields *cm.WebhookDefinitionData) []webhookValidationIssue {
	username, password := fields.HttpBasicUsername, fields.HttpBasicPassword

	const loneNull = "HTTP Basic Auth username and password must both be provided when setting to null"

	var issues []webhookValidationIssue

	addError := func(field, detail string) {
		issues = append(issues, webhookValidationIssue{Name: field, Path: []any{}, Details: detail})
	}

	switch {
	case username.Or("") != "" && password.Or("") == "":
		addError("httpBasicPassword", "HTTP Basic Auth password is required when username is provided")
	case password.Or("") != "" && username.Or("") == "":
		addError("httpBasicUsername", "HTTP Basic Auth username is required when password is provided")
	case stored != nil && username.IsNull() && password.IsEmpty() && stored.HttpBasicUsername.Or("") != "":
		addError("httpBasicPassword", loneNull)
	case stored != nil && password.IsNull() && username.IsEmpty() && stored.HttpBasicPassword.Or("") != "":
		addError("httpBasicUsername", loneNull)
	}

	if cm.WebhookBasicUsernameTooLong(username.Or("")) {
		addError("httpBasicUsername", "HTTP Basic Auth username must not exceed 50 characters")
	}

	if cm.WebhookBasicPasswordTooLong(password.Or("")) {
		addError("httpBasicPassword", "HTTP Basic Auth password must not exceed 255 characters")
	}

	return issues
}

func missingWebhookHeaderValueError(key string) *cm.ErrorStatusCode {
	var details jx.Encoder
	details.ObjStart()
	details.FieldStart("errors")
	details.Str("Header " + key + " is missing a value and cannot be set")
	details.ObjEnd()

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), details.Bytes())
}

func setWebhookDefinitionFields(webhookDefinition *cm.WebhookDefinition, webhookDefinitionFields cm.WebhookDefinitionData) {
	webhookDefinition.Name = webhookDefinitionFields.Name

	webhookDefinition.URL = webhookDefinitionFields.URL
	if webhookDefinitionFields.HttpBasicUsername.IsSet() {
		webhookDefinition.HttpBasicUsername = webhookDefinitionFields.HttpBasicUsername
	}

	if webhookDefinitionFields.HttpBasicPassword.IsSet() {
		webhookDefinition.HttpBasicPassword = webhookDefinitionFields.HttpBasicPassword
	}

	webhookDefinition.Headers = make(cm.WebhookDefinitionHeaders, len(webhookDefinitionFields.Headers))
	for i, header := range webhookDefinitionFields.Headers {
		stored := cm.WebhookDefinitionHeader{Key: header.Key, Secret: header.Secret}
		if value, ok := header.Value.Get(); ok {
			stored.Value.SetTo(value)
		}

		webhookDefinition.Headers[i] = stored
	}

	webhookDefinition.Topics = webhookDefinitionFields.Topics
	webhookDefinition.Filters = webhookDefinitionFields.Filters
	webhookDefinition.Active = cm.NewOptBool(webhookDefinitionFields.Active.Or(true))

	webhookDefinition.Transformation = webhookDefinitionFields.Transformation
	if webhookDefinition.Transformation.IsNull() {
		webhookDefinition.Transformation.Reset()
	}
}

func redactWebhookDefinitionSecrets(webhookDefinition cm.WebhookDefinition) cm.WebhookDefinition {
	redacted := webhookDefinition
	redacted.HttpBasicPassword.Reset()
	redacted.Headers = make(cm.WebhookDefinitionHeaders, len(webhookDefinition.Headers))
	copy(redacted.Headers, webhookDefinition.Headers)

	for index := range redacted.Headers {
		if redacted.Headers[index].Secret.Or(false) {
			redacted.Headers[index].Value.Reset()
		}
	}

	return redacted
}
