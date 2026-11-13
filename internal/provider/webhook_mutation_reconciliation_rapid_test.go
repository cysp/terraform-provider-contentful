package provider_test

import (
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestWebhookMutationReconciliationProperties(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rapid.Check(t, func(t *rapid.T) {
		plannedJSON, returnedJSON, reversedJSON := drawReconciliationJSON(t)
		values := drawReconciliationDuplicates(t, "Entry", "Asset")
		kinds := drawReconciliationDuplicates(t, "in", "equals")
		plannedValues := rapid.Permutation(values).Draw(t, "planned in values")
		returnedValues := rapid.Permutation(values).Draw(t, "returned in values")
		plannedKinds := rapid.Permutation(kinds).Draw(t, "planned filters")
		returnedKinds := rapid.Permutation(kinds).Draw(t, "returned filters")
		negated := rapid.Bool().Draw(t, "negated in")
		active := rapid.Bool().Draw(t, "returned active")
		secret := rapid.Bool().Draw(t, "secret header")
		header := NewTypedObject(WebhookHeaderValue{Value: types.StringValue("header-value"), Secret: types.BoolValue(secret)})

		plan := completeWebhookMutationPlan()
		plan.Filters = NewTypedList(reconciliationWebhookFilters(plannedKinds, plannedValues, negated))
		transformation := plan.Transformation.Value()
		transformation.Body = jsontypes.NewNormalizedValue(plannedJSON)
		plan.Transformation = NewTypedObject(transformation)
		plan.Active = types.BoolValue(active)

		plan.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"X-Test": header})
		if rapid.Bool().Draw(t, "response-owned active") {
			plan.Active = types.BoolUnknown()
		}

		unknownHeaders := rapid.Bool().Draw(t, "response-owned headers")
		if unknownHeaders {
			plan.Headers = NewTypedMapUnknown[TypedObject[WebhookHeaderValue]]()
		}

		webhookID := "webhook"

		if rapid.Bool().Draw(t, "generated identity") {
			plan.ID = types.StringUnknown()
			plan.WebhookID = types.StringUnknown()
			webhookID = rapid.StringMatching(`[a-z][a-z0-9]{0,7}`).Draw(t, "created webhook ID")
		}

		response := completeWebhookMutationResponse()
		response.Sys.ID = webhookID
		response.Filters.SetTo(reconciliationWebhookResponseFilters(returnedKinds, returnedValues, negated))
		responseTransformation := response.Transformation.Value
		responseTransformation.Body = []byte(returnedJSON)
		response.Transformation.SetTo(responseTransformation)
		response.Active = cm.NewOptBool(active)

		response.Headers[0].Secret = cm.NewOptBool(secret)
		if secret && !unknownHeaders && rapid.Bool().Draw(t, "redacted header") {
			response.Headers[0].Value = cm.OptString{}
		}

		// The API omits the Basic password. Its Plan fallback and a redacted
		// secret header survive even when another attribute contradicts Plan.
		expected := WebhookModel{
			IDIdentityModel: NewIDIdentityModelFromMultipartID("space", webhookID),
			SpaceID:         types.StringValue("space"), WebhookID: types.StringValue(webhookID),
			Name:              plan.Name,
			URL:               plan.URL,
			Topics:            plan.Topics,
			Filters:           NewTypedList(reconciliationWebhookFilters(returnedKinds, returnedValues, negated)),
			HTTPBasicUsername: plan.HTTPBasicUsername,
			HTTPBasicPassword: plan.HTTPBasicPassword,
			Headers:           NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"X-Test": header}),
			Active:            types.BoolValue(active),
		}
		transformation.Body = jsontypes.NewNormalizedValue(returnedJSON)
		expected.Transformation = NewTypedObject(transformation)

		state, responseDiags, consistencyDiags := ReconcileWebhookMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags)
		assert.Equal(t, plan.Filters, state.Filters, "equivalent duplicate permutations must restore exact Plan")
		assert.Equal(t, plan.Transformation, state.Transformation, "equivalent JSON must restore exact Plan text")
		assert.Equal(t, expected.Headers, state.Headers)
		assert.Equal(t, expected.Active, state.Active)
		assert.Equal(t, expected.IDIdentityModel, state.IDIdentityModel)

		var expectedPath string

		switch rapid.IntRange(0, 7).Draw(t, "contradiction") {
		case 0:
			changed := slices.Clone(kinds)
			changed[0] = "equals"
			changed = rapid.Permutation(changed).Draw(t, "changed filters")
			response.Filters.SetTo(reconciliationWebhookResponseFilters(changed, returnedValues, negated))
			expected.Filters = NewTypedList(reconciliationWebhookFilters(changed, returnedValues, negated))
			expectedPath = "filters"
		case 1:
			// Keep the same nested value set and length, changing one occurrence.
			changed := slices.Clone(values)
			changed[0] = "Asset"

			indices := make([]int, 0, len(returnedKinds))
			for index, kind := range returnedKinds {
				if kind == "in" {
					indices = append(indices, index)
				}
			}

			index := rapid.SampledFrom(indices).Draw(t, "in filter index")
			response.Filters.Value[index] = reconciliationWebhookResponseFilters([]string{"in"}, changed, negated)[0]
			expected.Filters.Elements()[index] = reconciliationWebhookFilters([]string{"in"}, changed, negated)[0]
			expectedPath = "filters"
		case 2:
			response.Topics[0], response.Topics[1] = response.Topics[1], response.Topics[0]
			expected.Topics = NewTypedListFromStringSlice([]string{"Entry.save", "Entry.create"})
			expectedPath = "topics"
		case 3:
			responseTransformation.Body = []byte(reversedJSON)
			response.Transformation.SetTo(responseTransformation)

			transformation.Body = jsontypes.NewNormalizedValue(reversedJSON)
			expected.Transformation = NewTypedObject(transformation)
			expectedPath = "transformation"
		case 4:
			// A known Optional+Computed Plan remains constrained even if the
			// equivalent case above exercised response ownership.
			plan.Active = types.BoolValue(active)
			response.Active = cm.NewOptBool(!active)
			expected.Active = types.BoolValue(!active)
			expectedPath = "active"
		case 5:
			plan.Headers = expected.Headers
			response.Headers[0].Value = cm.NewOptString("different-header")
			expected.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"X-Test": NewTypedObject(WebhookHeaderValue{
				Value: types.StringValue("different-header"), Secret: types.BoolValue(secret),
			})})
			expectedPath = "headers"
		case 6:
			response.Sys.Space.Sys.ID = "other-space"
			expectedPath = "space_id"
		case 7:
			plan.Filters = NewTypedListNull[TypedObject[WebhookFilterValue]]()
			response.Filters.SetTo([]cm.WebhookDefinitionFilter{})
			expected.Filters = NewTypedList([]TypedObject[WebhookFilterValue]{})
			expectedPath = "filters"
		}

		state, responseDiags, consistencyDiags = ReconcileWebhookMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags, expectedPath)
		assert.Equal(t, expected, state, "contradiction must retain complete response and narrow omission fallbacks")
	})
}

