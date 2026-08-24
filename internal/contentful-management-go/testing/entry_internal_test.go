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
			"empty":          jx.Raw(`{"en-US":[]}`),
			"raw-null":       jx.Raw(`null`),
			"localized-null": jx.Raw(`{"en-US":null}`),
		}),
	}
	entry := NewEntryFromRequest("space", "environment", "content-type", "entry", &request)

	response := projectEntryResponse(entry)

	assert.Contains(t, entry.Fields.Value, "empty")
	assert.JSONEq(t, `null`, string(entry.Fields.Value["raw-null"]))
	assert.JSONEq(t, `{"en-US":null}`, string(entry.Fields.Value["localized-null"]))

	assert.NotContains(t, response.Fields.Value, "empty")
	assert.NotContains(t, response.Fields.Value, "raw-null")
	assert.JSONEq(t, `{"en-US":null}`, string(response.Fields.Value["localized-null"]))

	onlyRawNull := NewEntryFromRequest("space", "environment", "content-type", "raw-null-entry", &cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{"raw-null": jx.Raw(`null`)}),
	})
	onlyRawNullResponse := projectEntryResponse(onlyRawNull)

	assert.JSONEq(t, `null`, string(onlyRawNull.Fields.Value["raw-null"]))
	assert.False(t, onlyRawNullResponse.Fields.IsSet())
}
