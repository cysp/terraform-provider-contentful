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

const webhookFilterDocDescription = "Webhook payload property path to evaluate, such as `sys.id` or `sys.environment.sys.id`."

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Filtering constraints applied after `topics`. Contentful combines filters with logical AND; each filter configures exactly one of `equals`, `in`, `regexp`, or `not`. Null or omitted filters default to the `master` environment; `filters = []` sends no constraints. Supported payload paths and event restrictions are defined by [Contentful webhook filters](https://www.contentful.com/developers/docs/extensibility/webhooks/filters/).",
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
			Description: "Literal value to compare with the selected payload property.",
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
			Description: "Literal values to compare with the selected payload property.",
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
			Description: "Regular-expression pattern matched against the selected payload property.",
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
			Description: "Header value. Non-secret values can contain Contentful JSON-pointer templates such as `{ /payload/sys/id }`; secret headers and HTTP Basic credentials are not transformed. Contentful's secret flag does not make the Terraform value sensitive. See [Secrets and Terraform state](../guides/secrets-and-state).",
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
			Description: "Content-Type for transformed webhook requests. Defaults to `application/vnd.contentful.management.v1+json`; `application/x-www-form-urlencoded` converts the JSON body to form data. See [Contentful transformations](https://www.contentful.com/developers/docs/extensibility/webhooks/transformations/) for supported media types and templates.",
			Optional:    true,
		},
		"include_content_length": schema.BoolAttribute{
			Description: "Whether Contentful includes a `Content-Length` header computed from the transformed request body. Contentful omits the header by default.",
			Optional:    true,
		},
		"body": schema.StringAttribute{
			Description: "JSON-encoded custom webhook request body. Use `jsonencode(...)` to construct structured values. When omitted, methods that send a body use Contentful's standard event payload. Contentful can resolve supported JSON-pointer templates and transformation helpers against the original webhook context.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
	}
}
