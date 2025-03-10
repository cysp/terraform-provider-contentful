package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type InvalidAttrType struct {
	attr.Type
}

type InvalidAttrValue struct {
	attr.Value
}

//nolint:ireturn
func (v InvalidAttrValue) Type(_ context.Context) attr.Type {
	return InvalidAttrType{}
}

func GenerateInvalidValueFromAttributesTestcases(t *testing.T, attributes map[string]attr.Value) map[string]map[string]attr.Value {
	t.Helper()

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
