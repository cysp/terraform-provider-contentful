package contentfulmanagement

type SpaceMap[Value any] struct {
	m map[string]map[string]Value
}

func NewSpaceMap[Value any]() SpaceMap[Value] {
	return SpaceMap[Value]{
		m: make(map[string]map[string]Value),
	}
}

//nolint:ireturn
func (sm *SpaceMap[Value]) Get(spaceID string, key string) Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		var zeroValue Value

		return zeroValue
	}

	value, exists := spaceValues[key]
	if !exists {
		var zeroValue Value

		return zeroValue
	}

	return value
}

func (sm *SpaceMap[Value]) Set(spaceID string, key string, value Value) {
	if sm.m[spaceID] == nil {
		sm.m[spaceID] = make(map[string]Value)
	}

	sm.m[spaceID][key] = value
}

func (sm *SpaceMap[Value]) Delete(spaceID string, key string) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return
	}

	delete(spaceValues, key)
}
