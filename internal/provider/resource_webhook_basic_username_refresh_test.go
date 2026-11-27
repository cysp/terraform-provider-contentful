package provider_test

import (
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func setRemoteWebhookBasic(t *testing.T, fixture *webhookSecretRecoveryFixture, version int, user, password string) {
	t.Helper()
	result, err := fixture.server.Handler().UpdateWebhookDefinition(t.Context(), &cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString(user), HttpBasicPassword: cm.NewOptNilString(password),
		Headers: cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true)}, {Key: "X-Sibling", Secret: cm.NewOptBool(true)}},
	}, cm.UpdateWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: "imported", XContentfulVersion: cm.NewOptInt(version)})
	require.NoError(t, err)
	require.IsType(t, &cm.WebhookDefinitionStatusCode{}, result)
}

func TestAccWebhookResourceOrdinaryPasswordAfterExternalUsernameChange(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	config := webhookRecoveryConfig(true, "")
	initial := importedWebhookRecoveryStep(t, fixture)
	ordinary := webhookRecoveryStep(strings.Replace(config, "http_basic_password_wo = var.password", `http_basic_password = "`+webhookPriorPassword+`"`, 1), webhookPriorPassword, webhookPriorHeader)
	ordinary.PostApplyFunc = func() {
		require.Equal(t, webhookPriorPassword, observations.lastApply().model(t).HTTPBasicPassword.ValueString())
	}
	change := webhookRecoveryStep(strings.Replace(config, `http_basic_username    = "user"`, `http_basic_username    = "other"`, 1), webhookPriorPassword, webhookPriorHeader)
	change.PreConfig = func() { setRemoteWebhookBasic(t, fixture, 1, "other", webhookRotatedPassword) }
	change.Check = func(*terraform.State) error {
		read := observations.lastRead().model(t)
		require.Equal(t, "other", read.HTTPBasicUsername.ValueString())
		require.True(t, read.HTTPBasicPassword.IsNull())
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookPUTWithFields("2", webhookPriorPassword, "", map[string]any{"httpBasicUsername": "other"}))
		fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
		basic := observations.lastApply().verifiers(t).Basic
		require.NotNil(t, basic)
		require.Equal(t, "other", basic.Username)

		return nil
	}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, ordinary, change})
}
