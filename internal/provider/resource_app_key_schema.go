package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AppKeyResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
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
				Description: "Key identifier. This is the JWK kid and Contentful app key ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"jwk": schema.SingleNestedAttribute{
				Description: "Public JSON Web Key for the app key. When omitted or null, Contentful generates a key pair and returns the private key once after creation.",
				CustomType:  NewTypedObjectNull[AppKeyJWKModel]().CustomType(ctx),
				Optional:    true,
				Computed:    true,
				Attributes:  AppKeyJWKSchemaAttributes(ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "Private key returned by Contentful when the key pair is generated. This is only available immediately after creation.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
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
			Description: "JWK algorithm. Defaults to RS256 when jwk is configured.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("RS256"),
			Validators: []validator.String{
				stringvalidator.OneOf("RS256"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"kty": schema.StringAttribute{
			Description: "JWK key type. Defaults to RSA when jwk is configured.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("RSA"),
			Validators: []validator.String{
				stringvalidator.OneOf("RSA"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"use": schema.StringAttribute{
			Description: "JWK public key use. Defaults to sig when jwk is configured.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("sig"),
			Validators: []validator.String{
				stringvalidator.OneOf("sig"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"kid": schema.StringAttribute{
			Description: "JWK key identifier. Required when jwk is configured. This becomes the Contentful app key ID.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"x5c": schema.ListAttribute{
			Description: "JWK X.509 certificate chain for the public key. Required when jwk is configured.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
				listvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"x5t": schema.StringAttribute{
			Description: "JWK X.509 certificate SHA-1 thumbprint. Required when jwk is configured.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}
