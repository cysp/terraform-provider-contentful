package provider_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentDeletionWaitsForNotFound(t *testing.T) {
	t.Parallel()

	var reads atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/spaces/space/preview_environments/preview", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		body := `{"sys":{"type":"PreviewEnvironment","id":"preview","version":0,"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}},"name":"Preview","description":"","configurations":[]}`

		if reads.Add(1) >= 3 {
			w.WriteHeader(http.StatusNotFound)

			body = `{"sys":{"type":"Error","id":"NotFound"},"message":"Preview environment not found"}`
		}

		_, err := io.WriteString(w, body)
		if err != nil {
			t.Errorf("write preview response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(server.Client()))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	waitForPreviewEnvironmentDeletion(ctx, t, client, "space", "preview")
	assert.Equal(t, int32(3), reads.Load())
}