func reconciliationWebhookFilters(kinds, values []string, negated bool) []TypedObject[WebhookFilterValue] {
	filters := make([]TypedObject[WebhookFilterValue], len(kinds))
	for index, kind := range kinds {
		if kind == "equals" {
			filters[index] = webhookEqualsFilter(types.StringValue("sys.type"), types.StringValue("Entry"))

			continue
		}

		filter := WebhookFilterValue{
			Not: NewTypedObjectNull[WebhookFilterNotValue](), Equals: NewTypedObjectNull[WebhookFilterEqualsValue](),
			In: webhookInValue(types.StringValue("sys.type"), NewTypedListFromStringSlice(values)), Regexp: NewTypedObjectNull[WebhookFilterRegexpValue](),
		}
		if negated {
			filter.Not = NewTypedObject(webhookNotValue(filter.Equals, filter.In, filter.Regexp))
			filter.In = NewTypedObjectNull[WebhookFilterInValue]()
		}

		filters[index] = NewTypedObject(filter)
	}

	return filters
}

func reconciliationWebhookResponseFilters(kinds, values []string, negated bool) []cm.WebhookDefinitionFilter {
	filters := make([]cm.WebhookDefinitionFilter, len(kinds))
	for index, kind := range kinds {
		if kind == "equals" {
			filters[index] = cmWebhookEqualsFilter("Entry")

			continue
		}

		in := cm.WebhookDefinitionFilterIn{[]byte(testJSON(map[string]string{"doc": "sys.type"})), []byte(testJSON(values))}
		if negated {
			filters[index].Not = cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{In: in})
		} else {
			filters[index].In = in
		}
	}

	return filters
}
