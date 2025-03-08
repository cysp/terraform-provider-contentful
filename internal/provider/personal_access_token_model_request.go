package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (model *PersonalAccessTokenModel) ToPersonalAccessTokenRequestFields(ctx context.Context) (cm.PersonalAccessTokenRequestFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.PersonalAccessTokenRequestFields{
		Name: model.Name.ValueString(),
	}

	scopes := make([]string, len(model.Scopes.Elements()))
	diags.Append(model.Scopes.ElementsAs(ctx, &scopes, false)...)

	req.Scopes = scopes

	if !model.ExpiresIn.IsNull() && !model.ExpiresIn.IsUnknown() {
		req.ExpiresIn = cm.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	}

	return req, diags
}
