package cmtesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"unicode/utf8"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

// The fake counts Unicode code points. The supplied probes establish the ASCII
// boundary and acceptance of BMP/astral examples, but not combining-sequence rules.
const livePreviewVariableMaxLength = 50000

type livePreviewVariableValidationError struct {
	Name        string         `json:"name"`
	Type        string         `json:"type,omitempty"`
	Value       any            `json:"value"`
	Details     string         `json:"details"`
	Path        []string       `json:"path,omitempty"`
	I18nContext map[string]any `json:"i18nContext"`
}

func validateLivePreviewVariablesData(raw []byte) ([]byte, *cm.LivePreviewVariablesErrorStatusCode, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any

	err := decoder.Decode(&value)
	if err != nil {
		return nil, nil, fmt.Errorf("decode live preview variables: %w", err)
	}

	variables, ok := livePreviewVariablesObject(value)

	var failures []livePreviewVariableValidationError
	if !ok {
		failures = append(failures, livePreviewVariableTypeError(value, "Object", nil))
	}

	for _, name := range slices.Sorted(maps.Keys(variables)) {
		if name == "__proto__" {
			return nil, &cm.LivePreviewVariablesErrorStatusCode{StatusCode: http.StatusBadRequest, Response: cm.NewLivePreviewVariablesServiceErrorLivePreviewVariablesError(cm.LivePreviewVariablesServiceError{StatusCode: http.StatusBadRequest, Error: "Bad Request", Message: "Invalid request payload JSON format"})}, nil
		}

		variable := variables[name]

		localized, isObject := livePreviewVariablesObject(variable)
		if !isObject {
			failures = append(failures, validateLivePreviewVariableText(variable, []string{name})...)

			continue
		}

		variables[name] = localized
		for _, locale := range slices.Sorted(maps.Keys(localized)) {
			// As in the taxonomy fake, en-US is the configured locale fixture.
			if locale != "en-US" {
				failures = append(failures, livePreviewVariableValidationError{Name: "unknown", Value: localized[locale], Details: fmt.Sprintf("The property %q is not allowed here.", locale), Path: []string{name, locale}, I18nContext: map[string]any{"code": "CmaError.Field.Validation.UnknownProperty", "parameters": map[string]any{"propertyName": map[string]any{"type": "string", "value": locale}}}})

				continue
			}

			failures = append(failures, validateLivePreviewVariableText(localized[locale], []string{name, locale})...)
		}
	}

	if len(failures) > 0 {
		details, err := json.Marshal(struct {
			Errors []livePreviewVariableValidationError `json:"errors"`
		}{Errors: failures})
		if err != nil {
			return nil, nil, fmt.Errorf("encode live preview variables validation: %w", err)
		}

		return nil, &cm.LivePreviewVariablesErrorStatusCode{StatusCode: http.StatusUnprocessableEntity, Response: cm.NewErrorLivePreviewVariablesError(NewContentfulManagementError("ValidationFailed", new("Validation error"), details))}, nil
	}

	normalized, err := json.Marshal(variables)
	if err != nil {
		return nil, nil, fmt.Errorf("encode live preview variables: %w", err)
	}

	return normalized, nil, nil
}

func livePreviewVariablesObject(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case []any:
		result := make(map[string]any, len(typed))
		for index, item := range typed {
			result[strconv.Itoa(index)] = item
		}

		return result, true
	default:
		return nil, false
	}
}

func validateLivePreviewVariableText(value any, location []string) []livePreviewVariableValidationError {
	if value == nil {
		return nil
	}

	text, ok := value.(string)
	if !ok {
		return []livePreviewVariableValidationError{livePreviewVariableTypeError(value, "Text", location)}
	}

	if utf8.RuneCountInString(text) <= livePreviewVariableMaxLength {
		return nil
	}

	return []livePreviewVariableValidationError{{Name: "type", Type: "Text", Value: value, Details: fmt.Sprintf("Maximum Text length is %d characters", livePreviewVariableMaxLength), Path: location, I18nContext: map[string]any{"code": "CmaError.Field.Validation.InvalidTextLength", "parameters": map[string]any{"maxLength": map[string]any{"type": "number", "value": livePreviewVariableMaxLength}}}}}
}

func livePreviewVariableTypeError(value any, expected string, location []string) livePreviewVariableValidationError {
	return livePreviewVariableValidationError{Name: "type", Type: expected, Value: value, Details: "The type of \"value\" is incorrect, expected type: " + expected, Path: location, I18nContext: map[string]any{"code": "CmaError.Field.Validation.IncorrectType", "parameters": map[string]any{"schemaType": map[string]any{"type": "string", "value": expected}}}}
}
