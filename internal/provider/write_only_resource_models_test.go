package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretModelWithWriteOnlySecret(t *testing.T) {
	t.Parallel()

	model, values, diags := AppSigningSecretModelWithWriteOnlySecrets(
		AppSigningSecretModel{},
		AppSigningSecretModel{
			ValueWO: types.StringValue("write-only-secret"),
		},
	)

	require.False(t, diags.HasError(), diags)
	assert.Equal(t, "write-only-secret", model.Value.ValueString())
	assert.True(t, model.ValueWO.IsNull())
	assert.Len(t, values, 1)
}

func TestWebhookModelWithWriteOnlyHeaderValue(t *testing.T) {
	t.Parallel()

	plan := WebhookModel{
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
			`x-"secret"\key`: NewTypedObject(WebhookHeaderValue{
				Value:   types.StringNull(),
				ValueWO: types.StringNull(),
				Secret:  types.BoolValue(true),
			}),
		}),
	}

	config := WebhookModel{
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
			`x-"secret"\key`: NewTypedObject(WebhookHeaderValue{
				Value:   types.StringNull(),
				ValueWO: types.StringValue("write-only-secret"),
				Secret:  types.BoolValue(true),
			}),
		}),
	}

	model, values, diags := WebhookModelWithWriteOnlySecrets(plan, config)

	require.False(t, diags.HasError(), diags)

	header := model.Headers.Elements()[`x-"secret"\key`].Value()
	assert.Equal(t, "write-only-secret", header.Value.ValueString())
	assert.True(t, header.ValueWO.IsNull())
	assert.Len(t, values, 1)
	assert.Equal(t, path.Root("headers").AtMapKey(`x-"secret"\key`).AtName("value_wo").String(), values[0].Path.String())
}
