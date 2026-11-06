package cmtesting

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewExtensionFromFields(spaceID, environmentID, extensionID string, fields cm.ExtensionData) cm.Extension {
	extension := cm.Extension{
		Sys: cm.NewExtensionSys(spaceID, environmentID, extensionID),
	}

	UpdateExtensionFromFields(&extension, fields)

	return extension
}

func UpdateExtensionFromFields(extension *cm.Extension, fields cm.ExtensionData) {
	extension.Sys.Version++

	extension.Extension = fields.Extension
	extension.Extension.FieldTypes = slices.Clone(fields.Extension.FieldTypes)

	extension.Parameters = fields.Parameters
}
