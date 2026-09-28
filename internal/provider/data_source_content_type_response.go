package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newContentTypeDataSourceItem(ctx context.Context, root path.Path, entity cm.ContentType, spaceID, environmentID string) (ContentTypeDataSourceItemModel, diag.Diagnostics) {
	var item ContentTypeDataSourceItemModel

	if entity.Sys.Space.Sys.ID != spaceID {
		return item, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(root.AtName("content_type_id"), "Unexpected response identity", "Content Type "+entity.Sys.ID+" returned a Space link outside the requested scope.")}
	}

	if entity.Sys.Environment.Sys.ID != environmentID {
		return item, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(root.AtName("content_type_id"), "Unsupported Content Type response identity", "Content Type "+entity.Sys.ID+" returned an environment link that does not echo the requested addressing context. This response cannot establish alias routing.")}
	}

	converted, diagnostics := newContentTypeResourceModelFromResponseAt(ctx, root, entity)
	if diagnostics.HasError() {
		return item, diagnostics
	}

	item = ContentTypeDataSourceItemModel{
		ContentTypeID:    types.StringValue(entity.Sys.ID),
		Name:             converted.Name,
		Description:      types.StringPointerValue(entity.Description.ValueStringPointer()),
		DisplayField:     types.StringPointerValue(entity.DisplayField.ValueStringPointer()),
		PublishedVersion: converted.PublishedVersion,
		Fields:           converted.Fields,
		Metadata:         converted.Metadata,
	}

	return item, diagnostics
}

func readContentTypes(ctx context.Context, client *cm.Client, spaceID, environmentID string) ([]ContentTypeDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read content types"

	// Project after paging so nested diagnostics use returned list indexes.
	entities, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.ContentType], diag.Diagnostics) {
			response, err := client.GetContentTypes(ctx, cm.GetContentTypesParams{
				SpaceID:       spaceID,
				EnvironmentID: environmentID,
				Skip:          cm.NewOptInt64(skip),
				Limit:         cm.NewOptInt64(defaultPageLimit),
			})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.ContentTypeCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.ContentType) (cm.ContentType, diag.Diagnostics) { return entity, nil },
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	items, itemDiagnostics := projectContentTypeDataSourceItems(ctx, entities, spaceID, environmentID)
	diagnostics.Append(itemDiagnostics...)

	return items, diagnostics
}

func projectContentTypeDataSourceItems(ctx context.Context, entities []cm.ContentType, spaceID, environmentID string) ([]ContentTypeDataSourceItemModel, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}

	items := make([]ContentTypeDataSourceItemModel, 0, len(entities))
	for index, entity := range entities {
		item, itemDiagnostics := newContentTypeDataSourceItem(ctx, path.Root("content_types").AtListIndex(index), entity, spaceID, environmentID)
		diagnostics.Append(itemDiagnostics...)

		if diagnostics.HasError() {
			return nil, diagnostics
		}

		items = append(items, item)
	}

	return items, diagnostics
}
