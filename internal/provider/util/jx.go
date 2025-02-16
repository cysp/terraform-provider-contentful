package util

import (
	"errors"
	"slices"

	"github.com/go-faster/jx"
)

var errInvalid = errors.New("invalid")

func JxNormalizeOpaqueBytes(bytes []byte, options JxEncodeOpaqueOptions) ([]byte, error) {
	decoder := jx.DecodeBytes(bytes)

	value, err := JxDecodeOpaque(decoder)
	if err != nil {
		return []byte{}, err
	}

	encoder := jx.Encoder{}

	err = JxEncodeOpaqueOrdered(&encoder, value, options)

	return encoder.Bytes(), err
}

//nolint:wrapcheck
func JxDecodeOpaque(decoder *jx.Decoder) (interface{}, error) {
	switch decoder.Next() {
	case jx.Object:
		value := make(map[string]interface{})
		err := decoder.ObjBytes(func(d *jx.Decoder, key []byte) error {
			element, err := JxDecodeOpaque(d)
			if err == nil {
				value[string(key)] = element
			}

			return err
		})

		return value, err

	case jx.Array:
		value := make([]interface{}, 0)
		err := decoder.Arr(func(d *jx.Decoder) error {
			element, err := JxDecodeOpaque(d)
			if err == nil {
				value = append(value, element)
			}

			return err
		})

		return value, err

	case jx.String:
		return decoder.Str()

	case jx.Number:
		return decoder.Float64()

	case jx.Bool:
		return decoder.Bool()

	case jx.Null:
		return nil, decoder.Null()

	case jx.Invalid:
		fallthrough
	default:
		return nil, errInvalid
	}
}

type JxEncodeOpaqueOptions struct {
	EscapeStrings bool
}

func JxEncodeOpaqueOrdered(encoder *jx.Encoder, value interface{}, options JxEncodeOpaqueOptions) error {
	switch value := value.(type) {
	case map[string]interface{}:
		return jxEncodeOpaqueOrderedObject(encoder, value, options)

	case []interface{}:
		return jxEncodeOpaqueOrderedArray(encoder, value, options)

	case string:
		return jxEncodeOpaqueOrderedString(encoder, value, options)

	case int:
		encoder.Int(value)

		return nil

	case int8:
		encoder.Int8(value)

		return nil

	case int16:
		encoder.Int16(value)

		return nil

	case int32:
		encoder.Int32(value)

		return nil

	case int64:
		encoder.Int64(value)

		return nil

	case uint:
		encoder.UInt(value)

		return nil

	case uint8:
		encoder.UInt8(value)

		return nil

	case uint16:
		encoder.UInt16(value)

		return nil

	case uint32:
		encoder.UInt32(value)

		return nil

	case uint64:
		encoder.UInt64(value)

		return nil

	case float32:
		encoder.Float32(value)

		return nil

	case float64:
		encoder.Float64(value)

		return nil

	case bool:
		encoder.Bool(value)

		return nil

	case nil:
		encoder.Null()

		return nil

	default:
		return errInvalid
	}
}

func jxEncodeOpaqueOrderedObject(encoder *jx.Encoder, value map[string]interface{}, options JxEncodeOpaqueOptions) error {
	encoder.ObjStart()

	keys := make([]string, 0, len(value))
	for k := range value {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, key := range keys {
		element := value[key]

		encoder.FieldStart(key)

		if err := JxEncodeOpaqueOrdered(encoder, element, options); err != nil {
			return err
		}
	}

	encoder.ObjEnd()

	return nil
}

func jxEncodeOpaqueOrderedArray(encoder *jx.Encoder, value []interface{}, options JxEncodeOpaqueOptions) error {
	encoder.ArrStart()

	for _, element := range value {
		if err := JxEncodeOpaqueOrdered(encoder, element, options); err != nil {
			return err
		}
	}

	encoder.ArrEnd()

	return nil
}

func jxEncodeOpaqueOrderedString(encoder *jx.Encoder, value string, options JxEncodeOpaqueOptions) error {
	if options.EscapeStrings {
		encoder.StrEscape(value)
	} else {
		encoder.Str(value)
	}

	return nil
}
