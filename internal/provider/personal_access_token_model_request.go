package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *PersonalAccessTokenModel) ToPersonalAccessTokenRequestData(_ context.Context) (cm.PersonalAccessTokenRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if model.Scopes.IsUnknown() {
		diags.AddAttributeError(
			path.Root("scopes"),
			"Unexpected unknown scopes",
			"Personal access token scopes must be known before they can be sent to Contentful.",
		)

		return cm.PersonalAccessTokenRequestData{}, diags
	}

	if model.Scopes.IsNull() {
		diags.AddAttributeError(
			path.Root("scopes"),
			"Unexpected null scopes",
			"Personal access token scopes are required.",
		)

		return cm.PersonalAccessTokenRequestData{}, diags
	}

	scopes, scopeDiags := knownStringListElements(path.Root("scopes"), model.Scopes.Elements())
	diags.Append(scopeDiags...)

	if diags.HasError() {
		return cm.PersonalAccessTokenRequestData{}, diags
	}

	req := cm.PersonalAccessTokenRequestData{
		Name:   model.Name.ValueString(),
		Scopes: scopes,
	}

	if !model.ExpiresIn.IsNull() && !model.ExpiresIn.IsUnknown() {
		req.ExpiresIn = cm.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	}

	return req, diags
}
