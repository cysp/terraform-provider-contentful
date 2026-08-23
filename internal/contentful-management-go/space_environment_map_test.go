package contentfulmanagement_test

import (
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestSpaceEnvironmentMapZeroValue(t *testing.T) {
	t.Parallel()

	var store cm.SpaceEnvironmentMap[int]

	assert.Zero(t, store.Get("missing-space", "missing-environment", "missing-object"))
	assert.Empty(t, store.List("missing-space", "missing-environment"))
	assert.NotPanics(t, func() {
		store.Delete("missing-space", "missing-environment", "missing-object")
	})

	store.Set("space", "environment", "object", 42)

	assert.Equal(t, 42, store.Get("space", "environment", "object"))
	assert.Equal(t, []int{42}, store.List("space", "environment"))
}

func TestSpaceEnvironmentMapScopesValues(t *testing.T) {
	t.Parallel()

	var store cm.SpaceEnvironmentMap[string]
	store.Set("space-one", "environment-one", "first", "one")
	store.Set("space-one", "environment-one", "second", "two")
	store.Set("space-one", "environment-two", "first", "three")
	store.Set("space-two", "environment-one", "first", "four")

	assert.ElementsMatch(t, []string{"one", "two"}, store.List("space-one", "environment-one"))
	assert.Equal(t, "three", store.Get("space-one", "environment-two", "first"))
	assert.Equal(t, "four", store.Get("space-two", "environment-one", "first"))
	assert.Empty(t, store.Get("missing-space", "environment-one", "first"))
	assert.Empty(t, store.Get("space-one", "missing-environment", "first"))
	assert.Empty(t, store.Get("space-one", "environment-one", "missing-object"))
	assert.Empty(t, store.List("missing-space", "environment-one"))
	assert.Empty(t, store.List("space-one", "missing-environment"))
}

func TestSpaceEnvironmentMapConcurrentZeroValueReads(t *testing.T) {
	t.Parallel()

	var (
		store      cm.SpaceEnvironmentMap[int]
		waitGroup  sync.WaitGroup
		unexpected atomic.Int64
	)

	start := make(chan struct{})

	for range 100 {
		waitGroup.Go(func() {
			<-start

			for range 1000 {
				if store.Get("space", "environment", "object") != 0 {
					unexpected.Add(1)
				}

				if len(store.List("space", "environment")) != 0 {
					unexpected.Add(1)
				}
			}
		})
	}

	close(start)
	waitGroup.Wait()

	assert.Zero(t, unexpected.Load())
}
