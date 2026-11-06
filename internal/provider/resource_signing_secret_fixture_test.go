package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

// The fixture returns independent, fixed timestamp values for successive writes.
// It also supports absent metadata, which is valid for both API models.
type signingSecretFixture struct {
	mu         sync.Mutex
	kind       string
	timestamps bool
	exists     bool
	values     []string
	mutations  []string
}

func (s *signingSecretFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	endpoint := "/organizations/organization/app_definitions/app/signing_secret"
	if s.kind == "webhook" {
		endpoint = "/spaces/space/webhook_settings/signing_secret"
	}

	if r.URL.Path != endpoint {
		http.Error(w, "unexpected endpoint", http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPut:
		var body struct {
			Value string `json:"value"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)

			return
		}

		s.values = append(s.values, body.Value)
		s.mutations = append(s.mutations, r.Method)
		s.exists = true
	case http.MethodDelete:
		s.mutations = append(s.mutations, r.Method)
		s.exists = false

		w.WriteHeader(http.StatusNoContent)

		return
	case http.MethodGet:
	default:
		http.Error(w, "unexpected method", http.StatusMethodNotAllowed)

		return
	}

	if !s.exists {
		w.WriteHeader(http.StatusNotFound)

		response := cmt.NewContentfulManagementError("NotFound", new("Absent"), nil)

		err := json.NewEncoder(w).Encode(&response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	var createdAt, updatedAt cm.OptDateTime

	if s.timestamps {
		day := 1
		if len(s.values) > 1 {
			day = 2
		}

		createdAt = cm.NewOptDateTime(time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC))
		updatedAt = cm.NewOptDateTime(time.Date(2026, time.September, day, 1, 0, 0, 0, time.UTC))
	}

	var response json.Marshaler
	if s.kind == "webhook" {
		response = &cm.WebhookSigningSecret{
			Sys: cm.WebhookSigningSecretSys{
				Type:      cm.WebhookSigningSecretSysTypeWebhookSigningSecret,
				Space:     cm.NewSpaceLink("space"),
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			RedactedValue: "opaque",
		}
	} else {
		response = &cm.AppSigningSecret{
			Sys: cm.AppSigningSecretSys{
				Type:          cm.AppSigningSecretSysTypeAppSigningSecret,
				Organization:  cm.NewOrganizationLink("organization"),
				AppDefinition: cm.NewAppDefinitionLink("app"),
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			},
			RedactedValue: "opaque",
		}
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *signingSecretFixture) requireValues(t *testing.T, values ...string) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	require.Equal(t, values, s.values)
}

func signingSecretConfig(kind, arguments string) string {
	scope := `organization_id = "organization"
app_definition_id = "app"`
	if kind == "webhook" {
		scope = `space_id = "space"`
	}

	return fmt.Sprintf("resource \"contentful_%s_signing_secret\" \"test\" {\n%s\n%s\n}\n", kind, scope, arguments)
}

func (s *signingSecretFixture) requireMutations(t *testing.T, methods ...string) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	require.Equal(t, methods, s.mutations)
}
