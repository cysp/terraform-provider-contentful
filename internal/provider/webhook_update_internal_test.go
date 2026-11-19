package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestWebhookRemoteValuesUnchanged(t *testing.T) {
	t.Parallel()

	state := tfsdk.State{Schema: WebhookResourceSchema(t.Context())}
	model := WebhookModel{
		WebhookID: types.StringValue("webhook"),
		Name:      types.StringValue("Webhook"),
		Timeouts:  TimeoutsNull(),
	}
	require.Empty(t, state.Set(t.Context(), &model))

	for _, test := range []struct {
		name, attribute string
		value           attr.Value
		unchanged       bool
	}{
		{"unknown computed ID", "webhook_id", types.StringUnknown(), true},
		{"conflicting known ID", "webhook_id", types.StringValue("other"), false},
		{"unknown required name", "name", types.StringUnknown(), false},
		{"unknown optional computed headers", "headers", NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](), false},
		{"null filters to empty", "filters", NewTypedList([]TypedObject[WebhookFilterValue]{}), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			plan := tfsdk.Plan(state)
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root(test.attribute), test.value))
			unchanged, diags := webhookRemoteValuesUnchanged(t.Context(), plan, state)
			require.Empty(t, diags)
			require.Equal(t, test.unchanged, unchanged)
		})
	}
}
