package terraformpluginframeworkreflection

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type UnexpectedValueStateError struct {
	ValueState attr.ValueState
}

func (e UnexpectedValueStateError) Error() string {
	return fmt.Sprintf("unexpected value state: 0x%02x", uint8(e.ValueState))
}
