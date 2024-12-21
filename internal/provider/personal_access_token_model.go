package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/cysp/terraform-provider-contentful/internal/tf"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *PersonalAccessTokenModel) ToCreatePersonalAccessTokenReq(ctx context.Context) (contentfulManagement.CreatePersonalAccessTokenReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.CreatePersonalAccessTokenReq{
		Name: model.Name.ValueString(),
	}

	scopes, scopesDiags := tf.ElementsAsStringSlice(ctx, model.Scopes)
	diags.Append(scopesDiags...)

	req.Scopes = scopes

	if tf.IsKnownAndPresent(model.ExpiresIn) {
		req.ExpiresIn = contentfulManagement.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	}

	return req, diags
}

func (model *PersonalAccessTokenModel) ReadFromResponse(ctx context.Context, personalAccessToken *contentfulManagement.PersonalAccessToken) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.Id = types.StringValue(personalAccessToken.Sys.ID)

	model.Name = types.StringValue(personalAccessToken.Name)

	scopesList, scopesListDiags := util.NewStringListValueFromStringSlice(ctx, personalAccessToken.Scopes)
	diags.Append(scopesListDiags...)

	model.Scopes = scopesList

	if token, ok := personalAccessToken.Token.Get(); ok {
		model.Token = types.StringValue(token)
	}

	model.ExpiresAt = timetypes.NewRFC3339TimePointerValue(personalAccessToken.Sys.ExpiresAt.ValueTimePointer())

	model.RevokedAt = timetypes.NewRFC3339TimePointerValue(personalAccessToken.RevokedAt.ValueTimePointer())

	return diags
}
