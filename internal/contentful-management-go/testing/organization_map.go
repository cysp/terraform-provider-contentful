package testing

type OrganizationMap[Value any] struct {
	m map[string]map[string]Value
}

func NewOrganizationMap[Value any]() OrganizationMap[Value] {
	return OrganizationMap[Value]{
		m: make(map[string]map[string]Value),
	}
}

//nolint:ireturn
func (sm *OrganizationMap[Value]) Get(organizationID string, key string) Value {
	organizationValues, exists := sm.m[organizationID]
	if !exists {
		var zeroValue Value

		return zeroValue
	}

	value := organizationValues[key]

	return value
}

func (sm *OrganizationMap[Value]) Set(organizationID string, key string, value Value) {
	if sm.m[organizationID] == nil {
		sm.m[organizationID] = make(map[string]Value)
	}

	sm.m[organizationID][key] = value
}

func (sm *OrganizationMap[Value]) Delete(organizationID string, key string) {
	organizationValues, exists := sm.m[organizationID]
	if !exists {
		return
	}

	delete(organizationValues, key)
}
