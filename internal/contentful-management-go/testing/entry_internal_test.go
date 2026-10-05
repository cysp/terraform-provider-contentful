package cmtesting

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
)

func TestProjectEntryResponseDoesNotMutateStoredEntry(t *testing.T) {
	t.Parallel()

	request := cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"empty":          jx.Raw(testJSON(map[string]any{"en-US": []any{}})),
			"raw-null":       jx.Raw(testJSON(nil)),
			"localized-null": jx.Raw(testJSON(map[string]any{"en-US": nil})),
		}),
	}
	entry := NewEntryFromRequest("space", "environment", "content-type", "entry", &request)

	response := projectEntryResponse(entry)

	assert.Contains(t, entry.Fields.Value, "empty")
	assert.JSONEq(t, testJSON(nil), string(entry.Fields.Value["raw-null"]))
	assert.JSONEq(t, testJSON(map[string]any{"en-US": nil}), string(entry.Fields.Value["localized-null"]))

	assert.NotContains(t, response.Fields.Value, "empty")
	assert.NotContains(t, response.Fields.Value, "raw-null")
	assert.JSONEq(t, testJSON(map[string]any{"en-US": nil}), string(response.Fields.Value["localized-null"]))

	onlyRawNull := NewEntryFromRequest("space", "environment", "content-type", "raw-null-entry", &cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{"raw-null": jx.Raw(testJSON(nil))}),
	})
	onlyRawNullResponse := projectEntryResponse(onlyRawNull)

	assert.JSONEq(t, testJSON(nil), string(onlyRawNull.Fields.Value["raw-null"]))
	assert.False(t, onlyRawNullResponse.Fields.IsSet())
}
