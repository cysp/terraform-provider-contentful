package provider_test

import (
	"encoding/json"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookDefinitionBasicPasswordJSONPresence(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response string
		empty    bool
		null     bool
	}{
		"absent": {response: rawWebhookWithoutBasicPassword, empty: true},
		"null":   {response: `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"type":"WebhookDefinition","id":"webhook","version":1},"name":"Webhook","url":"https://example.com/webhook","topics":[],"httpBasicPassword":null}`, null: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var response cm.WebhookDefinition
			require.NoError(t, json.Unmarshal([]byte(test.response), &response))
			assert.Equal(t, test.empty, response.HttpBasicPassword.IsEmpty())
			assert.Equal(t, test.null, response.HttpBasicPassword.IsNull())
		})
	}
}

func TestWebhookMutationResponsePreservesPasswordOnlyWhenAbsent(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		responsePassword cm.OptNilString
		planPassword     types.String
		expectedPassword types.String
		expectError      bool
	}{
		"known plan": {
			planPassword:     types.StringValue("planned-password"),
			expectedPassword: types.StringValue("planned-password"),
		},
		"null plan": {
			planPassword:     types.StringNull(),
			expectedPassword: types.StringNull(),
		},
		"explicit null response": {
			responsePassword: cm.NewOptNilStringNull(),
			planPassword:     types.StringValue("planned-password"),
			expectedPassword: types.StringNull(),
			expectError:      true,
		},
		"different returned value": {
			responsePassword: cm.NewOptNilString("returned-password"),
			planPassword:     types.StringValue("planned-password"),
			expectedPassword: types.StringValue("returned-password"),
			expectError:      true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := cm.WebhookDefinition{
				Sys:               cm.NewWebhookDefinitionSys("space", "webhook"),
				Name:              "Webhook",
				URL:               "https://example.com/webhook",
				HttpBasicPassword: test.responsePassword,
			}
			plan := WebhookModel{
				Name:              types.StringUnknown(),
				URL:               types.StringUnknown(),
				HTTPBasicPassword: test.planPassword,
				Filters:           NewTypedListNull[TypedObject[WebhookFilterValue]](),
				Headers:           NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
				Topics:            NewTypedList([]types.String{}),
			}

			state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(t.Context(), response, plan)

			require.False(t, responseDiags.HasError())
			assert.Equal(t, test.expectError, consistencyDiags.HasError())
			assert.Equal(t, test.expectedPassword, state.HTTPBasicPassword)

			for _, diagnostic := range consistencyDiags {
				assert.NotContains(t, strings.Join([]string{diagnostic.Summary(), diagnostic.Detail()}, " "), "planned-password")
			}
		})
	}
}
