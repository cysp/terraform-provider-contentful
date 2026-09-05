package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const webhookFilterDocDescription = "Contentful webhook payload property to evaluate. Supported paths are `sys.id`, `sys.environment.sys.id`, `sys.contentType.sys.id` for Entry events, `sys.createdBy.sys.id` and `sys.updatedBy.sys.id` except for Unpublish and Delete events, and `sys.deletedBy.sys.id` for Unpublish and Delete events."

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Filtering constraints applied after `topics`. Contentful combines all configured filters with logical AND, and each filter must configure exactly one of `equals`, `in`, `regexp`, or `not`. When `filters` is null or omitted, Contentful triggers the webhook only for the `master` environment by default; set `filters = []` to send no filtering constraints.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookFilterValue]().CustomType(ctx),
		},
		CustomType: NewTypedListNull[TypedObject[WebhookFilterValue]]().CustomType(ctx),
		Optional:   optional,
		Validators: []validator.List{
			listvalidator.NoNullValues(),
		},
	}
}

func (v WebhookFilterEqualsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Description: webhookFilterDocDescription,
			Required:    true,
		},
		"value": schema.StringAttribute{
			Description: "Literal value to compare with the selected payload property. Contentful requires 1 to 255 characters containing only letters, digits, underscores, hyphens, and dots.",
			Required:    true,
		},
	}
}

func (v WebhookFilterInValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Description: webhookFilterDocDescription,
			Required:    true,
		},
		"values": schema.ListAttribute{
			Description: "One or more literal values to compare with the selected payload property. Contentful requires each value to contain 1 to 255 characters using only letters, digits, underscores, hyphens, and dots.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}

func (v WebhookFilterNotValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Description: "Inverts an equality match. Configure exactly one operator within `not`.",
			Attributes:  WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
		},
		"in": schema.SingleNestedAttribute{
			Description: "Inverts an inclusion match. Configure exactly one operator within `not`.",
			Attributes:  WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
		},
		"regexp": schema.SingleNestedAttribute{
			Description: "Inverts a regular-expression match. Configure exactly one operator within `not`.",
			Attributes:  WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("equals", "in", "regexp"),
		},
	}
}

func (v WebhookFilterRegexpValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Description: webhookFilterDocDescription,
			Required:    true,
		},
		"pattern": schema.StringAttribute{
			Description: "Regular-expression pattern matched against the selected payload property. Contentful requires a pattern between 1 and 1024 characters.",
			Required:    true,
		},
	}
}

func (v WebhookFilterValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"not": schema.SingleNestedAttribute{
			Description: "Inverts exactly one nested `equals`, `in`, or `regexp` filter.",
			Attributes:  WebhookFilterNotValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterNotValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"equals": schema.SingleNestedAttribute{
			Description: "Matches when the selected payload property equals the configured value.",
			Attributes:  WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterEqualsValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"in": schema.SingleNestedAttribute{
			Description: "Matches when the selected payload property equals at least one configured value.",
			Attributes:  WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterInValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
		"regexp": schema.SingleNestedAttribute{
			Description: "Matches the selected payload property against a regular-expression pattern.",
			Attributes:  WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[WebhookFilterRegexpValue]().CustomType(ctx),
			Optional:    true,
			Validators:  exactlyOneWebhookFilterOperator("not", "equals", "in", "regexp"),
		},
	}
}

func exactlyOneWebhookFilterOperator(names ...string) []validator.Object {
	expressions := make([]path.Expression, 0, len(names))
	for _, name := range names {
		expressions = append(expressions, path.MatchRelative().AtParent().AtName(name))
	}

	return []validator.Object{objectvalidator.ExactlyOneOf(expressions...)}
}

func (v WebhookHeaderValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Description: "Header value. Contentful's secret flag does not make the Terraform value sensitive. Use a sensitive Terraform expression when needed. Sensitivity obscures CLI output; it does not encrypt or omit plan or state data.",
			Required:    true,
		},
		"secret": schema.BoolAttribute{
			Description: "Whether Contentful treats the header value as secret. Defaults to `false`.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func (v WebhookTransformationValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"method": schema.StringAttribute{
			Description: "HTTP method for outgoing webhook requests. Contentful defaults to `POST` and supports `POST`, `GET`, `PUT`, `PATCH`, and `DELETE`. `GET` and `DELETE` webhook calls do not include a request body.",
			Optional:    true,
		},
		"content_type": schema.StringAttribute{
			Description: "Content-Type for transformed webhook requests. Contentful defaults to `application/vnd.contentful.management.v1+json` and also supports its UTF-8 variant, `application/json` with or without the UTF-8 charset, and `application/x-www-form-urlencoded` with or without the UTF-8 charset.",
			Optional:    true,
		},
		"include_content_length": schema.BoolAttribute{
			Description: "Whether Contentful includes a `Content-Length` header computed from the transformed request body. Contentful omits the header by default.",
			Optional:    true,
		},
		"body": schema.StringAttribute{
			Description: "JSON-encoded custom webhook request body. Use `jsonencode(...)` to construct structured values. Contentful can resolve supported JSON-pointer templates and transformation helpers against the original webhook context.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
	}
}
