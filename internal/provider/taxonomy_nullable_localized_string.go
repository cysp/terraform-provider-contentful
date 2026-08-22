package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// taxonomyNullableLocalizedStringEquivalentAfterMutation implements the narrow
// CMA canonicalization observed for the Optional localized string maps on
// taxonomy concepts and concept schemes. Contentful persists a requested known
// empty map as null. It is directional: a planned null is not equivalent to a
// remote empty map, and no other map difference is ignored.
func taxonomyNullableLocalizedStringEquivalentAfterMutation(planned, remote types.Map) bool {
	if planned.IsUnknown() {
		return false
	}

	return planned.Equal(remote) || (!planned.IsNull() && len(planned.Elements()) == 0 && remote.IsNull())
}

// taxonomyNullableLocalizedStringAfterRefresh preserves the representation
// Terraform previously accepted for the same CMA empty-map canonicalization.
// A first Read (including import) and every meaningful remote difference take
// the remote response.
func taxonomyNullableLocalizedStringAfterRefresh(prior, remote types.Map) types.Map {
	if taxonomyNullableLocalizedStringEquivalentAfterMutation(prior, remote) {
		return prior
	}

	return remote
}
