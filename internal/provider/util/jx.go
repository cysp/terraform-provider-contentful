package util

import (
	"sort"

	"github.com/go-faster/jx"
)

func EncodeJxRawMapOrdered(encoder *jx.Encoder, object map[string]jx.Raw) {
	encoder.ObjStart()

	keys := make([]string, 0, len(object))
	for k := range object {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		encoder.FieldStart(k)

		elem := object[k]

		if len(elem) != 0 {
			encoder.Raw(elem)
		}
	}

	encoder.ObjEnd()
}
