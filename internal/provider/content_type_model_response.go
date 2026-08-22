package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	if publishedVersion, ok := contentType.Sys.PublishedVersion.Get(); ok {
		model.PublishedVersion = types.Int64Value(int64(publishedVersion))
	} else {
		model.PublishedVersion = types.Int64Null()
	}

	fieldsList, fieldsListDiags := NewFieldsListFromResponse(ctx, path.Root("fields"), contentType.Fields)
	diags.Append(fieldsListDiags...)

	model.Fields = fieldsList

	metadata, metadataDiags := NewContentTypeMetadataFromResponse(ctx, path.Root("metadata"), contentType.Metadata)
	diags.Append(metadataDiags...)

	model.Metadata = metadata
	model.Timeouts = TimeoutsNull()

	return model, diags
}

// ProjectContentTypeMutationResponse makes the CMA response the recovery state
// first. Exact plan values are restored only after every owned value has been
// proven equivalent; this prevents a successful but contradictory CMA response
// being hidden in Terraform state. Configuration remains separate from Plan so
// an omitted Optional+Computed taxonomy stays Contentful-owned even after plan
// reconciliation has made it known.
//
//nolint:contextcheck // attr.Value.Equal and TypedObject.Equal expose no context-aware alternative.
func ProjectContentTypeMutationResponse(ctx context.Context, contentType cm.ContentType, config, plan ContentTypeModel) (ContentTypeModel, diag.Diagnostics, diag.Diagnostics) {
	data, responseDiags := NewContentTypeResourceModelFromResponse(ctx, contentType)

	data.Timeouts = plan.Timeouts
	if responseDiags.HasError() {
		return data, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}
	mismatch := false

	for _, identity := range []struct {
		valuePath       path.Path
		planned, remote types.String
	}{
		{path.Root("space_id"), plan.SpaceID, data.SpaceID},
		{path.Root("environment_id"), plan.EnvironmentID, data.EnvironmentID},
		{path.Root("content_type_id"), plan.ContentTypeID, data.ContentTypeID},
	} {
		if !identity.planned.IsNull() && !identity.planned.IsUnknown() && !identity.planned.Equal(identity.remote) {
			consistencyDiags.AddAttributeError(identity.valuePath, "Unexpected Contentful content type response", "The response identity differed from the requested Contentful endpoint.")

			mismatch = true
		}
	}
	// The endpoint, rather than an accidental response body identity, is the
	// identity Terraform mutated and must remain usable for recovery.
	data.SpaceID, data.EnvironmentID, data.ContentTypeID = plan.SpaceID, plan.EnvironmentID, plan.ContentTypeID
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(plan.SpaceID.ValueString(), plan.EnvironmentID.ValueString(), plan.ContentTypeID.ValueString())

	var project []func()

	compare := func(name string, valuePath path.Path, planned, remote attr.Value, restore func()) {
		if contentTypeMutationValueEquivalent(name, valuePath, planned, remote, &consistencyDiags) {
			project = append(project, restore)

			return
		}

		mismatch = true
	}
	compare("name", path.Root("name"), plan.Name, data.Name, func() { data.Name = plan.Name })
	compare("description", path.Root("description"), plan.Description, data.Description, func() { data.Description = plan.Description })
	compare("display_field", path.Root("display_field"), plan.DisplayField, data.DisplayField, func() { data.DisplayField = plan.DisplayField })

	fieldsPath := path.Root("fields")
	if plan.Fields.IsUnknown() {
		consistencyDiags.AddAttributeError(fieldsPath, "Unexpected unknown value", "Terraform-owned content type fields must be known before Contentful can be changed.")

		mismatch = true
	} else if differencePath, equivalent := contentTypeFieldsEquivalentAt(fieldsPath, plan.Fields, data.Fields); equivalent {
		project = append(project, func() { data.Fields = plan.Fields })
	} else {
		consistencyDiags.AddAttributeError(differencePath, "Unexpected Contentful content type response", "The fields response differed meaningfully from the Terraform plan.")

		mismatch = true
	}

	metadataProjects, metadataMismatch := contentTypeMetadataReconciliations(config, plan, &data, &consistencyDiags)
	project = append(project, metadataProjects...)
	mismatch = mismatch || metadataMismatch

	if !mismatch {
		for _, restore := range project {
			restore()
		}
	}

	return data, responseDiags, consistencyDiags
}

