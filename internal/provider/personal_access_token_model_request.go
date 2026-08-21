package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *PersonalAccessTokenModel) ToPersonalAccessTokenRequestData(_ context.Context) (cm.PersonalAccessTokenRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, path.Root("name"))
	diags.Append(nameDiags...)

	var scopes []string

	if model.Scopes.IsNull() || model.Scopes.IsUnknown() {
		if model.Scopes.IsUnknown() {
			diags.AddAttributeError(path.Root("scopes"), "Unexpected unknown scopes", "Personal access token scopes must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path.Root("scopes"), "Unexpected null scopes", "Personal access token scopes are required.")
		}
	} else {
		var scopeDiags diag.Diagnostics

		scopes, scopeDiags = knownStringListElements(path.Root("scopes"), model.Scopes.Elements())
		diags.Append(scopeDiags...)
	}

	expiresIn := cm.OptNilInt{}

	switch {
	case model.ExpiresIn.IsUnknown():
		diags.AddAttributeError(
			path.Root("expires_in"),
			"Unexpected unknown integer",
			"The integer value must be known before it can be sent to Contentful.",
		)
	case model.ExpiresIn.IsNull():
	default:
		expiresIn = cm.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	}

	if diags.HasError() {
		return cm.PersonalAccessTokenRequestData{}, diags
	}

	return cm.PersonalAccessTokenRequestData{
		Name:      name,
		Scopes:    scopes,
		ExpiresIn: expiresIn,
	}, diags
}
