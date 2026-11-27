package provider

import (
	"context"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestWebhookCreationObservations(t *testing.T) {
	t.Parallel()

	prior := time.Date(2026, 11, 25, 1, 2, 3, 123456789, time.UTC)
	same := cm.NewOptDateTime(prior.In(time.FixedZone("other", 3600)))
	later := cm.NewOptDateTime(prior.Add(time.Nanosecond))

	absent := cm.OptDateTime{}
	for _, test := range []struct {
		name               string
		prior              *time.Time
		version            types.Int64
		observed           cm.OptDateTime
		recreated, changed bool
	}{
		{"legacy baseline", nil, types.Int64Value(0), same, false, true},
		{"same instant", &prior, types.Int64Value(0), same, false, false},
		{"missing prior version", &prior, types.Int64Null(), later, true, true},
		{"no ordering evidence", nil, types.Int64Null(), absent, false, false},
		{"omitted", &prior, types.Int64Value(0), absent, false, false},
		{"lower version missing timestamp", &prior, types.Int64Value(1), absent, true, true},
		{"lower version same instant", &prior, types.Int64Value(1), same, true, false},
		{"lower version without prior marker", nil, types.Int64Value(1), absent, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := observeWebhookCreation(test.prior, test.version, cm.WebhookDefinitionSys{CreatedAt: test.observed})
			require.Equal(t, test.recreated, got.recreated)
			require.Equal(t, test.changed, got.changed)
		})
	}
}

type failingWebhookCreationTime struct {
	writeOnlyTestPrivateData

	reject bool
}

func (p *failingWebhookCreationTime) SetKey(ctx context.Context, key string, data []byte) diag.Diagnostics {
	if key == createdAtPrivateKey && p.reject {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Creation time write failed", "Injected failure.")}
	}

	return p.writeOnlyTestPrivateData.SetKey(ctx, key, data)
}

func TestWebhookCreationTimePublicationRetry(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	private := &failingWebhookCreationTime{data: map[string][]byte{
		"created_at":                    []byte(`"2026-11-20T01:00:00Z"`),
		writeOnlySecretHashesPrivateKey: []byte(`"prior"`),
	}, reject: true}
	state := tfsdk.State{Schema: WebhookResourceSchema(ctx)}
	identitySchema := resourceIdentitySchema(webhookIdentityAttributeNames())
	identity := tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil)}
	model := WebhookModel{WebhookID: types.StringValue("existing"), HTTPBasicUsername: types.StringValue("user"), HTTPBasicPassword: types.StringValue("prior"), Timeouts: TimeoutsNull(), Version: types.Int64Value(1)}
	require.Empty(t, state.Set(ctx, &model))
	before := state.Raw
	model.HTTPBasicPassword = types.StringNull()
	observed := cm.WebhookDefinitionSys{Version: 1, CreatedAt: cm.NewOptDateTime(time.Date(2026, 11, 21, 1, 0, 0, 0, time.UTC))}
	publish := func() diag.Diagnostics {
		prior, diags := readWebhookCreationTime(ctx, private)
		require.Empty(t, diags)

		creation := observeWebhookCreation(prior, model.Version, observed)
		require.True(t, creation.recreated)

		return publishWebhookState(ctx, private, writeOnlySecretVerifierChange{changed: true}, creation, &identity, &state, model, 1, true)
	}
	result := publish()
	require.Len(t, result, 1)
	require.Equal(t, "Creation time write failed", result[0].Summary())
	require.True(t, before.Equal(state.Raw))
	require.Empty(t, private.data[writeOnlySecretHashesPrivateKey])
	require.JSONEq(t, `"2026-11-20T01:00:00Z"`, string(private.data["created_at"]))
	private.reject = false

	require.Empty(t, publish())
	require.False(t, before.Equal(state.Raw))
	require.JSONEq(t, `"2026-11-21T01:00:00Z"`, string(private.data["created_at"]))
}
