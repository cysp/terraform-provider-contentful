package provider

import (
	"encoding/json"
	"strings"
)

// testJSON encodes independently authored fixtures and expectations.
// Unsupported fixture values are test programming errors, including at package initialization.
func testJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func testJSONObject(body string) map[string]any {
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()

	var document map[string]any

	err := decoder.Decode(&document)
	if err != nil {
		panic(err)
	}

	return document
}

func mutateTestJSON(body string, mutate func(map[string]any)) string {
	document := testJSONObject(body)
	mutate(document)

	return testJSON(document)
}
