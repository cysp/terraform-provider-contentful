package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewContentTypeResourceModelFromResponse(ctx context.Context, contentType cm.ContentType) (ContentTypeResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := contentType.Sys.Space.Sys.ID
	environmentID := contentType.Sys.Environment.Sys.ID
	contentTypeID := contentType.Sys.ID

	model := ContentTypeResourceModel{
		ID:            types.StringValue(strings.Join([]string{spaceID, environmentID, contentTypeID}, "/")),
		SpaceID:       types.StringValue(spaceID),
		EnvironmentID: types.StringValue(environmentID),
		ContentTypeID: types.StringValue(contentTypeID),
	}

	model.Name = types.StringValue(contentType.Name)
	model.Description = types.StringValue(contentType.Description.Or(""))

	model.DisplayField = types.StringValue(contentType.DisplayField.Or(""))

	fieldsList, fieldsListDiags := NewFieldsListFromResponse(ctx, path.Root("fields"), contentType.Fields)
	diags.Append(fieldsListDiags...)

	model.Fields = fieldsList

	return model, diags
}
