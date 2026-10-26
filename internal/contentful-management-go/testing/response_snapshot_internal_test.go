package cmtesting

// These tests exercise the private DTO snapshot's representation guarantees,
// which include intentionally invalid values that HTTP codecs cannot preserve.
import (
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseSnapshotPreservesOptionalAndRawRepresentations(t *testing.T) {
	t.Parallel()

	original := cm.Entry{Fields: cm.OptEntryFields{
		Value: cm.EntryFields{"invalid": jx.Raw("not JSON"), "nil": nil, "empty": {}},
		Set:   false,
	}}
	snapshot := snapshotResponse(&original)
	require.NotSame(t, &original, snapshot)
	assert.False(t, snapshot.Fields.Set)
	assert.Equal(t, jx.Raw("not JSON"), snapshot.Fields.Value["invalid"])
	assert.Nil(t, snapshot.Fields.Value["nil"])
	require.NotNil(t, snapshot.Fields.Value["empty"])
	assert.Empty(t, snapshot.Fields.Value["empty"])
	snapshot.Fields.Value["invalid"][0] = '!'
	delete(snapshot.Fields.Value, "nil")
	assert.Equal(t, jx.Raw("not JSON"), original.Fields.Value["invalid"])
	assert.Contains(t, original.Fields.Value, "nil")

	contentType := cm.ContentType{Description: cm.OptNilString{Value: "inactive", Set: true, Null: true}}
	clonedContentType := snapshotResponse(contentType)
	assert.Equal(t, cm.OptNilString{Value: "inactive", Set: true, Null: true}, clonedContentType.Description)
}

func TestResponseSnapshotPreservesNilAndEmptyCollections(t *testing.T) {
	t.Parallel()

	nilFields := snapshotResponse(cm.EntryFields(nil))
	assert.Nil(t, nilFields)

	emptyFields := snapshotResponse(cm.EntryFields{})
	require.NotNil(t, emptyFields)
	assert.Empty(t, emptyFields)

	nilItems := snapshotResponse(cm.EntryCollection{})
	assert.Nil(t, nilItems.Items)
	emptyItems := snapshotResponse(cm.EntryCollection{Items: []cm.Entry{}})
	require.NotNil(t, emptyItems.Items)
	assert.Empty(t, emptyItems.Items)
}

func TestResponseSnapshotPreservesTimeAndResponseType(t *testing.T) {
	t.Parallel()

	timestamp := time.Now()
	assert.Equal(t, cm.NewOptDateTime(timestamp), snapshotResponse(cm.NewOptDateTime(timestamp)))

	zoned := time.Date(2026, time.October, 26, 1, 2, 3, 4, time.FixedZone("fixture", 3600))
	assert.Equal(t, cm.NewOptDateTime(zoned), snapshotResponse(cm.NewOptDateTime(zoned)))

	var absent cm.GetEntryRes
	assert.Nil(t, snapshotResponse(absent))

	var (
		typedNil *cm.Entry
		result   cm.GetEntryRes = typedNil
	)
	assert.IsType(t, typedNil, snapshotResponse(result))
	assert.Nil(t, snapshotResponse(result))
	result = &cm.Entry{Sys: cm.EntrySys{ID: "entry"}}
	snapshot := snapshotResponse(result)
	concrete, ok := snapshot.(*cm.Entry)
	require.True(t, ok)
	assert.Equal(t, "entry", concrete.Sys.ID)
}
