package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewContentTypeResourceModelFromResponse(ctx context.Context, contentType cm.ContentType) (ContentTypeModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := contentType.Sys.Space.Sys.ID
	environmentID := contentType.Sys.Environment.Sys.ID
	contentTypeID := contentType.Sys.ID

	model := ContentTypeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, contentTypeID),
		ContentTypeIdentityModel: ContentTypeIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			ContentTypeID: types.StringValue(contentTypeID),
		},
	}

	model.Name = types.StringValue(contentType.Name)
	model.Description = types.StringValue(contentType.Description.Or(""))

	model.DisplayField = types.StringValue(contentType.DisplayField.Or(""))

	fieldsList, fieldsListDiags := NewFieldsListFromResponse(ctx, path.Root("fields"), contentType.Fields)
	diags.Append(fieldsListDiags...)

	model.Fields = fieldsList

	metadata, metadataDiags := NewContentTypeMetadataFromResponse(ctx, path.Root("metadata"), contentType.Metadata)
	diags.Append(metadataDiags...)

	model.Metadata = metadata
	model.Timeouts = TimeoutsNull()

	return model, diags
}

// NewContentTypeResourceModelFromMutationResponse preserves configuration-owned
// metadata children after a lossy response projection. metadata and taxonomy
// are Optional+Computed, so null and unknown values are response-owned.
// annotations is Optional-only, so every known value, including null, remains
// configuration-owned. This merge never writes a nested unknown into state.
func NewContentTypeResourceModelFromMutationResponse(ctx context.Context, contentType cm.ContentType, plan ContentTypeModel) (ContentTypeModel, diag.Diagnostics) {
	model, diags := NewContentTypeResourceModelFromResponse(ctx, contentType)

	planMetadata, planMetadataOk := plan.Metadata.GetValue()
	if !planMetadataOk {
		return model, diags
	}

	responseMetadata, responseMetadataOk := model.Metadata.GetValue()
	if !responseMetadataOk {
		responseMetadata = ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		}
	}

	if !planMetadata.Annotations.IsUnknown() {
		responseMetadata.Annotations = planMetadata.Annotations
	}

	if !planMetadata.Taxonomy.IsNull() && !planMetadata.Taxonomy.IsUnknown() {
		responseMetadata.Taxonomy = planMetadata.Taxonomy
	}

	model.Metadata = NewTypedObject(responseMetadata)

	return model, diags
}
