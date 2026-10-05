package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contentTypeDiscoveryInputs() map[string]any {
	return map[string]any{"space_id": "space", "environment_id": "master", "content_type_id": "article"}
}

func TestContentTypeDataSourcesProjectCurrentModel(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "content_type")
	requests := 0
	singular := discoveryReadTest(t.Context(), t, NewContentTypeDataSource, contentTypeDiscoveryInputs(), roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++

		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/base/spaces/space/environments/master/content_types/article", request.URL.Path)
		assert.Empty(t, request.URL.RawQuery)
		assert.Equal(t, "Bearer synthetic-token", request.Header.Get("Authorization"))
		_, hasDeadline := request.Context().Deadline()
		assert.True(t, hasDeadline)

		return discoveryHTTPResponse(request, 200, body), nil
	}))
	require.False(t, singular.Diagnostics.HasError(), singular.Diagnostics)
	assert.Equal(t, 1, requests)

	var single ContentTypeDataSourceModel
	require.False(t, singular.State.Get(t.Context(), &single).HasError())
	assert.Equal(t, "space/master/article", single.ID.ValueString())
	assert.Equal(t, "Article draft", single.Name.ValueString())
	assert.Equal(t, "Editorial article", single.Description.ValueString())
	assert.Equal(t, "title", single.DisplayField.ValueString())
	assert.Equal(t, int64(3), single.PublishedVersion.ValueInt64())
	require.Len(t, single.Fields.Elements(), 4)
	title := single.Fields.Elements()[0].Value()
	assert.Equal(t, "title", title.ID.ValueString())
	assert.Equal(t, "Symbol", title.FieldType.ValueString())
	assert.JSONEq(t, testJSON(map[string]any{"en-US": "Untitled"}), title.DefaultValue.ValueString())
	assert.JSONEq(t, testJSON(map[string]any{"size": map[string]any{"min": 1}}), title.Validations.Elements()[0].ValueString())
	assert.True(t, title.Disabled.ValueBool())
	assert.False(t, title.Omitted.ValueBool())
	assert.True(t, title.Required.ValueBool())
	assert.False(t, title.Localized.ValueBool())
	assert.True(t, title.Items.IsNull())

	array := single.Fields.Elements()[1].Value()
	assert.Equal(t, "Array", array.FieldType.ValueString())
	assert.Equal(t, "Link", array.Items.Value().ItemsType.ValueString())
	assert.Equal(t, "Entry", array.Items.Value().LinkType.ValueString())
	assert.JSONEq(t, testJSON(map[string]any{"linkContentType": []any{"author"}}), array.Items.Value().Validations.Elements()[0].ValueString())
	assert.JSONEq(t, testJSON(map[string]any{"size": map[string]any{"max": 3}}), array.Validations.Elements()[0].ValueString())
	link := single.Fields.Elements()[2].Value()
	assert.Equal(t, "Link", link.FieldType.ValueString())
	assert.Equal(t, "Asset", link.LinkType.ValueString())
	assert.True(t, link.Items.IsNull())

	resource := single.Fields.Elements()[3].Value()
	require.Len(t, resource.AllowedResources.Elements(), 2)
	allowed := resource.AllowedResources.Elements()[0].Value()
	assert.True(t, allowed.External.IsNull())
	assert.Equal(t, "crn:contentful:::content:spaces/space/environments/master", allowed.ContentfulEntry.Value().Source.ValueString())
	assert.Equal(t, "author", allowed.ContentfulEntry.Value().ContentTypes.Elements()[0].ValueString())
	assert.Equal(t, "article", allowed.ContentfulEntry.Value().ContentTypes.Elements()[1].ValueString())
	assert.True(t, resource.AllowedResources.Elements()[1].Value().ContentfulEntry.IsNull())
	assert.Equal(t, "External:Product", resource.AllowedResources.Elements()[1].Value().External.Value().TypeID.ValueString())
	metadata := single.Metadata.Value()
	assert.Contains(t, metadata.Annotations.ValueString(), testJSON("Contentful:AggregateRoot"))
	require.Len(t, metadata.Taxonomy.Elements(), 2)
	assert.Equal(t, "topic", metadata.Taxonomy.Elements()[0].Value().TaxonomyConcept.Value().ID.ValueString())
	assert.True(t, metadata.Taxonomy.Elements()[0].Value().TaxonomyConcept.Value().Required.ValueBool())
	assert.True(t, metadata.Taxonomy.Elements()[0].Value().TaxonomyConceptScheme.IsNull())
	assert.Equal(t, "topics", metadata.Taxonomy.Elements()[1].Value().TaxonomyConceptScheme.Value().ID.ValueString())
	assert.False(t, metadata.Taxonomy.Elements()[1].Value().TaxonomyConceptScheme.Value().Required.ValueBool())

	plural := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++

		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/base/spaces/space/environments/master/content_types", request.URL.Path)
		assert.Equal(t, "0", request.URL.Query().Get("skip"))
		assert.Equal(t, "100", request.URL.Query().Get("limit"))
		assert.Len(t, request.URL.Query(), 2)

		return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 1, body)), nil
	}))
	require.False(t, plural.Diagnostics.HasError(), plural.Diagnostics)
	assert.Equal(t, 2, requests)

	var all ContentTypesDataSourceModel
	require.False(t, plural.State.Get(t.Context(), &all).HasError())
	assert.Equal(t, "space/master", all.ID.ValueString())
	require.Len(t, all.ContentTypes, 1)
	assert.Equal(t, single.ContentTypeDataSourceItemModel, all.ContentTypes[0])
}

