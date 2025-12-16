package provider

import (
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ContentfulContentTypeCounter struct {
	mu sync.RWMutex

	countByContentType cm.SpaceEnvironmentMap[int]
}

func (c *ContentfulContentTypeCounter) Get(spaceID, environmentID, contentTypeID string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.countByContentType.Get(spaceID, environmentID, contentTypeID)
}

func (c *ContentfulContentTypeCounter) Reset(spaceID, environmentID, contentTypeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.countByContentType.Delete(spaceID, environmentID, contentTypeID)
}

func (c *ContentfulContentTypeCounter) Increment(spaceID, environmentID, contentTypeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := c.countByContentType.Get(spaceID, environmentID, contentTypeID)
	c.countByContentType.Set(spaceID, environmentID, contentTypeID, count+1)
}
