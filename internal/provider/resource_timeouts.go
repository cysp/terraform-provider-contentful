package provider

import (
	"time"
)

const defaultResourceOperationTimeout = 2 * time.Minute

const minimumStoredResourceOperationTimeout = 10 * time.Second
