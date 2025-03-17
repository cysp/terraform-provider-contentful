package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func AttributesFromTerraformValue(ctx context.Context, attrTypes map[string]attr.Type, value tftypes.Value) (map[string]attr.Value, error) {
	attributes := map[string]attr.Value{}

	tfvals := map[string]tftypes.Value{}

	err := value.As(&tfvals)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	for key, tfval := range tfvals {
		a, err := attrTypes[key].ValueFromTerraform(ctx, tfval)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		attributes[key] = a
	}

	return attributes, nil
}

type UnexpectedTerraformTypeError struct {
	Expected tftypes.Type
	Actual   tftypes.Type
}

func (e UnexpectedTerraformTypeError) Error() string {
	return fmt.Sprintf("expected %s, actual %s", e.Expected, e.Actual)
}