func TestContentTypeDataSourcesPreserveNullAndEmptyStrings(t *testing.T) {
	t.Parallel()

	// displayField is required but nullable in the generated decoder. Its
	// omission is a decoding error; description is optional and may be absent.
	var (
		missing = testJSON(map[string]any{
			"sys": map[string]any{
				"type":        "ContentType",
				"id":          "article",
				"version":     1,
				"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
			},
			"name":         "Draft",
			"displayField": nil,
			"fields":       []any{},
		})
		empty = testJSON(map[string]any{
			"sys": map[string]any{
				"type":             "ContentType",
				"id":               "article",
				"version":          2,
				"publishedVersion": 1,
				"space":            map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment":      map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
			},
			"name":         "Draft",
			"description":  "",
			"displayField": "",
			"fields":       []any{},
		})
	)

	for _, test := range []struct {
		name          string
		body          string
		wantNull      bool
		wantPublished bool
	}{
		{"missing and null", missing, true, false},
		{"explicit empty", empty, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, NewContentTypeDataSource, contentTypeDiscoveryInputs(), roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, test.body), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

			var model ContentTypeDataSourceModel
			require.False(t, response.State.Get(t.Context(), &model).HasError())
			assert.Equal(t, test.wantNull, model.Description.IsNull())
			assert.Equal(t, test.wantNull, model.DisplayField.IsNull())
			assert.Equal(t, !test.wantPublished, model.PublishedVersion.IsNull())

			if !test.wantNull {
				assert.Empty(t, model.Description.ValueString())
				assert.Empty(t, model.DisplayField.ValueString())
			}
		})
	}
}

