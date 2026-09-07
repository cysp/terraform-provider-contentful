package provider

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func taxonomyCollectionOwnership[T attr.Value](config, plan T, valuePath path.Path, diags *diag.Diagnostics) bool {
	if config.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unknown configuration value", "Taxonomy collection ownership must be known before Contentful can be changed.")

		return false
	}

	if config.IsNull() {
		return false
	}

	diags.Append(rejectUnknownConfigurationOwnedRequestValue(plan, config, valuePath)...)
	taxonomyRejectNullConfigurationOwnedCollection(plan, valuePath, diags)

	return true
}

func taxonomyRejectNullConfigurationOwnedCollection(value interface {
	IsNull() bool
	IsUnknown() bool
}, valuePath path.Path, diags *diag.Diagnostics,
) {
	if !value.IsNull() {
		return
	}

	diags.AddAttributeError(valuePath, "Unavailable planned value", "A configuration-owned taxonomy collection must be non-null before Contentful can be changed.")
}

func taxonomyKnownPlanValue(value interface {
	IsNull() bool
	IsUnknown() bool
},
) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func compareTaxonomyMutationValue(resourceName, attributeName string, valuePath path.Path, planned, remote attr.Value, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A configuration-owned taxonomy value must be known before Contentful can be changed.")

		return false
	}

	if planned.Equal(remote) {
		return true
	}

	diags.AddAttributeError(taxonomyMutationDifferencePath(valuePath, planned, remote), "Unexpected Contentful "+resourceName+" response", fmt.Sprintf("The %s response differed meaningfully from the Terraform plan.", attributeName))

	return false
}

func compareTaxonomyNullableLocalizedStringMutationValue(resourceName, attributeName string, valuePath path.Path, planned, remote types.Map, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A configuration-owned taxonomy value must be known before Contentful can be changed.")

		return false
	}

	if taxonomyNullableLocalizedStringEquivalentAfterMutation(planned, remote) {
		return true
	}

	diags.AddAttributeError(taxonomyMutationDifferencePath(valuePath, planned, remote), "Unexpected Contentful "+resourceName+" response", fmt.Sprintf("The %s response differed meaningfully from the Terraform plan.", attributeName))

	return false
}

func taxonomyMutationIdentityConsistency(resourceName string, valuePath path.Path, planned, remote types.String, diags *diag.Diagnostics) bool {
	if !planned.IsNull() && !planned.IsUnknown() && !planned.Equal(remote) {
		diags.AddAttributeError(valuePath, "Unexpected Contentful "+resourceName+" response", "The response identity differed from the requested taxonomy endpoint.")

		return true
	}

	return false
}

func taxonomyMutationDifferencePath(valuePath path.Path, planned, remote attr.Value) path.Path {
	plannedList, plannedIsList := planned.(types.List)

	remoteList, remoteIsList := remote.(types.List)
	if plannedIsList && remoteIsList && !plannedList.IsNull() && !plannedList.IsUnknown() && !remoteList.IsNull() && !remoteList.IsUnknown() {
		return taxonomyListDifferencePath(valuePath, plannedList, remoteList)
	}

	plannedMap, plannedIsMap := planned.(types.Map)

	remoteMap, remoteIsMap := remote.(types.Map)
	if plannedIsMap && remoteIsMap && !plannedMap.IsNull() && !plannedMap.IsUnknown() && !remoteMap.IsNull() && !remoteMap.IsUnknown() {
		keys := make([]string, 0, len(plannedMap.Elements())+len(remoteMap.Elements()))
		for key := range plannedMap.Elements() {
			keys = append(keys, key)
		}

		for key := range remoteMap.Elements() {
			if _, ok := plannedMap.Elements()[key]; !ok {
				keys = append(keys, key)
			}
		}

		sort.Strings(keys)

		for _, key := range keys {
			plannedValue, plannedExists := plannedMap.Elements()[key]

			remoteValue, remoteExists := remoteMap.Elements()[key]
			if !plannedExists || !remoteExists || !plannedValue.Equal(remoteValue) {
				return valuePath.AtMapKey(key)
			}
		}
	}

	return valuePath
}

func taxonomyListDifferencePath(valuePath path.Path, planned, remote types.List) path.Path {
	plannedElements, remoteElements := planned.Elements(), remote.Elements()

	limit := min(len(plannedElements), len(remoteElements))
	for index := range limit {
		if !plannedElements[index].Equal(remoteElements[index]) {
			return valuePath.AtListIndex(index)
		}
	}

	return valuePath.AtListIndex(limit)
}
