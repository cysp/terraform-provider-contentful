package contentfulmanagementtestserver

type SpaceMap[Value any] struct {
	m map[string]map[string]Value
}

func NewSpaceMap[Value any]() SpaceMap[Value] {
	return SpaceMap[Value]{
		m: make(map[string]map[string]Value),
	}
}

func (sm *SpaceMap[Value]) List(spaceID string) map[string]Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return nil
	}

	return spaceValues
}

//nolint:ireturn
func (sm *SpaceMap[Value]) Get(spaceID string, key string) (Value, bool) {
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

func (sm *SpaceMap[Value]) DeleteSpace(spaceID string) {
	delete(sm.m, spaceID)
}

func (sm *SpaceMap[Value]) Clear() {
	sm.m = make(map[string]map[string]Value)
}
