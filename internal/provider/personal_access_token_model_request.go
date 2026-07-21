package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (model *PersonalAccessTokenModel) ToPersonalAccessTokenRequestData(ctx context.Context) (cm.PersonalAccessTokenRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.PersonalAccessTokenRequestData{
		Name: model.Name.ValueString(),
	}

	if model.Scopes.IsNull() || model.Scopes.IsUnknown() {
		if model.Scopes.IsUnknown() {
			diags.AddAttributeError(path.Root("scopes"), "Unexpected unknown scopes", "Personal access token scopes must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path.Root("scopes"), "Unexpected null scopes", "Personal access token scopes are required.")
		}

		return req, diags
	}

	scopes := make([]string, len(model.Scopes.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, model.Scopes, &scopes)...)

	req.Scopes = scopes

	if !model.ExpiresIn.IsNull() && !model.ExpiresIn.IsUnknown() {
		req.ExpiresIn = cm.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	}

	return req, diags
}
