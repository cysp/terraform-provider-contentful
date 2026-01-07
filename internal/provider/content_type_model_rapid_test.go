package provider_test

import (
	"context"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/testdata"
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestContentTypeModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rapid.Check(t, func(t *rapid.T) {
		spaceID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "spaceID")
		environmentID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "environmentID")
		contentTypeID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "contentTypeID")

		model := testdata.ContentTypeModel(spaceID, environmentID, contentTypeID).Draw(t, "model")

		request, diags := model.ToContentTypeRequestData(ctx)
		if diags.HasError() {
			t.Fatalf("ToContentTypeRequestData failed: %v", diags.Errors())
		}

		contentType := cmt.NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, request)

		result, diags := NewContentTypeResourceModelFromResponse(ctx, contentType)
		if diags.HasError() {
			t.Fatalf("NewContentTypeResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, result)
	})
}
