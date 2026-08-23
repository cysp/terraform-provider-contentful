package provider_test

import (
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestContentfulContentTypeCounterUnknown(t *testing.T) {
	t.Parallel()

	counter := ContentfulContentTypeCounter{}

	assert.Equal(t, 0, counter.Get("non-existent", "non-existent", "non-existent"))
}

func TestContentfulContentTypeCounterSingleThreaded(t *testing.T) {
	t.Parallel()

	counter := ContentfulContentTypeCounter{}
	contentTypeID := "test-id"

	counter.Increment("spaceID", "environmentID", contentTypeID)

	assert.Equal(t, 1, counter.Get("spaceID", "environmentID", contentTypeID))

	counter.Increment("spaceID", "environmentID", contentTypeID)
	assert.Equal(t, 2, counter.Get("spaceID", "environmentID", contentTypeID))

	counter.Reset("spaceID", "environmentID", contentTypeID)

	assert.Equal(t, 0, counter.Get("spaceID", "environmentID", contentTypeID))
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
				counter.Increment("spaceID", "environmentID", contentTypeID)
			}

			return nil
		})
	}

	err := errGroup.Wait()
	if err != nil {
		t.Fatal(err)
	}

	expected := numGoroutines * incrementsPerGoroutine

	assert.Equal(t, expected, counter.Get("spaceID", "environmentID", contentTypeID))

	counter.Reset("spaceID", "environmentID", contentTypeID)

	assert.Equal(t, 0, counter.Get("spaceID", "environmentID", contentTypeID))
}

func TestContentfulContentTypeCounterConcurrentUnknownReads(t *testing.T) {
	t.Parallel()

	var (
		waitGroup  sync.WaitGroup
		unexpected atomic.Int64
	)

	counter := ContentfulContentTypeCounter{}
	start := make(chan struct{})

	for range 100 {
		waitGroup.Go(func() {
			<-start

			for range 1000 {
				if counter.Get("spaceID", "environmentID", "contentTypeID") != 0 {
					unexpected.Add(1)
				}
			}
		})
	}

	close(start)
	waitGroup.Wait()

	assert.Zero(t, unexpected.Load())
}
