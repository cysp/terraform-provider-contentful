package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *PersonalAccessTokenModel) ReadFromResponse(ctx context.Context, personalAccessToken *cm.PersonalAccessToken) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.ID = types.StringValue(personalAccessToken.Sys.ID)

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
