package cmtesting

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID string, appInstallationFields cm.AppInstallationData) cm.AppInstallation {
	appInstallation := cm.AppInstallation{
		Sys: cm.NewAppInstallationSys(spaceID, environmentID, appDefinitionID),
	}

	UpdateAppInstallationFromFields(&appInstallation, appInstallationFields)

	return appInstallation
}

func UpdateAppInstallationFromFields(appInstallation *cm.AppInstallation, appInstallationFields cm.AppInstallationData) {
	appInstallation.Parameters = slices.Clone(appInstallationFields.Parameters)
}

// These rules model the PAT observations in
// docs/research/app-framework/secret-installation-parameters.md. They do not model
// App Identity authorization, defaults, or every parameter declaration type.
func validateAppInstallationParameters(req cm.AppInstallationData, declarations []cm.AppDefinitionParameter) (map[string]json.RawMessage, map[string]any) {
	if len(req.Parameters) == 0 {
		return nil, nil
	}

	var values map[string]json.RawMessage
	if json.Unmarshal(req.Parameters, &values) != nil || values == nil {
		return nil, map[string]any{
			"name": "type", "details": `The type of "parameters" is incorrect, expected type: object.`, "path": []string{"parameters"},
		}
	}

	for _, declaration := range declarations {
		value, present := values[declaration.ID]
		if !present {
			if declaration.Required.Or(false) {
				return nil, map[string]any{
					"name": "required", "details": fmt.Sprintf("The property %q is required here", declaration.ID), "path": []string{"parameters", declaration.ID},
				}
			}

			continue
		}

		if declaration.Type == "Secret" || declaration.Type == "Symbol" {
			var text *string
			if json.Unmarshal(value, &text) != nil || text == nil {
				return nil, map[string]any{
					"name": "type", "details": fmt.Sprintf("The type of %q is incorrect, expected type: String", declaration.ID), "path": []string{"parameters", declaration.ID}, "type": "String", "value": value,
				}
			}
		}
	}

	if len(declarations) != 0 {
		for key := range values {
			if !slices.ContainsFunc(declarations, func(declaration cm.AppDefinitionParameter) bool { return declaration.ID == key }) {
				return nil, map[string]any{
					"name": "unexpected", "details": fmt.Sprintf("The property %q is not expected", key), "path": []string{"parameters", key},
				}
			}
		}
	}

	return values, nil
}

func appInstallationParameters(req cm.AppInstallationData, definition cm.AppDefinition, previous *cm.AppInstallation) ([]byte, *cm.ErrorStatusCode, error) {
	values, failure := validateAppInstallationParameters(req, definition.Parameters.Value.Installation)
	if failure != nil {
		details, err := json.Marshal(map[string]any{"errors": []any{failure}})
		if err != nil {
			return nil, nil, fmt.Errorf("encode installation validation: %w", err)
		}

		return nil, NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), details), nil
	}

	if values == nil {
		return nil, nil, nil
	}

	var stored map[string]json.RawMessage
	if previous != nil {
		_ = json.Unmarshal(previous.Parameters, &stored)
	}

	for _, declaration := range definition.Parameters.Value.Installation {
		if declaration.Type != "Secret" {
			continue
		}

		var value, prior *string
		if json.Unmarshal(values[declaration.ID], &value) == nil && value != nil && json.Unmarshal(stored[declaration.ID], &prior) == nil && prior != nil && *value == appInstallationSecretMask(*prior) {
			values[declaration.ID] = stored[declaration.ID]
		}
	}

	parameters, err := json.Marshal(values)
	if err != nil {
		return nil, nil, fmt.Errorf("encode installation parameters: %w", err)
	}

	return parameters, nil, nil
}

func appInstallationResponse(stored *cm.AppInstallation, definition *cm.AppDefinition) (*cm.AppInstallation, error) {
	response := *stored

	response.Parameters = slices.Clone(stored.Parameters)
	if definition == nil || len(stored.Parameters) == 0 {
		return &response, nil
	}

	var values map[string]json.RawMessage

	err := json.Unmarshal(stored.Parameters, &values)
	if err != nil {
		return nil, fmt.Errorf("decode stored installation parameters: %w", err)
	}

	for _, declaration := range definition.Parameters.Value.Installation {
		if declaration.Type != "Secret" {
			continue
		}

		var value *string
		if json.Unmarshal(values[declaration.ID], &value) == nil && value != nil {
			values[declaration.ID] = json.RawMessage(strconv.Quote(appInstallationSecretMask(*value)))
		}
	}

	parameters, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("encode installation response: %w", err)
	}

	response.Parameters = parameters

	return &response, nil
}

func appInstallationSecretMask(value string) string {
	return strings.Repeat("*", len(utf16.Encode([]rune(value))))
}
