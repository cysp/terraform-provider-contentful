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

// NewContentTypeResourceModelForMutationState starts with the response
// projection and restores known plan-owned metadata children. Known Optional
// annotations, including null, are restored. Optional+Computed taxonomy is
// restored only when known and non-null; the response supplies omitted taxonomy
// and resolves unknown children so none enter state. Read skips reconciliation.
func NewContentTypeResourceModelForMutationState(ctx context.Context, contentType cm.ContentType, appliedPlan ContentTypeModel) (ContentTypeModel, diag.Diagnostics) {
	mutationState, diags := NewContentTypeResourceModelFromResponse(ctx, contentType)

	plannedMetadata, plannedMetadataIsKnown := appliedPlan.Metadata.GetValue()
	if !plannedMetadataIsKnown {
		return mutationState, diags
	}

	stateMetadata, stateMetadataIsKnown := mutationState.Metadata.GetValue()
	if !stateMetadataIsKnown {
		stateMetadata = ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		}
	}

	if !plannedMetadata.Annotations.IsUnknown() {
		stateMetadata.Annotations = plannedMetadata.Annotations
	}

	if !plannedMetadata.Taxonomy.IsNull() && !plannedMetadata.Taxonomy.IsUnknown() {
		stateMetadata.Taxonomy = plannedMetadata.Taxonomy
	}

	mutationState.Metadata = NewTypedObject(stateMetadata)

	return mutationState, diags
}
