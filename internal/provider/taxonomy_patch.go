package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// taxonomyPatch compares encoded request members and emits deterministic
// top-level JSON Patch replacements. The generated client omits optional
// members when their option is unset, so response-owned members can be left
// unset by the prepared mutation.
func taxonomyPatch(current, desired any) (cm.TaxonomyPatch, error) {
	currentFields, err := taxonomyRequestFields(current, "current")
	if err != nil {
		return nil, err
	}

	desiredFields, err := taxonomyRequestFields(desired, "desired")
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(desiredFields))
	for key, desiredValue := range desiredFields {
		currentValue, ok := currentFields[key]
		if !ok {
			keys = append(keys, key)

			continue
		}

		equal, err := taxonomyJSONEqual(currentValue, desiredValue)
		if err != nil {
			return nil, fmt.Errorf("compare taxonomy request member %q: %w", key, err)
		}

		if equal {
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	result := make(cm.TaxonomyPatch, 0, len(keys))
	for _, key := range keys {
		result = append(result, cm.TaxonomyPatchItem{
			Op:    cm.TaxonomyPatchItemOpAdd,
			Path:  "/" + key,
			Value: jx.Raw(desiredFields[key]),
		})
	}

	return result, nil
}

func taxonomyJSONEqual(left, right []byte) (bool, error) {
	var leftValue, rightValue any

	err := json.Unmarshal(left, &leftValue)
	if err != nil {
		return false, fmt.Errorf("decode current JSON: %w", err)
	}

	err = json.Unmarshal(right, &rightValue)
	if err != nil {
		return false, fmt.Errorf("decode desired JSON: %w", err)
	}

	return reflect.DeepEqual(leftValue, rightValue), nil
}

func taxonomyRequestFields(value any, label string) (map[string]json.RawMessage, error) {
	switch request := value.(type) {
	case cm.TaxonomyConceptRequest:
		value = &request
	case cm.TaxonomyConceptSchemeRequest:
		value = &request
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode %s taxonomy request: %w", label, err)
	}

	fields := map[string]json.RawMessage{}

	err = json.Unmarshal(encoded, &fields)
	if err != nil {
		return nil, fmt.Errorf("decode %s taxonomy request: %w", label, err)
	}

	return fields, nil
}

func taxonomyPatchErrorDiagnostics(resourceName string, err error) diag.Diagnostics {
	if err == nil {
		return nil
	}

	diags := diag.Diagnostics{}
	diags.AddError("Failed to build "+resourceName+" update", fmt.Sprintf("Build the state-to-plan taxonomy patch: %v", err))

	return diags
}
