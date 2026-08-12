package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func NormalizeJSON(rawJSON []byte) string {
	var v any

	decoder := json.NewDecoder(bytes.NewReader(rawJSON))
	decoder.UseNumber()

	err := decoder.Decode(&v)
	if err != nil {
		return string(rawJSON)
	}

	err = decoder.Decode(new(any))
	if !errors.Is(err, io.EOF) {
		return string(rawJSON)
	}

	normalized, err := json.Marshal(v)
	if err != nil {
		return string(rawJSON)
	}

	return string(normalized)
}

func NewNormalizedJSONTypesNormalizedValue(bytes []byte) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(NormalizeJSON(bytes))
}
