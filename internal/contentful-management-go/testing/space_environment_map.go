package testing

type SpaceEnvironmentMap[Value any] struct {
	m map[string]map[string]map[string]Value
}

func NewSpaceEnvironmentMap[Value any]() SpaceEnvironmentMap[Value] {
	return SpaceEnvironmentMap[Value]{
		m: make(map[string]map[string]map[string]Value),
	}
}

func (sm *SpaceEnvironmentMap[Value]) List(spaceID string, environmentID string) []Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		var zeroValue []Value

		return zeroValue
	}

	environmentValues, exists := spaceValues[environmentID]
	if !exists {
		var zeroValue []Value

		return zeroValue
	}

	value := make([]Value, 0, len(environmentValues))
	for _, v := range environmentValues {
		value = append(value, v)
	}

	return value
}

//nolint:ireturn
func (sm *SpaceEnvironmentMap[Value]) Get(spaceID string, environmentID string, key string) Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		var zeroValue Value

		return zeroValue
	}

	environmentValues, exists := spaceValues[environmentID]
	if !exists {
		var zeroValue Value

		return zeroValue
	}

	value := environmentValues[key]

	return value
}

func (sm *SpaceEnvironmentMap[Value]) Set(spaceID string, environmentID string, key string, value Value) {
	if sm.m[spaceID] == nil {
		sm.m[spaceID] = make(map[string]map[string]Value)
	}

	if sm.m[spaceID][environmentID] == nil {
		sm.m[spaceID][environmentID] = make(map[string]Value)
	}

	sm.m[spaceID][environmentID][key] = value
}

func (sm *SpaceEnvironmentMap[Value]) Delete(spaceID string, environmentID string, key string) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return
	}

	environmentValues, exists := spaceValues[environmentID]
	if !exists {
		return
	}

	delete(environmentValues, key)
}
