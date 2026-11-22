package provider

import (
	"context"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.MapNestedAttribute{
		Description: "HTTP headers sent by the webhook, keyed by header name. Omit this map on creation to adopt the headers returned by Contentful. If you omit it when changing another webhook setting, Terraform resends ordinary header values and retains stored secrets by exact header name. Set `headers = {}` to remove all custom headers.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookHeaderValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookHeaderValue]().CustomType(ctx),
		},
		CustomType: TypedMap[TypedObject[WebhookHeaderValue]]{}.CustomType(ctx),
		Optional:   optional,
		Computed:   true,
		Validators: []validator.Map{
			mapvalidator.NoNullValues(),
		},
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
			validateWebhookHeaderValues{},
		},
	}
}

// validateWebhookHeaderValues checks the effective map after UseStateForUnknown.
type validateWebhookHeaderValues struct{}

func (validateWebhookHeaderValues) Description(context.Context) string {
	return "Each configured header needs an initial value on creation or replacement. On update, only headers with secret = true may have a null planned value."
}

func (m validateWebhookHeaderValues) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (validateWebhookHeaderValues) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// An unchanged map can contain values inherited through omission or ignore_changes.
	// Request conversion checks those values if an update actually sends a PUT.
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.Equal(req.StateValue) {
		return
	}

	var planned map[string]TypedObject[WebhookHeaderValue]
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planned, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validatePlannedWebhookHeaderValues(req.Path, planned, req.State.Raw.IsNull())...)
}

func validatePlannedWebhookHeaderValues(headersPath path.Path, headers map[string]TypedObject[WebhookHeaderValue], creating bool) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, key := range slices.Sorted(maps.Keys(headers)) {
		header, known := headers[key].GetValue()
		if !known {
			continue
		}

		headerPath := headersPath.AtMapKey(key)
		if creating && header.Value.IsNull() {
			diags.AddAttributeError(headerPath.AtName("value"), "Missing initial webhook header value",
				"Supply an initial value for this header when creating or replacing a webhook.")

			continue
		}

		diags.Append(header.validateValueOmission(headerPath)...)
	}

	return diags
}
