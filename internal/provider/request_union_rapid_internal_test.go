// Same-package access exercises the private union operation with observable converters.
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Vary arity, alternative positions, and converter diagnostics independently.
// In particular, unknown alternatives remain errors beside multiple known
// alternatives; no converter may run until the entire union is resolved.
func TestRequestUnionStateCombinations(t *testing.T) {
	t.Parallel()

	values := []types.String{types.StringNull(), types.StringUnknown(), types.StringValue("")}

	type unionDecision struct {
		cardinalityError bool
		convert          bool
	}

	// Rows: zero / one / multiple known alternatives. Columns: no unknowns /
	// some unknowns. This specification is independent of converter diagnostics.
	decisions := [3][2]unionDecision{
		{{cardinalityError: true}, {}},
		{{convert: true}, {}},
		{{cardinalityError: true}, {cardinalityError: true}},
	}

	rapid.Check(t, func(t *rapid.T) {
		states := rapid.SliceOfN(rapid.IntRange(0, 2), 0, 5).Draw(t, "states")
		conversionResult := rapid.IntRange(0, 2).Draw(t, "conversion result")
		unionPath := path.Root("items").AtListIndex(rapid.IntRange(0, 9).Draw(t, "index")).AtName("union")
		alternatives := make([]knownUnionAlternative[string], len(states))
		wantErrors := []string{}
		wantWarnings := []string{}
		known := []int{}
		calls := []int{}
		unknownCount := 0

		for index, state := range states {
			name := fmt.Sprintf("alternative_%d", index)
			valuePath := unionPath.AtName(name)
			alternatives[index] = knownUnionAlternative[string]{
				Name: name, Path: valuePath, Value: values[state],
				Convert: func() (string, diag.Diagnostics) {
					calls = append(calls, index)

					switch conversionResult {
					case 1:
						return name, diag.Diagnostics{diag.NewAttributeWarningDiagnostic(valuePath, "Warning", "Representable value")}
					case 2:
						return "partial", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Error", "Unrepresentable value")}
					default:
						return name, nil
					}
				},
			}

			switch state {
			case 1:
				wantErrors = append(wantErrors, valuePath.String())
				unknownCount++
			case 2:
				known = append(known, index)
			}
		}

		decision := decisions[min(len(known), 2)][min(unknownCount, 1)]
		if decision.cardinalityError {
			wantErrors = append(wantErrors, unionPath.String())
		}

		wantValue := ""
		wantCalls := []int{}

		if decision.convert {
			selected := known[0]
			wantCalls = append(wantCalls, selected)
			wantValue = fmt.Sprintf("alternative_%d", selected)

			switch conversionResult {
			case 1:
				wantWarnings = append(wantWarnings, unionPath.AtName(wantValue).String())
			case 2:
				wantErrors = append(wantErrors, unionPath.AtName(wantValue).String())
				wantValue = ""
			}
		}

		actual, diags := convertExactlyOneKnownAlternative(unionPath, alternatives...)
		errors, warnings := []string{}, []string{}

		for _, diagnostic := range diags {
			withPath, ok := diagnostic.(diag.DiagnosticWithPath)
			require.True(t, ok, "expected attribute diagnostic, got %T", diagnostic)

			if diagnostic.Severity() == diag.SeverityError {
				errors = append(errors, withPath.Path().String())
			} else {
				warnings = append(warnings, withPath.Path().String())
			}
		}

		require.Equal(t, wantErrors, errors)
		require.Equal(t, wantWarnings, warnings)
		require.Equal(t, wantCalls, calls)
		require.Equal(t, wantValue, actual, "conversion errors must discard partial output")
	})
}
