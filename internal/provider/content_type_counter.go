package provider

import (
	"sync"
)

type ContentfulContentTypeCounter struct {
	mu sync.RWMutex

	countByContentType map[string]int
}

func (c *ContentfulContentTypeCounter) Get(contentTypeID string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.countByContentType == nil {
		return 0
	}

	return c.countByContentType[contentTypeID]
}

func (c *ContentfulContentTypeCounter) Reset(contentTypeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	countByContentType := c.getOrCreateCountByContentTypeMap()
	delete(countByContentType, contentTypeID)
}

func (c *ContentfulContentTypeCounter) Increment(contentTypeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	countByContentType := c.getOrCreateCountByContentTypeMap()
	countByContentType[contentTypeID]++
}

func (c *ContentfulContentTypeCounter) getOrCreateCountByContentTypeMap() map[string]int {
	if c.countByContentType == nil {
		c.countByContentType = make(map[string]int)
	}

	return c.countByContentType
}
