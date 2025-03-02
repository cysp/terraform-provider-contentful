package contentfulmanagementtestserver

type SpaceEnvironmentMap[Value any] struct {
	m map[string]map[string]map[string]Value
}

func NewSpaceEnvironmentMap[Value any]() SpaceEnvironmentMap[Value] {
	return SpaceEnvironmentMap[Value]{
		m: make(map[string]map[string]map[string]Value),
	}
}

func (sm *SpaceEnvironmentMap[Value]) ListSpace(spaceID string) map[string]map[string]Value {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return nil
	}

	return spaceValues
}

func (sm *SpaceEnvironmentMap[Value]) ListSpaceEnvironment(spaceID string, environmentID string) map[string]Value {
	spaceValues := sm.ListSpace(spaceID)
	if spaceValues == nil {
		return nil
	}

	environmentValues, exists := spaceValues[environmentID]
	if !exists {
		return nil
	}

	return environmentValues
}

//nolint:ireturn
func (sm *SpaceEnvironmentMap[Value]) Get(spaceID string, environmentID string, key string) (Value, bool) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		var zeroValue Value

		return zeroValue, false
	}

	environmentValues, exists := spaceValues[environmentID]
	if !exists {
		var zeroValue Value

		return zeroValue, false
	}

	value, exists := environmentValues[key]

	return value, exists
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

func (sm *SpaceEnvironmentMap[Value]) DeleteSpace(spaceID string) {
	delete(sm.m, spaceID)
}

func (sm *SpaceEnvironmentMap[Value]) DeleteSpaceEnvironment(spaceID string, environmentID string) {
	spaceValues, exists := sm.m[spaceID]
	if !exists {
		return
	}

	delete(spaceValues, environmentID)
}

func (sm *SpaceEnvironmentMap[Value]) Clear() {
	sm.m = make(map[string]map[string]map[string]Value)
}
