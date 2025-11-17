package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (model *TagModel) ToTagData(_ context.Context) (cm.TagData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	tagData := cm.TagData{
		Name: model.Name.ValueString(),
	}

	return tagData, diags
}
