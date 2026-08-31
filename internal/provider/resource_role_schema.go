package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rolePermissionValuesValidator struct{}

func (rolePermissionValuesValidator) Description(context.Context) string {
	return `"all" must be the only value in a permission value list`
}

func (v rolePermissionValuesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (rolePermissionValuesValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	values, valueCount, ok := knownStringListValuesForValidation(req.ConfigValue)
	if !ok {
		return
	}

	resp.Diagnostics.Append(validateRolePermissionValues(req.Path, valueCount, values)...)
}

type rolePolicyActionsValidator struct{}

func (rolePolicyActionsValidator) Description(context.Context) string {
	return `"all" must be the only action in a policy action list`
}

func (v rolePolicyActionsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (rolePolicyActionsValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	actions, actionCount, ok := knownStringListValuesForValidation(req.ConfigValue)
	if !ok {
		return
	}

	resp.Diagnostics.Append(validateRolePolicyActions(req.Path, actionCount, actions)...)
}

func knownStringListValuesForValidation(value types.List) ([]string, int, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, 0, false
	}

	elements := value.Elements()
	values := make([]string, 0, len(elements))

	for _, element := range elements {
		stringValue, ok := element.(types.String)
		if !ok || stringValue.IsNull() {
			return nil, 0, false
		}

		if stringValue.IsUnknown() {
			continue
		}

		values = append(values, stringValue.ValueString())
	}

	return values, len(elements), true
}

func RoleResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space where the role exists.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "System ID of the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the role.",
				Optional:    true,
			},
			"permissions": schema.MapAttribute{
				Description: "Map of Contentful permission names to their values. Use an empty list to disable a permission, `[\"read\"]` for read-only access where supported, and `[\"manage\"]` or `[\"all\"]` for read and write access. Terraform `[\"all\"]` is sent to Contentful as the scalar `\"all\"`; `\"all\"` must be the only value in its list.",
				ElementType: NewTypedListNull[types.String]().Type(ctx),
				CustomType:  NewTypedMapNull[TypedList[types.String]]().CustomType(ctx),
				Required:    true,
				Validators: []validator.Map{
					mapvalidator.NoNullValues(),
					mapvalidator.ValueListsAre(
						listvalidator.NoNullValues(),
						rolePermissionValuesValidator{},
					),
				},
			},
			"policies": schema.ListNestedAttribute{
				Description: "Policies allow or deny access to resources in fine-grained detail. For example, limit read access to only entries of a specific content type or write access to only certain parts of an entry (e.g. a specific locale).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: RolePolicyValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectUnknown[RolePolicyValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[RolePolicyValue]]{}.CustomType(ctx),
				Required:   true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (v RolePolicyValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"actions": schema.ListAttribute{
			Description: "Actions that the policy allows or denies. Terraform `[\"all\"]` sends Contentful’s scalar `\"all\"`, which aliases the content actions `read`, `create`, `update`, `delete`, `archive`, `unarchive`, `publish`, and `unpublish`; use `[\"access\"]` for environment access. `\"all\"` must be the only action in the list.",
			ElementType: types.StringType,
			CustomType:  TypedList[types.String]{}.CustomType(ctx),
			Required:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
				rolePolicyActionsValidator{},
			},
		},
		"constraint": schema.StringAttribute{
			Description: "JSON constraint that defines the scope of the policy.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"effect": schema.StringAttribute{
			Description: "Whether the policy allows or denies the specified actions.",
			Required:    true,
		},
	}
}
