// Same-package access exercises the private union operation with observable converters.
package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	stateNames := []string{"null", "unknown", "known"}

	// Exhaust all 364 ordered state vectors of arity 0..5. The 349 invalid
	// vectors must never convert. Each of the 15 valid vectors covers three
	// converter outcomes with both empty and nonempty inputs: 439 leaf cases.
	for arity := range 6 {
		combinations := 1
		for range arity {
			combinations *= len(stateNames)
		}

		for combination := range combinations {
			states := make([]int, arity)
			names := make([]string, arity)
			knownCount, unknownCount := 0, 0
			remaining := combination

			for index := range arity {
				states[index] = remaining % len(stateNames)
				remaining /= len(stateNames)
				names[index] = stateNames[states[index]]

				switch states[index] {
				case 1:
					unknownCount++
				case 2:
					knownCount++
				}
			}

			name := strings.Join(names, "_")
			if arity == 0 {
				name = "no_alternatives"
			}

			t.Run(fmt.Sprintf("arity_%d/%s", arity, name), func(t *testing.T) {
				t.Parallel()

				if knownCount != 1 || unknownCount != 0 {
					checkRequestUnionStateCombination(t, states, "", diag.SeverityInvalid)

					return
				}

				for _, input := range []struct{ name, value string }{{"empty", ""}, {"nonempty", "value"}} {
					for _, outcome := range []struct {
						name     string
						severity diag.Severity
					}{{"success", diag.SeverityInvalid}, {"warning", diag.SeverityWarning}, {"error", diag.SeverityError}} {
						t.Run(input.name+"/"+outcome.name, func(t *testing.T) {
							t.Parallel()
							checkRequestUnionStateCombination(t, states, input.value, outcome.severity)
						})
					}
				}
			})
		}
	}
}

func checkRequestUnionStateCombination(t *testing.T, states []int, knownValue string, conversionSeverity diag.Severity) {
	t.Helper()

	type diagnosticExpectation struct {
		severity diag.Severity
		path     string
	}

	unionPath := path.Root("items").AtListIndex(3).AtName("union")
	alternatives := make([]knownUnionAlternative[string], len(states))
	wantDiagnostics := []diagnosticExpectation{}
	known := []int{}
	calls := []int{}
	unknownCount := 0

	for index, state := range states {
		name := fmt.Sprintf("alternative_%d", index)
		valuePath := unionPath.AtName(name)
		input := ""

		if knownValue != "" {
			input = fmt.Sprintf("%s-%d", knownValue, index)
		}

		value := []types.String{types.StringNull(), types.StringUnknown(), types.StringValue(input)}[state]
		alternatives[index] = knownUnionAlternative[string]{
			Name: name, Path: valuePath, Value: value,
			Convert: func() (string, diag.Diagnostics) {
				calls = append(calls, index)

				switch conversionSeverity {
				case diag.SeverityWarning:
					return value.ValueString(), diag.Diagnostics{diag.NewAttributeWarningDiagnostic(valuePath, "Warning", "Representable value")}
				case diag.SeverityError:
					return "partial", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Error", "Unrepresentable value")}
				case diag.SeverityInvalid:
				}

				return value.ValueString(), nil
			},
		}

		switch state {
		case 1:
			wantDiagnostics = append(wantDiagnostics, diagnosticExpectation{diag.SeverityError, valuePath.String()})
			unknownCount++
		case 2:
			known = append(known, index)
		}
	}

	// Rows: zero / one / multiple known alternatives. Columns: no unknowns /
	// some unknowns. Converter outcomes do not affect union validity.
	decisions := [3][2]struct{ cardinalityError, convert bool }{
		{{cardinalityError: true}, {}},
		{{convert: true}, {}},
		{{cardinalityError: true}, {cardinalityError: true}},
	}
	decision := decisions[min(len(known), 2)][min(unknownCount, 1)]

	if decision.cardinalityError {
		wantDiagnostics = append(wantDiagnostics, diagnosticExpectation{diag.SeverityError, unionPath.String()})
	}

	wantValue := ""
	wantCalls := []int{}

	if decision.convert {
		selected := known[0]
		wantCalls = append(wantCalls, selected)

		if knownValue != "" {
			wantValue = fmt.Sprintf("%s-%d", knownValue, selected)
		}

		if conversionSeverity != diag.SeverityInvalid {
			wantDiagnostics = append(wantDiagnostics, diagnosticExpectation{conversionSeverity, unionPath.AtName(fmt.Sprintf("alternative_%d", selected)).String()})
		}

		if conversionSeverity == diag.SeverityError {
			wantValue = ""
		}
	}

	actual, diags := convertExactlyOneKnownAlternative(unionPath, alternatives...)
	actualDiagnostics := make([]diagnosticExpectation, 0, len(diags))

	for _, diagnostic := range diags {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok, "expected attribute diagnostic, got %T", diagnostic)
		actualDiagnostics = append(actualDiagnostics, diagnosticExpectation{diagnostic.Severity(), withPath.Path().String()})
	}

	assert.Equal(t, wantDiagnostics, actualDiagnostics)
	assert.Equal(t, wantCalls, calls)
	assert.Equal(t, wantValue, actual, "conversion errors must discard partial output")
}
