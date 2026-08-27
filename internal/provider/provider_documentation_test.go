package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworklist "github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderSurfaceDocumentation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	provider := New("test")

	testProviderDocumentation(t, providerDocumentationTest{
		category:        "resources",
		documentKind:    "Resource",
		exampleFilename: "resource.tf",
		readmeHeading:   "Resources",
		registeredTypes: resourceTypeNames(ctx, provider.Resources(ctx)),
	})
	testProviderDocumentation(t, providerDocumentationTest{
		category:        "data-sources",
		documentKind:    "Data Source",
		exampleFilename: "data-source.tf",
		readmeHeading:   "Data Sources",
		registeredTypes: dataSourceTypeNames(ctx, provider.DataSources(ctx)),
	})
	testProviderDocumentation(t, providerDocumentationTest{
		category:        "list-resources",
		documentKind:    "List Resource",
		exampleFilename: "list-resource.tfquery.hcl",
		readmeHeading:   "List Resources",
		registeredTypes: listResourceTypeNames(ctx, provider.ListResources(ctx)),
	})
}

func TestAppSigningSecretDocumentationContract(t *testing.T) {
	t.Parallel()

	resourceSchema := AppSigningSecretResourceSchema(t.Context())
	valueAttribute, ok := resourceSchema.Attributes["value"].(resourceschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, valueAttribute.Required)
	assert.True(t, valueAttribute.Sensitive)

	for _, expected := range []string{
		"`Sensitive` masks routine Terraform CLI and HCP Terraform UI output.",
		"After Create or Update, Terraform stores the complete configured value in resource state",
		"saved plan files can also contain it",
		"Protect access to both.",
	} {
		assert.Contains(t, valueAttribute.Description, expected)
	}

	for _, expected := range []string{
		"Contentful returns only `redactedValue` for the secret, not the complete submitted value.",
		"Refresh preserves the value already in Terraform state and cannot detect an out-of-band replacement.",
		"An imported resource initially has a null value",
		"the next configured apply writes and stores the configured replacement",
		"Protect state and saved plan artifacts; state protection depends on the configured Terraform backend.",
	} {
		assert.Contains(t, resourceSchema.Description, expected)
	}

	document, err := os.ReadFile(filepath.Join("..", "..", "docs", "resources", "app_signing_secret.md"))
	require.NoError(t, err)

	for _, expected := range []string{
		"`Sensitive` masks routine Terraform CLI and HCP Terraform UI output.",
		"After Create or Update, Terraform stores the complete configured value in resource state",
		"saved plan files can also contain it",
		"Protect access to both.",
		"Contentful returns only `redactedValue` for the secret, not the complete submitted value.",
		"An imported resource initially has a null value",
		"the next configured apply writes and stores the configured replacement",
		"Protect state and saved plan artifacts; state protection depends on the configured Terraform backend.",
	} {
		assert.Contains(t, string(document), expected)
	}
}

type providerDocumentationTest struct {
	category        string
	documentKind    string
	exampleFilename string
	readmeHeading   string
	registeredTypes []string
}

func testProviderDocumentation(t *testing.T, test providerDocumentationTest) {
	t.Helper()
	t.Run(test.readmeHeading, func(t *testing.T) {
		t.Parallel()

		repositoryRoot := filepath.Join("..", "..")
		documentationDirectory := filepath.Join(repositoryRoot, "docs", test.category)
		exampleDirectory := filepath.Join(repositoryRoot, "examples", test.category)

		require.NotEmpty(t, test.registeredTypes)
		require.Len(t, uniqueStrings(test.registeredTypes), len(test.registeredTypes), "provider type names must be unique")

		registeredTypes := append([]string(nil), test.registeredTypes...)
		sort.Strings(registeredTypes)

		documentedTypes := documentationTypeNames(t, documentationDirectory)
		assert.Equal(t, registeredTypes, documentedTypes, "generated documentation must exactly cover registered provider types")
		assert.Subset(
			t,
			registeredTypes,
			exampleTypeNames(t, exampleDirectory, test.exampleFilename),
			"generator examples must only describe registered provider types",
		)

		readme, err := os.ReadFile(filepath.Join(repositoryRoot, "README.md"))
		require.NoError(t, err)
		assert.Equal(
			t,
			expectedReadmeInventory(test.readmeHeading, test.category, registeredTypes),
			readmeInventory(string(readme), test.readmeHeading),
			"README inventory must exactly cover registered provider types",
		)

		for _, typeName := range registeredTypes {
			documentPath := filepath.Join(documentationDirectory, strings.TrimPrefix(typeName, "contentful_")+".md")
			document, readErr := os.ReadFile(documentPath)
			require.NoError(t, readErr)
			assert.Contains(t, string(document), "# "+typeName+" ("+test.documentKind+")")
			assert.Contains(t, string(document), "# generated by https://github.com/hashicorp/terraform-plugin-docs")
		}
	})
}

func exampleTypeNames(t *testing.T, directory string, exampleFilename string) []string {
	t.Helper()

	entries, err := os.ReadDir(directory)
	require.NoError(t, err)

	typeNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		_, statErr := os.Stat(filepath.Join(directory, entry.Name(), exampleFilename))
		require.NoError(t, statErr, "%s must contain %s", entry.Name(), exampleFilename)
		typeNames = append(typeNames, entry.Name())
	}

	sort.Strings(typeNames)

	return typeNames
}

func resourceTypeNames(ctx context.Context, factories []func() resource.Resource) []string {
	typeNames := make([]string, 0, len(factories))
	for _, factory := range factories {
		response := resource.MetadataResponse{}
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "contentful"}, &response)
		typeNames = append(typeNames, response.TypeName)
	}

	return typeNames
}

func dataSourceTypeNames(ctx context.Context, factories []func() datasource.DataSource) []string {
	typeNames := make([]string, 0, len(factories))
	for _, factory := range factories {
		response := datasource.MetadataResponse{}
		factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "contentful"}, &response)
		typeNames = append(typeNames, response.TypeName)
	}

	return typeNames
}

func listResourceTypeNames(ctx context.Context, factories []func() frameworklist.ListResource) []string {
	typeNames := make([]string, 0, len(factories))
	for _, factory := range factories {
		response := resource.MetadataResponse{}
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "contentful"}, &response)
		typeNames = append(typeNames, response.TypeName)
	}

	return typeNames
}

func documentationTypeNames(t *testing.T, directory string) []string {
	t.Helper()

	entries, err := os.ReadDir(directory)
	require.NoError(t, err)

	typeNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		typeNames = append(typeNames, "contentful_"+strings.TrimSuffix(entry.Name(), ".md"))
	}

	sort.Strings(typeNames)

	return typeNames
}

func expectedReadmeInventory(heading string, category string, typeNames []string) []string {
	lines := make([]string, 0, len(typeNames)+1)

	lines = append(lines, "## "+heading)
	for _, typeName := range typeNames {
		lines = append(lines, "- [`"+typeName+"`](docs/"+category+"/"+strings.TrimPrefix(typeName, "contentful_")+".md)")
	}

	return lines
}

func readmeInventory(readme string, heading string) []string {
	sectionPrefix := "## " + heading + "\n"

	_, section, found := strings.Cut(readme, sectionPrefix)
	if !found {
		return nil
	}

	section, _, _ = strings.Cut(section, "\n## ")
	lines := []string{"## " + heading}

	for line := range strings.SplitSeq(section, "\n") {
		if strings.HasPrefix(line, "- [") {
			lines = append(lines, line)
		}
	}

	return lines
}

func uniqueStrings(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		unique[value] = struct{}{}
	}

	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}

	return result
}
