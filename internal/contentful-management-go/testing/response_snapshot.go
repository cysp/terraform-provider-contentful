package cmtesting

import (
	"reflect"
	"time"
)

// snapshotResponse detaches a generated response DTO from stored fixture data.
// Call while holding the store lock, after any response projection or redaction.
// DTO graphs are acyclic and contain exported fields; time.Time is immutable.
// Copying their Go representation preserves optional flags and adversarial raw
// values that a codec round trip would normalize or reject.
//
//nolint:ireturn // Preserve each endpoint response interface and concrete variant.
func snapshotResponse[T any](response T) T {
	var snapshot T

	value := reflect.ValueOf(response)
	if !value.IsValid() {
		return response
	}

	reflect.ValueOf(&snapshot).Elem().Set(snapshotResponseValue(value))

	return snapshot
}

func snapshotResponseValue(value reflect.Value) reflect.Value {
	switch value.Kind() { //nolint:exhaustive // Only generated DTO kinds are supported; reject all others.
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		result := reflect.New(value.Type().Elem())
		result.Elem().Set(snapshotResponseValue(value.Elem()))

		return result
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := range value.Len() {
			result.Index(index).Set(snapshotResponseValue(value.Index(index)))
		}

		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		if value.Type().Key().Kind() != reflect.String {
			panic("response DTO map keys must be strings")
		}

		result := reflect.MakeMapWithSize(value.Type(), value.Len())

		entries := value.MapRange()
		for entries.Next() {
			result.SetMapIndex(entries.Key(), snapshotResponseValue(entries.Value()))
		}

		return result
	case reflect.Struct:
		if value.Type() == reflect.TypeFor[time.Time]() {
			return value
		}

		result := reflect.New(value.Type()).Elem()
		for index := range value.NumField() {
			result.Field(index).Set(snapshotResponseValue(value.Field(index)))
		}

		return result
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return value
	default:
		panic("unsupported response DTO kind: " + value.Kind().String())
	}
}
