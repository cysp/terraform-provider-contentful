package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppKeyIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
	KeyKID          types.String `tfsdk:"key_kid"`
}

type AppKeyModel struct {
	IDIdentityModel
	AppKeyIdentityModel

	JWK        TypedObject[AppKeyJWKModel] `tfsdk:"jwk"`
	PrivateKey types.String                `tfsdk:"private_key"`
	CreatedAt  timetypes.RFC3339           `tfsdk:"created_at"`
	UpdatedAt  timetypes.RFC3339           `tfsdk:"updated_at"`
	LastUsedAt timetypes.RFC3339           `tfsdk:"last_used_at"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type AppKeyJWKModel struct {
	Alg types.String            `tfsdk:"alg"`
	Kty types.String            `tfsdk:"kty"`
	Use types.String            `tfsdk:"use"`
	X5c TypedList[types.String] `tfsdk:"x5c"`
	Kid types.String            `tfsdk:"kid"`
	X5t types.String            `tfsdk:"x5t"`
}
