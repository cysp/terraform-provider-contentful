package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mutationResponseReconciler verifies every known API-backed Plan value.
// Callers publish their candidate Plan representation only when it reports no
// errors.
type mutationResponseReconciler struct {
	resourceName string
	diagnostics  diag.Diagnostics
}

func (r *mutationResponseReconciler) compareSemantic(valuePath path.Path, subject, mismatchSummary string, planUnknown bool, projectionDiags diag.Diagnostics, equivalent func() (bool, diag.Diagnostics)) bool {
	if planUnknown {
		return false
	}

	if len(projectionDiags) != 0 {
		r.diagnostics.AddAttributeError(valuePath, "Provider cannot fully represent "+subject, "Contentful accepted the request, but the returned value contains data this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the "+r.resourceName+" in Contentful before applying again.")

		return false
	}

	valuesEquivalent, comparisonDiags := equivalent()

	if comparisonDiags.HasError() {
		r.diagnostics.AddAttributeError(valuePath, "Provider could not compare "+subject, "Contentful accepted the request, but the provider could not compare the returned value with the value Terraform applied. Terraform retained the Contentful response. Review the "+r.resourceName+" in Contentful before applying again.")

		return false
	}

	return r.acceptEquivalent(valuePath, mismatchSummary, valuesEquivalent)
}

func (r *mutationResponseReconciler) compareExact(valuePath path.Path, mismatchSummary string, plan, response attr.Value) bool {
	if plan.IsUnknown() {
		return false
	}

	return r.acceptEquivalent(valuePath, mismatchSummary, plan.Equal(response))
}

func (r *mutationResponseReconciler) acceptEquivalent(valuePath path.Path, mismatchSummary string, equivalent bool) bool {
	if !equivalent {
		r.diagnostics.AddAttributeError(valuePath, mismatchSummary, "Contentful accepted the request but returned a value that differs from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the "+r.resourceName+" and configuration before applying again.")
	}

	return equivalent
}

func (r *mutationResponseReconciler) pinIdentity(valuePath path.Path, plan, response types.String, target *types.String) {
	if plan.IsNull() || plan.IsUnknown() {
		return
	}

	if !plan.Equal(response) {
		r.diagnostics.AddAttributeError(valuePath, "Contentful returned a different "+r.resourceName+" identity", "Contentful accepted the request but returned an identity that differs from the requested Terraform resource. Terraform retained the requested identity as the resource target and the remaining returned values as recovery state. Review the "+r.resourceName+" in Contentful before applying again.")
	}

	*target = plan
}

func normalizedJSONEquivalent(ctx context.Context, planned, response jsontypes.Normalized) (bool, diag.Diagnostics) {
	if planned.Equal(response) {
		return true, nil
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false, nil
	}

	return planned.StringSemanticEquals(ctx, response)
}

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

func orderedObjectListsEquivalent[T any](planned, response TypedList[TypedObject[T]], elementsEquivalent func(T, T) (bool, diag.Diagnostics)) (bool, diag.Diagnostics) {
	if planned.Equal(response) {
		return true, nil
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false, nil
	}

	plannedElements := planned.Elements()

	responseElements := response.Elements()
	if len(plannedElements) != len(responseElements) {
		return false, nil
	}

	for index := range plannedElements {
		plannedElement := plannedElements[index]
		responseElement := responseElements[index]

		plannedValue, plannedKnown := plannedElement.GetValue()

		responseValue, responseKnown := responseElement.GetValue()
		if !plannedKnown || !responseKnown {
			return false, nil
		}

		equivalent, elementDiags := elementsEquivalent(plannedValue, responseValue)
		if elementDiags.HasError() || !equivalent {
			return false, elementDiags
		}
	}

	return true, nil
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
