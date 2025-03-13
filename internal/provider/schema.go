package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ObjectAttrTypesFromSchemaAttributes(_ context.Context, attributes map[string]schema.Attribute) map[string]attr.Type {
	attrTypes := make(map[string]attr.Type, len(attributes))

	for name, attribute := range attributes {
		attrTypes[name] = attribute.GetType()
	}

	return attrTypes
}
