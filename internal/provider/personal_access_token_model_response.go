package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPersonalAccessTokenResourceModelFromResponse(_ context.Context, personalAccessToken cm.PersonalAccessToken, existingToken types.String, expiresIn types.Int64) (PersonalAccessTokenModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := PersonalAccessTokenModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(personalAccessToken.Sys.ID),
	}

	model.Name = types.StringValue(personalAccessToken.Name)

	model.Scopes = NewTypedListFromStringSlice(personalAccessToken.Scopes)

	model.Token = existingToken
	if token, ok := personalAccessToken.Token.Get(); ok {
		model.Token = types.StringValue(token)
	}

	model.ExpiresIn = expiresIn
	model.ExpiresAt = timetypes.NewRFC3339TimePointerValue(personalAccessToken.Sys.ExpiresAt.ValueTimePointer())

	model.RevokedAt = timetypes.NewRFC3339TimePointerValue(personalAccessToken.RevokedAt.ValueTimePointer())

	return model, diags
}
