package cmtesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationMapZeroValueGetDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store OrganizationMap[string]

	_ = store.Get("organization", "object")

	assert.Nil(t, store.m)
}

func TestOrganizationMapZeroValueValuesDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store OrganizationMap[string]

	_ = store.Values("organization")

	assert.Nil(t, store.m)
}

func TestOrganizationMapZeroValueDeleteDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store OrganizationMap[string]

	store.Delete("organization", "object")

	assert.Nil(t, store.m)
}
