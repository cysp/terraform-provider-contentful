package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentDataSerialization(t *testing.T) {
	t.Parallel()

	data, err := new(cm.PreviewEnvironmentData{
		Name:        "Preview",
		Description: "",
		Configurations: []cm.PreviewEnvironmentConfigurationData{{
			URL:        "https://preview.invalid/{entry.sys.id}",
			EntityType: "ContentType",
			EntityId:   "page",
			Enabled:    true,
		}},
	}).MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"name": "Preview",
		"description": "",
		"configurations": [{
			"url": "https://preview.invalid/{entry.sys.id}",
			"entityType": "ContentType",
			"entityId": "page",
			"enabled": true
		}]
	}`, string(data))
}

func TestPreviewEnvironmentCreateDataSerialization(t *testing.T) {
	t.Parallel()

	createData := cm.NewPreviewEnvironmentCreateData(cm.PreviewEnvironmentData{
		Name:        "Preview",
		Description: "",
		Configurations: []cm.PreviewEnvironmentConfigurationData{
			{
				URL:        "https://preview.invalid/{entry.sys.id}",
				EntityType: "ContentType",
				EntityId:   "page",
				Enabled:    true,
			},
		},
	})
	data, err := createData.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"name": "Preview",
		"description": "",
		"configurations": [{
			"url": "https://preview.invalid/{entry.sys.id}",
			"entityType": "ContentType",
			"entityId": "page",
			"enabled": true
		}]
	}`, string(data))
}
