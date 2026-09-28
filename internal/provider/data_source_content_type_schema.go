package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func contentTypeDataSourceAllowedResourceAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"contentful_entry": schema.SingleNestedAttribute{
			Description: "Allowed Contentful entry resource, or null for an external resource.",
			Computed:    true,
			CustomType:  NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue]().CustomType(ctx),
			Attributes: map[string]schema.Attribute{
				"source":        schema.StringAttribute{Description: "Source of the allowed Contentful entries.", Computed: true},
				"content_types": schema.ListAttribute{Description: "Allowed content type IDs.", Computed: true, ElementType: types.StringType, CustomType: NewTypedListNull[types.String]().CustomType(ctx)},
			},
		},
		"external": schema.SingleNestedAttribute{
			Description: "Allowed external resource, or null for Contentful entries.",
			Computed:    true,
			CustomType:  NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue]().CustomType(ctx),
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{Description: "External resource type.", Computed: true},
			},
		},
	}
}

func contentTypeDataSourceFieldAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":            schema.StringAttribute{Description: "ID of the field within the content type.", Computed: true},
		"name":          schema.StringAttribute{Description: "Name of the field.", Computed: true},
		"type":          schema.StringAttribute{Description: "Contentful field type.", Computed: true},
		"link_type":     schema.StringAttribute{Description: "Linked resource type for a Link field, or null when absent.", Computed: true},
		"disabled":      schema.BoolAttribute{Description: "Whether the field is hidden in the entry editor, or null when absent.", Computed: true},
		"omitted":       schema.BoolAttribute{Description: "Whether the field is omitted from Delivery and Preview responses, or null when absent.", Computed: true},
		"required":      schema.BoolAttribute{Description: "Whether an Entry needs a value before publication, or null when absent.", Computed: true},
		"localized":     schema.BoolAttribute{Description: "Whether the field supports localized values, or null when absent.", Computed: true},
		"default_value": schema.StringAttribute{Description: "Normalized JSON object of locale defaults, or null when absent.", Computed: true, CustomType: jsontypes.NormalizedType{}},
		"items": schema.SingleNestedAttribute{
			Description: "Array item type and validations, or null for other fields.",
			Computed:    true,
			CustomType:  NewTypedObjectNull[ContentTypeFieldItemsValue]().CustomType(ctx),
			Attributes: map[string]schema.Attribute{
				"type":        schema.StringAttribute{Description: "Array item type, or null when absent.", Computed: true},
				"link_type":   schema.StringAttribute{Description: "Linked resource type for Link items, or null when absent.", Computed: true},
				"validations": schema.ListAttribute{Description: "Ordered array-item validation rules as normalized JSON strings.", Computed: true, ElementType: jsontypes.NormalizedType{}, CustomType: NewTypedListNull[jsontypes.Normalized]().CustomType(ctx)},
			},
		},
		"validations": schema.ListAttribute{
			Description: "Ordered field validation rules as normalized JSON strings.",
			Computed:    true,
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
		},
		"allowed_resources": schema.ListNestedAttribute{
			Description: "Allowed resources for a Resource Link field, or null when absent.",
			Computed:    true,
			CustomType:  NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]]().CustomType(ctx),
			NestedObject: schema.NestedAttributeObject{
				CustomType: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue]().CustomType(ctx),
				Attributes: contentTypeDataSourceAllowedResourceAttributes(ctx),
			},
		},
	}
}

func contentTypeDataSourceTaxonomyAttributes(ctx context.Context) map[string]schema.Attribute {
	concept := func(description string, customType basetypes.ObjectTypable) schema.SingleNestedAttribute {
		return schema.SingleNestedAttribute{
			Description: description,
			Computed:    true,
			CustomType:  customType,
			Attributes: map[string]schema.Attribute{
				"id":       schema.StringAttribute{Description: "Taxonomy ID.", Computed: true},
				"required": schema.BoolAttribute{Description: "Whether this association is required, or null when absent.", Computed: true},
			},
		}
	}

	return map[string]schema.Attribute{
		"taxonomy_concept":        concept("Taxonomy concept link, or null for a concept scheme.", NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue]().CustomType(ctx)),
		"taxonomy_concept_scheme": concept("Taxonomy concept scheme link, or null for a concept.", NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue]().CustomType(ctx)),
	}
}

func contentTypeDataSourceItemAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"content_type_id": schema.StringAttribute{Description: "System ID of the content type.", Computed: true},
		"name":            schema.StringAttribute{Description: "Current CMA name of the content type.", Computed: true},
		"description":     schema.StringAttribute{Description: "Current CMA description; null when omitted, and empty when explicitly empty.", Computed: true},
		"display_field":   schema.StringAttribute{Description: "Current CMA display field ID; null when explicitly null, and empty when explicitly empty. The curated decoder rejects an omitted displayField.", Computed: true},
		"published_version": schema.Int64Attribute{
			Description: "Most recently activated Contentful version, or null when never activated. This does not prove the returned fields match the activated version.",
			Computed:    true,
		},
		"fields": schema.ListNestedAttribute{
			Description: "Current CMA fields in returned order, including unactivated draft changes.",
			Computed:    true,
			CustomType:  NewTypedListUnknown[TypedObject[ContentTypeFieldValue]]().CustomType(ctx),
			NestedObject: schema.NestedAttributeObject{
				CustomType: NewTypedObjectNull[ContentTypeFieldValue]().CustomType(ctx),
				Attributes: contentTypeDataSourceFieldAttributes(ctx),
			},
		},
		"metadata": schema.SingleNestedAttribute{
			Description: "Current CMA annotations and taxonomy, or null when metadata is absent.",
			Computed:    true,
			CustomType:  NewTypedObjectNull[ContentTypeMetadataValue]().CustomType(ctx),
			Attributes: map[string]schema.Attribute{
				"annotations": schema.StringAttribute{Description: "Normalized JSON annotations, or null when absent.", Computed: true, CustomType: jsontypes.NormalizedType{}},
				"taxonomy": schema.ListNestedAttribute{
					Description: "Taxonomy concept and concept scheme links in returned order, or null when absent.",
					Computed:    true,
					CustomType:  NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]]().CustomType(ctx),
					NestedObject: schema.NestedAttributeObject{
						CustomType: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue]().CustomType(ctx),
						Attributes: contentTypeDataSourceTaxonomyAttributes(ctx),
					},
				},
			},
		},
	}
}

func ContentTypeDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := contentTypeDataSourceItemAttributes(ctx)
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the Space used to address the content type.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["environment_id"] = schema.StringAttribute{Description: "ID of the environment or environment alias used to address the content type. A returned environment link that differs from this ID is rejected as unsupported response identity.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["content_type_id"] = schema.StringAttribute{Description: "ID of the content type to retrieve.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform lookup ID in `space_id/environment_id/content_type_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Reads a Contentful Content Type's current CMA model without activating or changing it. Alias routes require the response environment link to echo the requested alias ID; this Content Type behavior has not been verified live.", Attributes: attributes}
}

func ContentTypesDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Reads all Contentful Content Types in an environment from the current CMA model, including unactivated draft changes. No snapshot is promised across pages. Alias routes require returned environment links to echo the requested alias ID; this Content Type behavior has not been verified live.",
		Attributes: map[string]schema.Attribute{
			"space_id":       schema.StringAttribute{Description: "ID of the Space used to address the collection.", Required: true, Validators: []validator.String{discoveryIDValidator{}}},
			"environment_id": schema.StringAttribute{Description: "ID of the environment or environment alias used to address the collection. Mismatched returned environment links are rejected as unsupported response identity.", Required: true, Validators: []validator.String{discoveryIDValidator{}}},
			"id":             schema.StringAttribute{Description: "Composite Terraform lookup ID in `space_id/environment_id` form.", Computed: true},
			"timeouts":       timeouts.Attributes(ctx),
			"content_types": schema.ListNestedAttribute{
				Description: "Current CMA content types in returned order. Duplicate IDs are retained; an empty collection returns an empty list. Order is not promised across reads.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: contentTypeDataSourceItemAttributes(ctx),
				},
			},
		},
	}
}
