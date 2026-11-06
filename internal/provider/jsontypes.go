package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

var errTrailingJSON = errors.New("unexpected data after JSON value")

func parseNormalizedJSONValue(rawJSON []byte) (jsontypes.Normalized, error) {
	var v any

	decoder := json.NewDecoder(bytes.NewReader(rawJSON))
	decoder.UseNumber()

	err := decoder.Decode(&v)
	if err != nil {
		return jsontypes.NewNormalizedNull(), fmt.Errorf("decode JSON: %w", err)
	}

	err = decoder.Decode(new(any))
	if !errors.Is(err, io.EOF) {
		return jsontypes.NewNormalizedNull(), errTrailingJSON
	}

	normalized, err := json.Marshal(v)
	if err != nil {
		return jsontypes.NewNormalizedNull(), fmt.Errorf("encode JSON: %w", err)
	}

	return jsontypes.NewNormalizedValue(string(normalized)), nil
}

func NormalizeJSON(rawJSON []byte) string {
	value, err := parseNormalizedJSONValue(rawJSON)
	if err != nil {
		return string(rawJSON)
	}

	return value.ValueString()
}

func NewNormalizedJSONValue(rawJSON []byte) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(NormalizeJSON(rawJSON))
}
