package cmtesting

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestAppKeyStoreZeroValue(t *testing.T) {
	t.Parallel()

	var store appKeyStore

	assert.False(t, store.Contains("missing-key"))
	_, exists := store.Get("missing-organization", "missing-app-definition", "missing-key")
	assert.False(t, exists)
	assert.Empty(t, store.List("missing-organization", "missing-app-definition"))
	assert.NotPanics(t, func() {
		store.Delete("missing-organization", "missing-app-definition", "missing-key")
	})

	store.Set("organization", "app-definition", cm.AppKey{Sys: cm.AppKeySys{ID: "key"}})

	appKey, exists := store.Get("organization", "app-definition", "key")
	assert.True(t, exists)
	assert.Equal(t, "key", appKey.Sys.ID)
	assert.True(t, store.Contains("key"))
}

func TestAppKeyStoreZeroValueLookupsDoNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var containsStore appKeyStore

	_ = containsStore.Contains("key")
	assert.Nil(t, containsStore.records)

	var getStore appKeyStore

	_, _ = getStore.Get("organization", "app-definition", "key")
	assert.Nil(t, getStore.records)

	var listStore appKeyStore

	_ = listStore.List("organization", "app-definition")
	assert.Nil(t, listStore.records)
}

func TestAppKeyStoreZeroValueDeleteDoesNotInitializeStorage(t *testing.T) {
	t.Parallel()

	var store appKeyStore

	store.Delete("organization", "app-definition", "key")

	assert.Nil(t, store.records)
}

func TestAppKeyStoreScopesValues(t *testing.T) {
	t.Parallel()

	var store appKeyStore
	store.Set("organization-one", "app-definition-one", cm.AppKey{Sys: cm.AppKeySys{ID: "first"}})
	store.Set("organization-one", "app-definition-two", cm.AppKey{Sys: cm.AppKeySys{ID: "second"}})
	store.Set("organization-two", "app-definition-one", cm.AppKey{Sys: cm.AppKeySys{ID: "third"}})

	assert.Equal(t, []cm.AppKey{{Sys: cm.AppKeySys{ID: "first"}}}, store.List("organization-one", "app-definition-one"))
	assert.Equal(t, []cm.AppKey{{Sys: cm.AppKeySys{ID: "second"}}}, store.List("organization-one", "app-definition-two"))
	assert.Equal(t, []cm.AppKey{{Sys: cm.AppKeySys{ID: "third"}}}, store.List("organization-two", "app-definition-one"))
	assert.Empty(t, store.List("missing-organization", "app-definition-one"))
	assert.Empty(t, store.List("organization-one", "missing-app-definition"))
	_, exists := store.Get("organization-one", "app-definition-one", "missing-key")
	assert.False(t, exists)
}
