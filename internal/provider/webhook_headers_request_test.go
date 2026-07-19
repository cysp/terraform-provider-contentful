package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/require"
)

func TestWebhookHeadersRejectNullAndUnknownObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]TypedObject[WebhookHeaderValue]{
		"null":    NewTypedObjectNull[WebhookHeaderValue](),
		"unknown": NewTypedObjectUnknown[WebhookHeaderValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"authorization": value})
			_, diags := ToWebhookDefinitionHeaders(t.Context(), path.Root("headers"), headers)
			require.True(t, diags.HasError())
		})
	}
}
