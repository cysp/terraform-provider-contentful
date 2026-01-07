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

func TestEntryModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rapid.Check(t, func(t *rapid.T) {
		spaceID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "spaceID")
		environmentID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "environmentID")
		contentTypeID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "contentTypeID")
		entryID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "entryID")

		model := testdata.EntryModel(spaceID, environmentID, contentTypeID, entryID).Draw(t, "model")

		entryRequest, diags := model.ToEntryRequest(ctx)
		if diags.HasError() {
			t.Fatalf("ToEntryRequest failed: %v", diags.Errors())
		}

		entry := cmt.NewEntryFromRequest(
			model.SpaceID.ValueString(),
			model.EnvironmentID.ValueString(),
			model.ContentTypeID.ValueString(),
			model.EntryID.ValueString(),
			&entryRequest,
		)

		result, diags := NewEntryResourceModelFromResponse(ctx, entry)
		if diags.HasError() {
			t.Fatalf("NewEntryResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, result)
	})
}
