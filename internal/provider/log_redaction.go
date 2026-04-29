package provider

import (
	"reflect"
	"strings"
)

const logRedactionReplacement = "<redacted>"

type LogRedactionRule struct {
	path      []string
	condition *logRedactionCondition
}

type logRedactionCondition struct {
	path  []string
	value any
}

func RedactPath(path string) LogRedactionRule {
	return LogRedactionRule{
		path: splitLogRedactionPath(path),
	}
}

func RedactPathWhen(path string, conditionPath string, conditionValue any) LogRedactionRule {
	return LogRedactionRule{
		path: splitLogRedactionPath(path),
		condition: &logRedactionCondition{
			path:  splitLogRedactionPath(conditionPath),
			value: conditionValue,
		},
	}
}

func RedactForLogging(value any, rules ...LogRedactionRule) any {
	if value == nil || len(rules) == 0 {
		return value
	}

	redacted, changed := redactLogValue(reflect.ValueOf(value), nil, reflect.Value{}, rules)
	if !changed {
		return value
	}

	return redacted.Interface()
}

func redactLogValue(value reflect.Value, path []string, parent reflect.Value, rules []LogRedactionRule) (reflect.Value, bool) {
	if !value.IsValid() {
		return value, false
	}

	if shouldRedactLogPath(path, parent, rules) {
		return logRedactionValue(value.Type()), true
	}

	if !logRedactionRulesCanMatchBelow(path, rules) {
		return value, false
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Interface:
		return redactLogInterfaceValue(value, path, parent, rules)

	case reflect.Pointer:
		return redactLogPointerValue(value, path, parent, rules)

	case reflect.Struct:
		return redactLogStructValue(value, path, rules)

	case reflect.Map:
		return redactLogMapValue(value, path, rules)

	case reflect.Slice:
		return redactLogSliceValue(value, path, rules)

	case reflect.Array:
		return redactLogArrayValue(value, path, rules)

	default:
		return value, false
	}
}

func redactLogInterfaceValue(value reflect.Value, path []string, parent reflect.Value, rules []LogRedactionRule) (reflect.Value, bool) {
	if value.IsNil() {
		return value, false
	}

	redacted, changed := redactLogValue(value.Elem(), path, parent, rules)
	if !changed {
		return value, false
	}

	replacement := reflect.New(value.Type()).Elem()
	replacement.Set(redacted)

	return replacement, true
}

func redactLogPointerValue(value reflect.Value, path []string, parent reflect.Value, rules []LogRedactionRule) (reflect.Value, bool) {
	if value.IsNil() {
		return value, false
	}

	redacted, changed := redactLogValue(value.Elem(), path, parent, rules)
	if !changed {
		return value, false
	}

	replacement := reflect.New(value.Type().Elem())
	replacement.Elem().Set(redacted)

	return replacement, true
}

func redactLogStructValue(value reflect.Value, path []string, rules []LogRedactionRule) (reflect.Value, bool) {
	replacement := reflect.Value{}
	changed := false

	for i := range value.NumField() {
		field := value.Type().Field(i)
		if !field.IsExported() {
			continue
		}

		fieldPath := appendLogRedactionPath(path, logRedactionFieldName(field))

		redactedField, fieldChanged := redactLogValue(value.Field(i), fieldPath, value, rules)
		if !fieldChanged {
			continue
		}

		if !changed {
			replacement = reflect.New(value.Type()).Elem()
			replacement.Set(value)

			changed = true
		}

		if replacement.Field(i).CanSet() {
			replacement.Field(i).Set(redactedField)
		}
	}

	if !changed {
		return value, false
	}

	return replacement, true
}

func redactLogMapValue(value reflect.Value, path []string, rules []LogRedactionRule) (reflect.Value, bool) {
	if value.IsNil() {
		return value, false
	}

	replacement := reflect.Value{}
	changed := false

	iter := value.MapRange()
	for iter.Next() {
		mapValue := iter.Value()
		childPath := appendLogRedactionPath(path, logRedactionMapKeyPathSegment(iter.Key()))

		redactedValue, valueChanged := redactLogValue(mapValue, childPath, value, rules)
		if !valueChanged {
			continue
		}

		if !changed {
			replacement = reflect.MakeMapWithSize(value.Type(), value.Len())
			copyLogMapValues(value, replacement)

			changed = true
		}

		replacement.SetMapIndex(iter.Key(), redactedValue)
	}

	if !changed {
		return value, false
	}

	return replacement, true
}