func contentTypeMetadataReconciliations(config, plan ContentTypeModel, data *ContentTypeModel, diags *diag.Diagnostics) ([]func(), bool) {
	planMetadata, planKnown := plan.Metadata.GetValue()
	if !planKnown {
		return nil, false
	}

	configMetadata, configKnown := config.Metadata.GetValue()
	taxonomyOwned := config.Metadata.IsUnknown() || (configKnown && !configMetadata.Taxonomy.IsNull())
	dataMetadata, dataKnown := data.Metadata.GetValue()
	annotationsPath := path.Root("metadata").AtName("annotations")
	taxonomyPath := path.Root("metadata").AtName("taxonomy")

	if !dataKnown {
		annotationsEquivalent := planMetadata.Annotations.IsUnknown() || contentTypeNormalizedJSONEquivalent(planMetadata.Annotations, jsontypes.NewNormalizedNull())
		taxonomyEquivalent := !taxonomyOwned || planMetadata.Taxonomy.IsUnknown() || planMetadata.Taxonomy.IsNull()

		if !annotationsEquivalent {
			diags.AddAttributeError(annotationsPath, "Unexpected Contentful content type response", "The metadata annotations response was missing from Contentful.")
		}

		if !taxonomyEquivalent {
			diags.AddAttributeError(taxonomyPath, "Unexpected Contentful content type response", "The metadata taxonomy response was missing from Contentful.")
		}

		return nil, !annotationsEquivalent || !taxonomyEquivalent
	}

	projects := []func(){}
	mismatch := false

	if !planMetadata.Annotations.IsUnknown() {
		if contentTypeNormalizedJSONEquivalent(planMetadata.Annotations, dataMetadata.Annotations) {
			projects = append(projects, func() {
				metadata := data.Metadata.Value()
				metadata.Annotations = planMetadata.Annotations
				data.Metadata = NewTypedObject(metadata)
			})
		} else {
			diags.AddAttributeError(annotationsPath, "Unexpected Contentful content type response", "The metadata annotations response differed meaningfully from the Terraform plan.")

			mismatch = true
		}
	}

	if taxonomyOwned && !planMetadata.Taxonomy.IsUnknown() {
		if planMetadata.Taxonomy.Equal(dataMetadata.Taxonomy) {
			projects = append(projects, func() {
				metadata := data.Metadata.Value()
				metadata.Taxonomy = planMetadata.Taxonomy
				data.Metadata = NewTypedObject(metadata)
			})
		} else {
			diags.AddAttributeError(taxonomyPath, "Unexpected Contentful content type response", "The metadata taxonomy response differed meaningfully from the Terraform plan.")

			mismatch = true
		}
	}

	return projects, mismatch
}

func contentTypeMutationValueEquivalent(name string, valuePath path.Path, planned, remote attr.Value, diags *diag.Diagnostics) bool {
	if planned.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown value", "A Terraform-owned content type value must be known before Contentful can be changed.")

		return false
	}

	if planned.Equal(remote) {
		return true
	}

	diags.AddAttributeError(valuePath, "Unexpected Contentful content type response", "The "+name+" response differed meaningfully from the Terraform plan.")

	return false
}

func contentTypeNormalizedJSONEquivalent(planned, remote jsontypes.Normalized) bool {
	if planned.IsNull() || planned.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return planned.Equal(remote)
	}

	return NormalizeJSON([]byte(planned.ValueString())) == NormalizeJSON([]byte(remote.ValueString()))
}

