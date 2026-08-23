package contentfulmanagement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpaceEnvironmentMapZeroValueGetDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store SpaceEnvironmentMap[string]

	_ = store.Get("space", "environment", "object")

	assert.Nil(t, store.m)
}

func TestSpaceEnvironmentMapZeroValueListDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store SpaceEnvironmentMap[string]

	_ = store.List("space", "environment")

	assert.Nil(t, store.m)
}

func TestSpaceEnvironmentMapZeroValueDeleteDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store SpaceEnvironmentMap[string]

	store.Delete("space", "environment", "object")

	assert.Nil(t, store.m)
}
