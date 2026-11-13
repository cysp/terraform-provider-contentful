package provider_test

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

// Compare opaque JSON independently of provider normalization. UseNumber keeps
// numeric tokens exact; only object representation and string escaping change.
// Nil remains absent, distinct from JSON null and empty objects or arrays.
func canonicalModelJSON(t *testing.T, value jx.Raw) jx.Raw {
	t.Helper()

	if value == nil {
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()

	var decoded any

	require.NoError(t, decoder.Decode(&decoded))

	var trailing any

	require.ErrorIs(t, decoder.Decode(&trailing), io.EOF)

	encoded, err := json.Marshal(decoded)
	require.NoError(t, err)

	return encoded
}

func canonicalAppDefinitionParameters(t *testing.T, parameters cm.OptAppDefinitionParameters) cm.OptAppDefinitionParameters {
	t.Helper()

	canonicalParameters := func(values []cm.AppDefinitionParameter) []cm.AppDefinitionParameter {
		values = slices.Clone(values)
		for index := range values {
			values[index].Default = canonicalModelJSON(t, values[index].Default)
			values[index].Options = slices.Clone(values[index].Options)

			for optionIndex := range values[index].Options {
				values[index].Options[optionIndex] = canonicalModelJSON(t, values[index].Options[optionIndex])
			}
		}

		return values
	}
	parameters.Value.Installation = canonicalParameters(parameters.Value.Installation)
	parameters.Value.Instance = canonicalParameters(parameters.Value.Instance)

	return parameters
}
