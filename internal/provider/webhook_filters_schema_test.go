package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestWebhookFilterSchemaRequiresExactlyOneOperator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		scope      string
		operators  []string
		wantErrors bool
	}{
		"outer/zero": {
			scope:      "outer",
			wantErrors: true,
		},
		"outer/multiple": {
			scope:      "outer",
			operators:  []string{"equals", "in"},
			wantErrors: true,
		},
		"outer/single": {
			scope:     "outer",
			operators: []string{"equals"},
		},
		"not/zero": {
			scope:      "not",
			wantErrors: true,
		},
		"not/multiple": {
			scope:      "not",
			operators:  []string{"equals", "in"},
			wantErrors: true,
		},
		"not/single": {
			scope:     "not",
			operators: []string{"equals"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := webhookFilterSchemaValue(t, testCase.scope, testCase.operators...)
			operatorName := "equals"
			operatorPath := path.Root("filter").AtName(operatorName)
			operatorValue := filter.Value().Equals
			operatorAttributes := WebhookFilterValue{}.SchemaAttributes(t.Context())

			if testCase.scope == "not" {
				not := filter.Value().Not
				operatorValue = not.Value().Equals
				operatorAttributes = WebhookFilterNotValue{}.SchemaAttributes(t.Context())
				operatorPath = path.Root("filter").AtName("not").AtName(operatorName)
			}

			operatorAttribute, ok := operatorAttributes[operatorName].(schema.SingleNestedAttribute)
			require.True(t, ok)

			config := webhookFilterSchemaConfig(t, filter)

			operatorPathExpression := path.MatchRoot("filter").AtName(operatorName)
			if testCase.scope == "not" {
				operatorPathExpression = path.MatchRoot("filter").AtName("not").AtName(operatorName)
			}

			request := validator.ObjectRequest{
				Config:         config,
				ConfigValue:    DiagsNoErrorsMust(operatorValue.ToObjectValue(t.Context())),
				Path:           operatorPath,
				PathExpression: operatorPathExpression,
			}

			response := validator.ObjectResponse{}
			for _, objectValidator := range operatorAttribute.Validators {
				objectValidator.ValidateObject(t.Context(), request, &response)
			}

			if testCase.wantErrors {
				require.True(t, response.Diagnostics.HasError(), response.Diagnostics)
			} else {
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			}
		})
	}
}

func webhookFilterSchemaValue(t *testing.T, scope string, operators ...string) TypedObject[WebhookFilterValue] {
	t.Helper()
	ctx := t.Context()

	equals := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterEqualsValue](ctx, map[string]attr.Value{
		"doc":   types.StringValue("sys.type"),
		"value": types.StringValue("Entry"),
	}))
	inOperator := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterInValue](ctx, map[string]attr.Value{
		"doc":    types.StringValue("sys.type"),
		"values": NewTypedList([]types.String{types.StringValue("Entry")}),
	}))
	regexp := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterRegexpValue](ctx, map[string]attr.Value{
		"doc":     types.StringValue("sys.type"),
		"pattern": types.StringValue("Entry"),
	}))

	filter := map[string]attr.Value{
		"not":    NewTypedObjectNull[WebhookFilterNotValue](),
		"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
		"in":     NewTypedObjectNull[WebhookFilterInValue](),
		"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
	}

	if scope == "not" {
		not := map[string]attr.Value{
			"equals": NewTypedObjectNull[WebhookFilterEqualsValue](),
			"in":     NewTypedObjectNull[WebhookFilterInValue](),
			"regexp": NewTypedObjectNull[WebhookFilterRegexpValue](),
		}

		for _, operator := range operators {
			switch operator {
			case "equals":
				not[operator] = equals
			case "in":
				not[operator] = inOperator
			case "regexp":
				not[operator] = regexp
			}
		}

		filter["not"] = DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterNotValue](ctx, not))
	} else {
		for _, operator := range operators {
			switch operator {
			case "equals":
				filter[operator] = equals
			case "in":
				filter[operator] = inOperator
			case "regexp":
				filter[operator] = regexp
			}
		}
	}

	return DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookFilterValue](ctx, filter))
}

func webhookFilterSchemaConfig(t *testing.T, filter TypedObject[WebhookFilterValue]) tfsdk.Config {
	t.Helper()
	ctx := t.Context()
	filterType := NewTypedObjectNull[WebhookFilterValue]().CustomType(ctx)
	rootValue := types.ObjectValueMust(
		map[string]attr.Type{"filter": filterType},
		map[string]attr.Value{"filter": filter},
	)
	rootRaw, err := rootValue.ToTerraformValue(ctx)
	require.NoError(t, err)

	return tfsdk.Config{
		Raw: rootRaw,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"filter": schema.SingleNestedAttribute{
					Attributes: WebhookFilterValue{}.SchemaAttributes(ctx),
					CustomType: filterType,
					Optional:   true,
				},
			},
		},
	}
}
