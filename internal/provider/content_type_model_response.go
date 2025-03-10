package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *ContentTypeModel) ReadFromResponse(ctx context.Context, contentType *cm.ContentType) diag.Diagnostics {
	diags := diag.Diagnostics{}

	spaceID := contentType.Sys.Space.Sys.ID
	environmentID := contentType.Sys.Environment.Sys.ID
	contentTypeID := contentType.Sys.ID

	m.ID = types.StringValue(strings.Join([]string{spaceID, environmentID, contentTypeID}, "/"))
	m.SpaceID = types.StringValue(spaceID)
	m.EnvironmentID = types.StringValue(environmentID)
	m.ContentTypeID = types.StringValue(contentTypeID)

	m.Name = types.StringValue(contentType.Name)
	m.Description = types.StringValue(contentType.Description.Or(""))
	m.DisplayField = types.StringValue(contentType.DisplayField.Or(""))

	fieldsList, fieldsListDiags := NewFieldsListFromResponse(ctx, path.Root("fields"), contentType.Fields)
	diags.Append(fieldsListDiags...)

	m.Fields = fieldsList

	return diags
}
