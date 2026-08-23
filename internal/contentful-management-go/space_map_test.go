package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestSpaceMapZeroValue(t *testing.T) {
	t.Parallel()

	var store cm.SpaceMap[int]

	assert.Zero(t, store.Get("missing-space", "missing-object"))
	assert.NotPanics(t, func() {
		store.Delete("missing-space", "missing-object")
	})

	store.Set("space", "object", 42)

	assert.Equal(t, 42, store.Get("space", "object"))
}

func TestSpaceMapScopesValues(t *testing.T) {
	t.Parallel()

	var store cm.SpaceMap[string]
	store.Set("space-one", "first", "one")
	store.Set("space-one", "second", "two")
	store.Set("space-two", "first", "three")

	assert.Equal(t, "one", store.Get("space-one", "first"))
	assert.Equal(t, "two", store.Get("space-one", "second"))
	assert.Equal(t, "three", store.Get("space-two", "first"))
	assert.Empty(t, store.Get("missing-space", "first"))
	assert.Empty(t, store.Get("space-one", "missing-object"))
}
