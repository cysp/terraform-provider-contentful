package provider_test

import (
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceWarnsAboutIncompleteUpdate(t *testing.T) {
	t.Parallel()

	const (
		address      = "contentful_webhook.test"
		dependency   = "depends_on = [terraform_data.prerequisite]"
		prerequisite = "\nresource \"terraform_data\" \"prerequisite\" { input = \"planned\" }"
	)

	withPrerequisite := func(username, password, extra string) string {
		return webhookBasicAuthConfig("Existing", username, password, extra+"\n"+dependency) + prerequisite
	}
	for _, test := range []struct{ name, invalid, replacement, diagnostic string }{
		{
			name: "missing password", diagnostic: "Missing webhook Basic password",
			invalid:     withPrerequisite(`"user"`, "null", ""),
			replacement: withPrerequisite(`"user"`, `"replacement"`, "lifecycle { ignore_changes = [http_basic_password] }"),
		},
		{
			name: "missing username", diagnostic: "Missing webhook Basic username",
			invalid: withPrerequisite("null", "null", `http_basic_password_wo = "replacement"`),
			replacement: withPrerequisite(`"user"`, "null", `http_basic_password_wo = "replacement"
lifecycle { ignore_changes = [http_basic_username] }`),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			recorder := &taxonomyRequestBodyRecorder{next: server}

			var warned atomic.Bool
			testWebhookPlanWarning(t, recorder, "Incomplete Basic credentials for an in-place webhook update", &warned, []resource.TestStep{
				{Config: webhookBasicAuthConfig("Existing", "null", "null", "")},
				{Config: test.invalid, PlanOnly: true, ExpectNonEmptyPlan: true, PreConfig: func() { require.False(t, warned.Load()) }},
				{Config: test.invalid, ExpectError: regexp.MustCompile(test.diagnostic), PreConfig: func() { require.True(t, warned.Load(), "incomplete in-place pair needs a planning warning") }},
				{
					Config: test.replacement, Taint: []string{address}, PreConfig: func() { require.Empty(t, recorder.matchingRequests(http.MethodPut, "")) },
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("terraform_data.prerequisite", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionReplace),
					}},
				},
				{
					Config:   webhookBasicAuthConfig("Delegated rename", `"user"`, "null", dependency) + prerequisite,
					PlanOnly: true, ExpectNonEmptyPlan: true,
					PreConfig: func() { warned.Store(false) },
				},
				{
					Config: withPrerequisite(`"changed"`, "terraform_data.pending.output", "") + `
resource "terraform_data" "pending" { input = "resolved-at-apply" }
`,
					PlanOnly: true, ExpectNonEmptyPlan: true,
					PreConfig: func() { require.False(t, warned.Load(), "delegation is valid") },
				},
			})
			require.False(t, warned.Load(), "an unresolved counterpart can still form a valid pair")
		})
	}
}
