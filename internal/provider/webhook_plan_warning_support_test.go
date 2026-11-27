package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type webhookPlanWarningObserver struct {
	tfprotov6.ProviderServer

	summary string
	warned  *atomic.Bool
}

func testWebhookPlanWarning(t *testing.T, handler http.Handler, summary string, warned *atomic.Bool, steps []resource.TestStep) {
	t.Helper()

	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)
	factories := makeTestAccProtoV6ProviderFactories(testProviderOptionsWithHTTPServer(httpServer)...)
	factory := factories["contentful"]
	factories["contentful"] = func() (tfprotov6.ProviderServer, error) {
		server, err := factory()
		if err != nil {
			return nil, err
		}

		return webhookPlanWarningObserver{ProviderServer: server, summary: summary, warned: warned}, nil
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: steps})
}

func (p webhookPlanWarningObserver) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	response, err := p.ProviderServer.PlanResourceChange(ctx, req)
	if response != nil {
		for _, d := range response.Diagnostics {
			if d.Severity == tfprotov6.DiagnosticSeverityWarning && d.Summary == p.summary && d.Attribute == nil {
				p.warned.Store(true)
			}
		}
	}

	if err != nil {
		return response, fmt.Errorf("observe webhook plan: %w", err)
	}

	return response, nil
}
