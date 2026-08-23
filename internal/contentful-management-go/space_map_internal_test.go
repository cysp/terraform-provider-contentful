package contentfulmanagement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpaceMapZeroValueGetDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store SpaceMap[string]

	_ = store.Get("space", "object")

	assert.Nil(t, store.m)
}

func TestSpaceMapZeroValueDeleteDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store SpaceMap[string]

	store.Delete("space", "object")

	assert.Nil(t, store.m)
}