// Contentful did not canonicalize any modeled scalar, field structure, field
// ordering, validation ordering, default, annotation, or taxonomy value in the
// documented CMA contract or focused live probes. Structural values therefore
// remain exact. Only JSON object representation is normalized, matching the
// provider's existing JSON request/response semantics; JSON array order remains
// significant.
func contentTypeNormalizedJSONListEquivalentAt(valuePath path.Path, planned, remote TypedList[jsontypes.Normalized]) (path.Path, bool) {
	if planned.IsNull() || planned.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return valuePath, planned.Equal(remote)
	}

	limit := min(len(planned.Elements()), len(remote.Elements()))
	for index := range limit {
		if !contentTypeNormalizedJSONEquivalent(planned.Elements()[index], remote.Elements()[index]) {
			return valuePath.AtListIndex(index), false
		}
	}

	if len(planned.Elements()) != len(remote.Elements()) {
		return valuePath.AtListIndex(limit), false
	}

	return valuePath, true
}

func contentTypeFieldItemsEquivalentAt(valuePath path.Path, planned, remote TypedObject[ContentTypeFieldItemsValue]) (path.Path, bool) {
	if planned.IsNull() || planned.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return valuePath, planned.Equal(remote)
	}

	plannedValue, remoteValue := planned.Value(), remote.Value()
	for _, field := range []struct {
		name            string
		planned, remote attr.Value
	}{
		{"type", plannedValue.ItemsType, remoteValue.ItemsType},
		{"link_type", plannedValue.LinkType, remoteValue.LinkType},
	} {
		if !field.planned.Equal(field.remote) {
			return valuePath.AtName(field.name), false
		}
	}

	return contentTypeNormalizedJSONListEquivalentAt(valuePath.AtName("validations"), plannedValue.Validations, remoteValue.Validations)
}

func contentTypeFieldsEquivalentAt(valuePath path.Path, planned, remote TypedList[TypedObject[ContentTypeFieldValue]]) (path.Path, bool) {
	if planned.IsNull() || planned.IsUnknown() || remote.IsNull() || remote.IsUnknown() {
		return valuePath, planned.Equal(remote)
	}

	limit := min(len(planned.Elements()), len(remote.Elements()))
	for index := range limit {
		plannedField, remoteField := planned.Elements()[index], remote.Elements()[index]
		fieldPath := valuePath.AtListIndex(index)

		if plannedField.IsNull() || plannedField.IsUnknown() || remoteField.IsNull() || remoteField.IsUnknown() {
			if !plannedField.Equal(remoteField) {
				return fieldPath, false
			}

			continue
		}

		plannedValue, remoteValue := plannedField.Value(), remoteField.Value()
		for _, field := range []struct {
			name            string
			planned, remote attr.Value
		}{
			{"id", plannedValue.ID, remoteValue.ID},
			{"name", plannedValue.Name, remoteValue.Name},
			{"type", plannedValue.FieldType, remoteValue.FieldType},
			{"link_type", plannedValue.LinkType, remoteValue.LinkType},
			{"disabled", plannedValue.Disabled, remoteValue.Disabled},
			{"omitted", plannedValue.Omitted, remoteValue.Omitted},
			{"required", plannedValue.Required, remoteValue.Required},
			{"localized", plannedValue.Localized, remoteValue.Localized},
			{"allowed_resources", plannedValue.AllowedResources, remoteValue.AllowedResources},
		} {
			if !field.planned.Equal(field.remote) {
				return fieldPath.AtName(field.name), false
			}
		}

		if !contentTypeNormalizedJSONEquivalent(plannedValue.DefaultValue, remoteValue.DefaultValue) {
			return fieldPath.AtName("default_value"), false
		}

		if differencePath, equivalent := contentTypeFieldItemsEquivalentAt(fieldPath.AtName("items"), plannedValue.Items, remoteValue.Items); !equivalent {
			return differencePath, false
		}

		if differencePath, equivalent := contentTypeNormalizedJSONListEquivalentAt(fieldPath.AtName("validations"), plannedValue.Validations, remoteValue.Validations); !equivalent {
			return differencePath, false
		}
	}

	if len(planned.Elements()) != len(remote.Elements()) {
		return valuePath.AtListIndex(limit), false
	}

	return valuePath, true
}
