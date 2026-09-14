package provider

import (
	"strings"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func signingSecretErrorDetail(response any, err error, values ...types.String) string {
	return redactSigningSecretValues(util.ErrorDetailFromContentfulManagementResponse(response, err), values...)
}

func redactSigningSecretValues(text string, values ...types.String) string {
	redacted := text

	for _, value := range values {
		if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
			redacted = strings.ReplaceAll(redacted, value.ValueString(), "***")
		}
	}

	return redacted
}
