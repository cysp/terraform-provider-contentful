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

		dec := jx.DecodeBytes(elem)
		if str, err := dec.Str(); err == nil {
			encoder.StrEscape(str)
		} else {
			encoder.Raw(elem)
		}
	}

	encoder.ObjEnd()
}
