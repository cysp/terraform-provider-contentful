package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func LivePreviewVariablesResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages the complete Contentful live preview variables document in an environment. Existing documents must be imported before management. Updates replace all variables; destroy deletes the document without a version precondition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Composite Terraform resource identifier in space_id/environment_id form.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"space_id": schema.StringAttribute{
				Description:   "ID of the space containing the variables document.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"environment_id": schema.StringAttribute{
				Description:   "Environment ID passed unchanged to Contentful. If it is an alias, subsequent operations follow its routing; the provider does not resolve or bind its target.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"variables": schema.StringAttribute{
				Description: "Complete variables object encoded with jsonencode(...). Values may be strings, null, or locale-keyed objects of strings/nulls. Locale keys must exist in the environment. Omitted variable or locale keys are removed on update. Empty strings, nulls, and empty objects remain distinct; an empty root object keeps a present document. Values are stored in Terraform state.",
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{livePreviewVariablesValidator{}},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
