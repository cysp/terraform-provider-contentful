package provider

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// taxonomyLabelMapsEquivalentAt implements the deliberately narrow taxonomy
// label equivalence rule. The remote side may contain one or more preferred
// locales as known empty lists when those locales were absent from the owned
// configured map. Every other key and every value remains exact.
func taxonomyLabelMapsEquivalentAt(planned, remote, prefLabel types.Map, valuePath path.Path) (bool, path.Path) {
	if planned.IsNull() || planned.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return planned.Equal(remote), valuePath
	}

	plannedElements := planned.Elements()
	remoteElements := remote.Elements()
	keys := make([]string, 0, len(plannedElements)+len(remoteElements))

	seen := make(map[string]struct{}, len(plannedElements)+len(remoteElements))
	for locale := range plannedElements {
		seen[locale] = struct{}{}
		keys = append(keys, locale)
	}

	for locale := range remoteElements {
		if _, ok := seen[locale]; !ok {
			keys = append(keys, locale)
		}
	}

	sort.Strings(keys)

	for _, locale := range keys {
		plannedValue, plannedExists := plannedElements[locale]

		remoteValue, remoteExists := remoteElements[locale]
		if !plannedExists {
			if preferredLocale(prefLabel, locale) && taxonomyLabelListIsKnownEmpty(remoteValue) {
				continue
			}

			return false, valuePath.AtMapKey(locale)
		}

		if !remoteExists || !plannedValue.Equal(remoteValue) {
			return false, taxonomyLabelValueDifferencePath(valuePath.AtMapKey(locale), plannedValue, remoteValue)
		}
	}

	return true, valuePath
}

func preferredLocale(prefLabel types.Map, locale string) bool {
	if prefLabel.IsNull() || prefLabel.IsUnknown() {
		return false
	}

	_, ok := prefLabel.Elements()[locale]

	return ok
}

func taxonomyLabelValueDifferencePath(localePath path.Path, planned, remote attr.Value) path.Path {
	plannedList, plannedOK := planned.(types.List)

	remoteList, remoteOK := remote.(types.List)
	if !plannedOK || !remoteOK || plannedList.IsNull() || plannedList.IsUnknown() || remoteList.IsNull() || remoteList.IsUnknown() {
		return localePath
	}

	return taxonomyListDifferencePath(localePath, plannedList, remoteList)
}

func taxonomyLabelListIsKnownEmpty(value attr.Value) bool {
	list, ok := value.(types.List)

	return ok && !list.IsNull() && !list.IsUnknown() && len(list.Elements()) == 0
}

// taxonomyLabelMapAfterRefresh stabilizes only a response-added preferred
// locale represented as a known empty list. It never restores a configured
// locale omitted by the response.
func taxonomyLabelMapAfterRefresh(prior, remote, prefLabel types.Map) types.Map {
	if prior.IsNull() || prior.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return remote
	}

	projected := make(map[string]attr.Value, len(remote.Elements()))
	for locale, value := range remote.Elements() {
		projected[locale] = value
		if _, priorExists := prior.Elements()[locale]; priorExists {
			continue
		}

		if preferredLocale(prefLabel, locale) && taxonomyLabelListIsKnownEmpty(value) {
			delete(projected, locale)
		}
	}

	return types.MapValueMust(types.ListType{ElemType: types.StringType}, projected)
}