func TestContentTypesDataSourcePagesPreserveResponseOrderAndDuplicates(t *testing.T) {
	t.Parallel()

	item := func(id, name string) string {
		return testJSON(map[string]any{
			"sys": map[string]any{
				"type":        "ContentType",
				"id":          id,
				"version":     1,
				"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
			},
			"name":         name,
			"displayField": nil,
			"fields":       []any{},
		})
	}

	zeta := item("zeta", "First zeta")
	alpha := item("alpha", "Alpha")
	zetaAgain := item("zeta", "Second zeta")

	var offsets []string

	response := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		offsets = append(offsets, request.URL.Query().Get("skip"))
		if len(offsets) == 1 {
			return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 3, zeta)), nil
		}

		return discoveryHTTPResponse(request, 200, discoveryPage(1, 2, 3, alpha, zetaAgain)), nil
	}))
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	assert.Equal(t, []string{"0", "1"}, offsets)

	var model ContentTypesDataSourceModel
	require.False(t, response.State.Get(t.Context(), &model).HasError())
	require.Len(t, model.ContentTypes, 3)
	assert.Equal(t, []string{"zeta", "alpha", "zeta"}, []string{model.ContentTypes[0].ContentTypeID.ValueString(), model.ContentTypes[1].ContentTypeID.ValueString(), model.ContentTypes[2].ContentTypeID.ValueString()})
	assert.Equal(t, "First zeta", model.ContentTypes[0].Name.ValueString())
	assert.Equal(t, "Second zeta", model.ContentTypes[2].Name.ValueString())
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestContentTypeDataSourcesRejectResponseIdentity(t *testing.T) {
	t.Parallel()

	item := testJSON(map[string]any{
		"sys": map[string]any{
			"type":        "ContentType",
			"id":          "article",
			"version":     1,
			"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
			"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
		},
		"name":         "Draft",
		"displayField": nil,
		"fields":       []any{},
	})
	for _, test := range []struct{ name, body, want string }{
		{"content type", mutateTestJSON(item, func(document map[string]any) { document["sys"].(map[string]any)["id"] = "different" }), "requested ID"},
		{"space", mutateTestJSON(item, func(document map[string]any) {
			document["sys"].(map[string]any)["space"].(map[string]any)["sys"].(map[string]any)["id"] = "other"
		}), "outside the requested scope"},
		{"environment", mutateTestJSON(item, func(document map[string]any) {
			document["sys"].(map[string]any)["environment"].(map[string]any)["sys"].(map[string]any)["id"] = "target"
		}), "cannot establish alias routing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, NewContentTypeDataSource, contentTypeDiscoveryInputs(), roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, test.body), nil
			}))
			require.True(t, response.Diagnostics.HasError())
			assert.Contains(t, fmt.Sprint(response.Diagnostics), test.want)
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}

func TestContentTypesDataSourceLaterPageFailurePublishesNothing(t *testing.T) {
	t.Parallel()

	first := testJSON(map[string]any{
		"sys": map[string]any{
			"type":        "ContentType",
			"id":          "article",
			"version":     1,
			"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
			"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
		},
		"name":         "Draft",
		"displayField": nil,
		"fields":       []any{},
	})

	for _, test := range []struct {
		name, body, want string
		status           int
	}{
		{"permission", testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "AccessDenied"}, "message": "denied"}), "AccessDenied", 403},
		{"missing", testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}}), "NotFound", 404},
		{"decoder", `{"items":`, "decode", 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			count := 0
			response := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++
				if count == 1 {
					return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, first)), nil
				}

				return discoveryHTTPResponse(request, test.status, test.body), nil
			}))
			require.True(t, response.Diagnostics.HasError())
			assert.Contains(t, fmt.Sprint(response.Diagnostics), test.want)
			assert.Equal(t, 2, count)
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}

func TestContentTypesDataSourceWarningPathUsesResponseOrder(t *testing.T) {
	t.Parallel()

	// The generated decoder rejects unknown taxonomy link types. Typed input
	// tests the provider projector's lenient warning behavior separately.
	alpha := cm.ContentType{Sys: cm.NewContentTypeSys("space", "master", "alpha"), Name: "Alpha", DisplayField: cm.NewNilString(""), Fields: []cm.ContentTypeFieldsItem{}}
	alpha.Metadata.SetTo(cm.ContentTypeMetadata{Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{Sys: cm.ContentTypeMetadataTaxonomyItemSys{Type: cm.ContentTypeMetadataTaxonomyItemSysTypeLink, ID: "future", LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkType("Future")}}}})
	zeta := cm.ContentType{Sys: cm.NewContentTypeSys("space", "master", "zeta"), Name: "Zeta", DisplayField: cm.NewNilString(""), Fields: []cm.ContentTypeFieldsItem{}}
	items, diagnostics := projectContentTypeDataSourceItems(t.Context(), []cm.ContentType{zeta, alpha}, "space", "master")
	assert.False(t, diagnostics.HasError())
	require.Len(t, diagnostics.Warnings(), 1)
	withPath, ok := diagnostics.Warnings()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, `content_types[1].metadata.taxonomy[0]`, withPath.Path().String())
	require.Len(t, items, 2)
	assert.Equal(t, "zeta", items[0].ContentTypeID.ValueString())
	assert.Equal(t, "alpha", items[1].ContentTypeID.ValueString())
	assert.False(t, items[1].Metadata.Value().Taxonomy.Elements()[0].IsNull())
}

