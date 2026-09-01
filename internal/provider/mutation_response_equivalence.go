package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func unorderedElementsEquivalent[T any](planned, response []T, elementsEquivalent func(T, T) (bool, diag.Diagnostics)) (bool, diag.Diagnostics) {
	if len(planned) != len(response) {
		return false, nil
	}

	var comparisonDiags diag.Diagnostics

	matched := make([]bool, len(response))

	for _, plannedElement := range planned {
		found := false

		for responseIndex, responseElement := range response {
			if matched[responseIndex] {
				continue
			}

			equivalent, elementDiags := elementsEquivalent(plannedElement, responseElement)
			comparisonDiags.Append(elementDiags...)

			if elementDiags.HasError() {
				return false, comparisonDiags
			}

			if !equivalent {
				continue
			}

			matched[responseIndex] = true
			found = true

			break
		}

		if !found {
			return false, comparisonDiags
		}
	}

	return true, comparisonDiags
}

// Matching each occurrence separately accepts order changes without discarding
// duplicate multiplicity.
func unorderedStringListsEquivalent(planned, response TypedList[types.String]) bool {
	if planned.Equal(response) {
		return true
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false
	}

	equivalent, _ := unorderedElementsEquivalent(planned.Elements(), response.Elements(), func(plannedElement, responseElement types.String) (bool, diag.Diagnostics) {
		return plannedElement.Equal(responseElement), nil
	})

	return equivalent
}

func webhookFiltersEquivalent(planned, response TypedList[TypedObject[WebhookFilterValue]]) bool {
	if planned.Equal(response) {
		return true
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false
	}

	equivalent, _ := unorderedElementsEquivalent(planned.Elements(), response.Elements(), func(plannedElement, responseElement TypedObject[WebhookFilterValue]) (bool, diag.Diagnostics) {
		return webhookFilterEquivalent(plannedElement, responseElement), nil
	})

	return equivalent
}

func webhookFilterEquivalent(planned, response TypedObject[WebhookFilterValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return webhookFilterNotEquivalent(plannedValue.Not, responseValue.Not) &&
		plannedValue.Equals.Equal(responseValue.Equals) &&
		webhookFilterInEquivalent(plannedValue.In, responseValue.In) &&
		plannedValue.Regexp.Equal(responseValue.Regexp)
}

func webhookFilterNotEquivalent(planned, response TypedObject[WebhookFilterNotValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return plannedValue.Equals.Equal(responseValue.Equals) &&
		webhookFilterInEquivalent(plannedValue.In, responseValue.In) &&
		plannedValue.Regexp.Equal(responseValue.Regexp)
}

func webhookFilterInEquivalent(planned, response TypedObject[WebhookFilterInValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return plannedValue.Doc.Equal(responseValue.Doc) && unorderedStringListsEquivalent(plannedValue.Values, responseValue.Values)
}

func rolePermissionsEquivalent(planned, response TypedMap[TypedList[types.String]]) bool {
	if planned.Equal(response) {
		return true
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false
	}

	plannedElements := planned.Elements()

	responseElements := response.Elements()
	if len(plannedElements) != len(responseElements) {
		return false
	}

	for name, plannedActions := range plannedElements {
		responseActions, ok := responseElements[name]
		if !ok || !unorderedStringListsEquivalent(plannedActions, responseActions) {
			return false
		}
	}

	return true
}

func rolePoliciesEquivalent(ctx context.Context, planned, response TypedList[TypedObject[RolePolicyValue]]) (bool, diag.Diagnostics) {
	if planned.Equal(response) {
		return true, nil
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false, nil
	}

	return unorderedElementsEquivalent(planned.Elements(), response.Elements(), func(plannedElement, responseElement TypedObject[RolePolicyValue]) (bool, diag.Diagnostics) {
		return rolePolicyEquivalent(ctx, plannedElement, responseElement)
	})
}

func rolePolicyEquivalent(ctx context.Context, planned, response TypedObject[RolePolicyValue]) (bool, diag.Diagnostics) {
	if planned.Equal(response) {
		return true, nil
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown || !plannedValue.Effect.Equal(responseValue.Effect) ||
		!unorderedStringListsEquivalent(plannedValue.Actions, responseValue.Actions) {
		return false, nil
	}

	plannedConstraint := plannedValue.Constraint

	responseConstraint := responseValue.Constraint
	if plannedConstraint.Equal(responseConstraint) {
		return true, nil
	}

	if plannedConstraint.IsNull() || plannedConstraint.IsUnknown() || responseConstraint.IsNull() || responseConstraint.IsUnknown() {
		return false, nil
	}

	return plannedConstraint.StringSemanticEquals(ctx, responseConstraint)
}
