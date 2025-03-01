package contentfulmanagementtestserver

type SpacedMap[Value any] struct {
	m map[string]map[string]Value
}

func NewSpacedMap[Value any]() SpacedMap[Value] {
	return SpacedMap[Value]{
		m: make(map[string]map[string]Value),
	}
}

func (sm *SpacedMap[Value]) List(spaceID string) map[string]Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return nil
	}

	return spaceValues
}

//nolint:ireturn
func (sm *SpacedMap[Value]) Get(spaceID string, key string) (Value, bool) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		var zeroValue Value

		return zeroValue, false
	}

	value, exists := spaceValues[key]
	if !exists {
		var zeroValue Value

		return zeroValue, false
	}

	return value, true
}

func (sm *SpacedMap[Value]) Set(spaceID string, key string, value Value) {
	if sm.m[spaceID] == nil {
		sm.m[spaceID] = make(map[string]Value)
	}

	sm.m[spaceID][key] = value
}

func (sm *SpacedMap[Value]) Delete(spaceID string, key string) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return
	}

	delete(spaceValues, key)
}

func (sm *SpacedMap[Value]) DeleteSpace(spaceID string) {
	delete(sm.m, spaceID)
}

func (sm *SpacedMap[Value]) Clear() {
	sm.m = make(map[string]map[string]Value)
}
