package integration_tests_test

import "encoding/json"

// testJSON encodes independently authored fixtures and expectations.
// Unsupported fixture values are test programming errors, including at package initialization.
func testJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}