func TestContentTypeDataSourceCancelledReadPublishesNothing(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	response := discoveryReadTest(ctx, t, NewContentTypeDataSource, contentTypeDiscoveryInputs(), roundTripFunc(func(request *http.Request) (*http.Response, error) {
		cancel()

		return discoveryHTTPResponse(request, 200, discoveryFixture(t, "content_type")), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
}

func TestContentTypesDataSourceScopeAndEmpty(t *testing.T) {
	t.Parallel()

	item := func(id, environment string) string {
		return testJSON(map[string]any{
			"sys": map[string]any{
				"type":        "ContentType",
				"id":          id,
				"version":     1,
				"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": environment}},
			},
			"name":         "Draft",
			"displayField": nil,
			"fields":       []any{},
		})
	}

	inputs := map[string]any{"space_id": "space", "environment_id": "master"}
	valid := item("alpha", "master")
	wrongScope := item("zeta", "concrete")
	response := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 2, 2, valid, wrongScope)), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
	withPath, ok := response.Diagnostics.Errors()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, "content_types[1].content_type_id", withPath.Path().String())
	assert.Contains(t, fmt.Sprint(response.Diagnostics), "cannot establish alias routing")

	empty := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 0, 0)), nil
	}))
	require.False(t, empty.Diagnostics.HasError(), empty.Diagnostics)

	var model ContentTypesDataSourceModel
	require.False(t, empty.State.Get(t.Context(), &model).HasError())
	assert.NotNil(t, model.ContentTypes)
	assert.Empty(t, model.ContentTypes)

	// A collection item ID is decoded output, not a lookup input validator.
	computed := discoveryReadTest(t.Context(), t, NewContentTypesDataSource, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 1, item("a/b", "master"))), nil
	}))
	require.False(t, computed.Diagnostics.HasError(), computed.Diagnostics)
	require.False(t, computed.State.Get(t.Context(), &model).HasError())
	require.Len(t, model.ContentTypes, 1)
	assert.Equal(t, "a/b", model.ContentTypes[0].ContentTypeID.ValueString())
}

func TestContentTypesDataSourceCancellationAfterFirstPagePublishesNothing(t *testing.T) {
	t.Parallel()

	first := testJSON(map[string]any{
		"sys": map[string]any{
			"type":        "ContentType",
			"id":          "article",
			"version":     1,
			"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
			"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
		},
		"name":         "Draft",
		"displayField": nil,
		"fields":       []any{},
	})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	reads := 0
	response := discoveryReadTest(ctx, t, NewContentTypesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		reads++

		cancel()

		return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, first)), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
	assert.Equal(t, 1, reads)
}

func TestContentTypeDataSourcesRejectUnresolvedInputsAndMissingEntities(t *testing.T) {
	t.Parallel()

	for _, value := range []any{nil, tftypes.UnknownValue, "", ".", "..", "a/b"} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			t.Parallel()

			inputs := contentTypeDiscoveryInputs()
			inputs["content_type_id"] = value
			reads := 0
			response := discoveryReadTest(t.Context(), t, NewContentTypeDataSource, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				reads++

				return discoveryHTTPResponse(request, 500, ""), nil
			}))
			assert.True(t, response.Diagnostics.HasError())
			assert.True(t, response.State.Raw.IsNull())
			assert.Zero(t, reads)
		})
	}

	for _, test := range []struct {
		name, body, want string
		status           int
	}{
		{"not found", testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}}), "NotFound", 404},
		{"omitted displayField", testJSON(map[string]any{
			"sys": map[string]any{
				"type":        "ContentType",
				"id":          "article",
				"version":     1,
				"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "master"}},
			},
			"name":   "Draft",
			"fields": []any{},
		}), "displayField", 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, NewContentTypeDataSource, contentTypeDiscoveryInputs(), roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, test.status, test.body), nil
			}))
			require.True(t, response.Diagnostics.HasError())
			assert.Contains(t, fmt.Sprint(response.Diagnostics), test.want)
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}
