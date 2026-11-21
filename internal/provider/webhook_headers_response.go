package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func webhookResponseHeaders(headers cm.WebhookDefinitionHeaders, fallbackHeaders TypedMap[TypedObject[WebhookHeaderValue]]) TypedMap[TypedObject[WebhookHeaderValue]] {
	values := make(map[string]TypedObject[WebhookHeaderValue], len(headers))

	for _, header := range headers {
		secret := header.Secret.Or(false)
		value := types.StringNull()

		if returned, present := header.Value.Get(); present {
			value = types.StringValue(returned)
		} else if secret {
			fallback := fallbackHeaders.Elements()[header.Key].Value().Value
			if !fallback.IsUnknown() {
				value = fallback
			}
		}

		values[header.Key] = NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(secret), Value: value})
	}

	return NewTypedMap(values)
}
