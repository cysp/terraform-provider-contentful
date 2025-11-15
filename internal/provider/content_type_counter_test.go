package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestContentfulContentTypeCounterUnknown(t *testing.T) {
	t.Parallel()

	counter := ContentfulContentTypeCounter{}

	assert.Equal(t, 0, counter.Get("non-existent"))
}

func TestContentfulContentTypeCounterSingleThreaded(t *testing.T) {
	t.Parallel()

	counter := ContentfulContentTypeCounter{}
	contentTypeID := "test-id"

	counter.Increment(contentTypeID)

	assert.Equal(t, 1, counter.Get(contentTypeID))

	counter.Increment(contentTypeID)

	assert.Equal(t, 2, counter.Get(contentTypeID))

	counter.Reset(contentTypeID)

	assert.Equal(t, 0, counter.Get(contentTypeID))
}

func TestContentfulContentTypeCounterConcurrent(t *testing.T) {
	t.Parallel()

	counter := ContentfulContentTypeCounter{}
	contentTypeID := "test-id"

	var errGroup errgroup.Group

	numGoroutines := 100
	incrementsPerGoroutine := 1000

	for range numGoroutines {
		errGroup.Go(func() error {
			for range incrementsPerGoroutine {
				counter.Increment(contentTypeID)
			}

			return nil
		})
	}

	err := errGroup.Wait()
	if err != nil {
		t.Fatal(err)
	}

	expected := numGoroutines * incrementsPerGoroutine

	assert.Equal(t, expected, counter.Get(contentTypeID))

	counter.Reset(contentTypeID)

	assert.Equal(t, 0, counter.Get(contentTypeID))
}