func redactLogSliceValue(value reflect.Value, path []string, rules []LogRedactionRule) (reflect.Value, bool) {
	if value.IsNil() {
		return value, false
	}

	replacement := reflect.Value{}
	changed := false

	for i := range value.Len() {
		childPath := appendLogRedactionPath(path, "*")

		redactedValue, valueChanged := redactLogValue(value.Index(i), childPath, value, rules)
		if !valueChanged {
			continue
		}

		if !changed {
			replacement = reflect.MakeSlice(value.Type(), value.Len(), value.Cap())
			reflect.Copy(replacement, value)

			changed = true
		}

		replacement.Index(i).Set(redactedValue)
	}

	if !changed {
		return value, false
	}

	return replacement, true
}

func redactLogArrayValue(value reflect.Value, path []string, rules []LogRedactionRule) (reflect.Value, bool) {
	replacement := reflect.Value{}
	changed := false

	for i := range value.Len() {
		childPath := appendLogRedactionPath(path, "*")

		redactedValue, valueChanged := redactLogValue(value.Index(i), childPath, value, rules)
		if !valueChanged {
			continue
		}

		if !changed {
			replacement = reflect.New(value.Type()).Elem()
			replacement.Set(value)

			changed = true
		}

		replacement.Index(i).Set(redactedValue)
	}

	if !changed {
		return value, false
	}

	return replacement, true
}

func copyLogMapValues(source reflect.Value, target reflect.Value) {
	iter := source.MapRange()
	for iter.Next() {
		target.SetMapIndex(iter.Key(), iter.Value())
	}
}

func logRedactionMapKeyPathSegment(key reflect.Value) string {
	if key.Kind() == reflect.String {
		return key.String()
	}

	return "*"
}

func shouldRedactLogPath(path []string, parent reflect.Value, rules []LogRedactionRule) bool {
	for _, rule := range rules {
		if !logRedactionPathMatches(rule.path, path) {
			continue
		}

		if rule.condition == nil {
			return true
		}

		if logRedactionConditionMatches(parent, *rule.condition) {
			return true
		}
	}

	return false
}

func logRedactionRulesCanMatchBelow(path []string, rules []LogRedactionRule) bool {
	for _, rule := range rules {
		if logRedactionPathHasPrefix(rule.path, path) {
			return true
		}
	}

	return false
}

func logRedactionPathMatches(rulePath []string, path []string) bool {
	if len(rulePath) != len(path) {
		return false
	}

	return logRedactionPathHasPrefix(rulePath, path)
}

func logRedactionPathHasPrefix(rulePath []string, path []string) bool {
	if len(path) > len(rulePath) {
		return false
	}

	for i, segment := range path {
		if rulePath[i] == "*" {
			continue
		}

		if normalizeLogRedactionPathSegment(rulePath[i]) != normalizeLogRedactionPathSegment(segment) {
			return false
		}
	}

	return true
}

func logRedactionConditionMatches(parent reflect.Value, condition logRedactionCondition) bool {
	value, valueFound := logRedactionValueAtPath(parent, condition.path)
	if !valueFound {
		return false
	}

	actual, valueComparable := logRedactionComparableValue(value)
	if !valueComparable {
		return false
	}

	return reflect.DeepEqual(actual, condition.value)
}

func logRedactionValueAtPath(value reflect.Value, path []string) (reflect.Value, bool) {
	current := value

	for _, segment := range path {
		current = unwrapLogRedactionValue(current)
		if !current.IsValid() {
			return reflect.Value{}, false
		}

		//nolint:exhaustive
		switch current.Kind() {
		case reflect.Struct:
			field, ok := logRedactionStructField(current, segment)
			if !ok {
				return reflect.Value{}, false
			}

			current = field

		case reflect.Map:
			mapValue, ok := logRedactionMapValue(current, segment)
			if !ok {
				return reflect.Value{}, false
			}

			current = mapValue

		default:
			return reflect.Value{}, false
		}
	}

	return current, true
}

