package provider

type ContentfulContentTypeCounter struct {
	m map[string]int
}

func (c *ContentfulContentTypeCounter) getOrCreateMap() map[string]int {
	if c.m == nil {
		c.m = make(map[string]int)
	}

	return c.m
}

func (c *ContentfulContentTypeCounter) Get(contentTypeID string) int {
	return c.getOrCreateMap()[contentTypeID]
}

func (c *ContentfulContentTypeCounter) Reset(contentTypeID string) {
	delete(c.getOrCreateMap(), contentTypeID)
}

func (c *ContentfulContentTypeCounter) Increment(contentTypeID string) {
	c.getOrCreateMap()[contentTypeID]++
}
