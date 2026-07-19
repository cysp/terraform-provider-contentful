package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AppKeyResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Contentful App Key.",
		MarkdownDescription: "Manages a Contentful App Key from caller-supplied public key material. The corresponding private key is not sent to Contentful or stored by this resource. Contentful permits three keys per app and requires each public-key fingerprint to be globally unique. Use `lifecycle { create_before_destroy = true }` only when rotating to different key material, a free key slot exists, and the old and new keys must overlap.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in organization_id/app_definition_id/key_kid form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the app.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition for which the app key is created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_kid": schema.StringAttribute{
				Description: "Contentful App Key sys.id. This equals jwk.kid and jwk.x5t.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"jwk": schema.SingleNestedAttribute{
				Description: "Public JSON Web Key for the app key. Generate and retain the corresponding private key outside this resource.",
				CustomType:  NewTypedObjectNull[AppKeyJWKModel]().CustomType(ctx),
				Required:    true,
				Attributes:  AppKeyJWKSchemaAttributes(ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the app key was created.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the app key was last updated.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Description: "Timestamp when the app key was last used.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Delete: true}),
		},
	}
}

func AppKeyJWKSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"alg": schema.StringAttribute{
			Description: "JWK algorithm. Must be RS256.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("RS256"),
			},
		},
		"kty": schema.StringAttribute{
			Description: "JWK key type. Must be RSA.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("RSA"),
			},
		},
		"use": schema.StringAttribute{
			Description: "JWK public key use. Must be sig.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("sig"),
			},
		},
		"kid": schema.StringAttribute{
			Description: "JWK key identifier. This becomes the Contentful app key ID and must match the x5c fingerprint.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"x5c": schema.ListAttribute{
			Description: "JWK public key material. The single value must use canonical standard base64 encoding. Contentful documents RSA-4096 key generation; the provider validates the encoding and its fingerprint relationships without enforcing undocumented key formats or sizes.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
				listvalidator.SizeBetween(1, 1),
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"x5t": schema.StringAttribute{
			Description: "JWK key thumbprint. This must be the unpadded base64url-encoded SHA-256 digest of the decoded bytes in x5c[0].",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}