func logRedactionStructField(value reflect.Value, name string) (reflect.Value, bool) {
	valueType := value.Type()

	for i := range value.NumField() {
		field := valueType.Field(i)
		if !field.IsExported() {
			continue
		}

		if normalizeLogRedactionPathSegment(logRedactionFieldName(field)) == normalizeLogRedactionPathSegment(name) {
			return value.Field(i), true
		}
	}

	return reflect.Value{}, false
}

func logRedactionMapValue(value reflect.Value, key string) (reflect.Value, bool) {
	if value.Type().Key().Kind() != reflect.String {
		return reflect.Value{}, false
	}

	mapValue := value.MapIndex(reflect.ValueOf(key).Convert(value.Type().Key()))
	if mapValue.IsValid() {
		return mapValue, true
	}

	iter := value.MapRange()
	for iter.Next() {
		if normalizeLogRedactionPathSegment(iter.Key().String()) == normalizeLogRedactionPathSegment(key) {
			return iter.Value(), true
		}
	}

	return reflect.Value{}, false
}

func logRedactionComparableValue(value reflect.Value) (any, bool) {
	value = unwrapLogRedactionValue(value)
	if !value.IsValid() {
		return nil, false
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), true

	case reflect.String:
		return value.String(), true

	case reflect.Struct:
		if field, ok := logRedactionStructField(value, "value"); ok {
			return logRedactionComparableValue(field)
		}
	}

	return nil, false
}

func unwrapLogRedactionValue(value reflect.Value) reflect.Value {
	for value.IsValid() {
		//nolint:exhaustive
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if value.IsNil() {
				return reflect.Value{}
			}

			value = value.Elem()

		default:
			return value
		}
	}

	return value
}

func logRedactionValue(valueType reflect.Type) reflect.Value {
	//nolint:exhaustive
	switch valueType.Kind() {
	case reflect.String:
		return reflect.ValueOf(logRedactionReplacement).Convert(valueType)

	case reflect.Pointer:
		replacement := reflect.New(valueType.Elem())
		replacement.Elem().Set(logRedactionValue(valueType.Elem()))

		return replacement

	case reflect.Interface:
		replacement := reflect.New(valueType).Elem()
		replacement.Set(reflect.ValueOf(logRedactionReplacement))

		return replacement

	case reflect.Struct:
		replacement := reflect.New(valueType).Elem()

		if valueField, ok := logRedactionSettableStructField(replacement, "value"); ok && valueField.Kind() == reflect.String {
			valueField.SetString(logRedactionReplacement)
		}

		if setField, ok := logRedactionSettableStructField(replacement, "set"); ok && setField.Kind() == reflect.Bool {
			setField.SetBool(true)
		}

		if nullField, ok := logRedactionSettableStructField(replacement, "null"); ok && nullField.Kind() == reflect.Bool {
			nullField.SetBool(false)
		}

		return replacement

	default:
		return reflect.Zero(valueType)
	}
}

func logRedactionSettableStructField(value reflect.Value, name string) (reflect.Value, bool) {
	field, ok := logRedactionStructField(value, name)
	if !ok || !field.CanSet() {
		return reflect.Value{}, false
	}

	return field, true
}

func logRedactionFieldName(field reflect.StructField) string {
	for _, tagName := range []string{"json", "tfsdk"} {
		tag := field.Tag.Get(tagName)
		if tag == "" {
			continue
		}

		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}

	return field.Name
}

func splitLogRedactionPath(path string) []string {
	rawSegments := strings.Split(path, ".")
	segments := make([]string, 0, len(rawSegments))

	for _, segment := range rawSegments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		segments = append(segments, segment)
	}

	return segments
}

func appendLogRedactionPath(path []string, segment string) []string {
	next := make([]string, 0, len(path)+1)
	next = append(next, path...)
	next = append(next, segment)

	return next
}

func normalizeLogRedactionPathSegment(segment string) string {
	segment = strings.ReplaceAll(segment, "_", "")
	segment = strings.ReplaceAll(segment, "-", "")

	return strings.ToLower(segment)
}
