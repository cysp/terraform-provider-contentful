package cmtesting_test

import (
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationMapZeroValue(t *testing.T) {
	t.Parallel()

	var store cmt.OrganizationMap[int]

	assert.Zero(t, store.Get("missing-organization", "missing-object"))
	assert.Empty(t, store.Values("missing-organization"))
	assert.NotPanics(t, func() {
		store.Delete("missing-organization", "missing-object")
	})

	store.Set("organization", "object", 42)

	assert.Equal(t, 42, store.Get("organization", "object"))
	assert.Equal(t, []int{42}, store.Values("organization"))
}

func TestOrganizationMapScopesValues(t *testing.T) {
	t.Parallel()

	var store cmt.OrganizationMap[string]
	store.Set("organization-one", "first", "one")
	store.Set("organization-one", "second", "two")
	store.Set("organization-two", "first", "three")

	assert.ElementsMatch(t, []string{"one", "two"}, store.Values("organization-one"))
	assert.Equal(t, "three", store.Get("organization-two", "first"))
	assert.Empty(t, store.Get("missing-organization", "first"))
	assert.Empty(t, store.Get("organization-one", "missing-object"))
	assert.Empty(t, store.Values("missing-organization"))
}
