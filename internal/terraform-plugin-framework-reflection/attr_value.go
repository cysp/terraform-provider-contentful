package terraformpluginframeworkreflection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type AttrValueWithObjectAttrTypes interface {
	attr.Value

	ObjectAttrTypes(ctx context.Context) map[string]attr.Type
}
