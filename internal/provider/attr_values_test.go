package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func GenerateInvalidValueFromAttributesTestcases(t *testing.T, attributes map[string]attr.Value) map[string]map[string]attr.Value {
	t.Helper()

	type InvalidAttrValue struct {
		attr.Value
	}

	testcases := map[string]map[string]attr.Value{
		"empty": {},
	}

	for name := range attributes {
		testcaseAbsent := make(map[string]attr.Value)

		for key, value := range attributes {
			if key != name {
				testcaseAbsent[key] = value
			}
		}

		testcases[name+"/absent"] = testcaseAbsent

		testcaseInvalid := make(map[string]attr.Value)

		for key, value := range attributes {
			testcaseInvalid[key] = value
			if key == name {
				testcaseInvalid[key] = InvalidAttrValue{}
			}
		}

		testcases[name+"/invalid"] = testcaseInvalid
	}

	return testcases
}
