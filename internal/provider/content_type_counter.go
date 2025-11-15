package provider

import (
	"sync"
)

type ContentfulContentTypeCounter struct {
	m  map[string]int
	mu sync.RWMutex
}

func (c *ContentfulContentTypeCounter) Get(contentTypeID string) int {
	return c.withReadMap(func(m map[string]int) int {
		return m[contentTypeID]
	})
}

func (c *ContentfulContentTypeCounter) Reset(contentTypeID string) {
	c.withWriteMap(func(m map[string]int) {
		delete(m, contentTypeID)
	})
}

func (c *ContentfulContentTypeCounter) Increment(contentTypeID string) {
	c.withWriteMap(func(m map[string]int) {
		m[contentTypeID]++
	})
}

func (c *ContentfulContentTypeCounter) withReadMap(operation func(map[string]int) int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.m == nil {
		return 0 // Default value if map hasn't been created yet
	}

	return operation(c.m)
}

func (c *ContentfulContentTypeCounter) withWriteMap(operation func(map[string]int)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.m == nil {
		c.m = make(map[string]int)
	}

	operation(c.m)
}
