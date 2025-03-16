package terraformpluginframeworkreflection

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func ValuesEqual[T attr.Value](a, b attr.Value) bool {
	aValue, aValueOk := a.(T)
	bValue, bValueOk := b.(T)

	if !aValueOk || !bValueOk {
		return false
	}

	aIsUnknown := aValue.IsUnknown()
	bIsUnknown := bValue.IsUnknown()

	if aIsUnknown || bIsUnknown {
		return aIsUnknown == bIsUnknown
	}

	aIsNull := aValue.IsNull()
	bIsNull := bValue.IsNull()

	if aIsNull || bIsNull {
		return aIsNull == bIsNull
	}

	return ValueAttributesEqual(aValue, bValue)
}
