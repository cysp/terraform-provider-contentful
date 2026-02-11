package provider_test

import (
	"regexp"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentStatusReadyDataSource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetEnvironment("0p38pssr0fi3", "master", "ready", cm.EnvironmentData{
		Name: "master",
	})

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("master"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "id", "0p38pssr0fi3/master"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "space_id", "0p38pssr0fi3"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "environment_id", "master"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "status", "ready"),
				),
			},
		},
	})
}

func TestAccEnvironmentStatusReadyDataSourcePolling(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetEnvironment("0p38pssr0fi3", "staging", "queued", cm.EnvironmentData{
		Name: "Staging Environment",
	})

	go func() {
		time.Sleep(300 * time.Millisecond)
		server.SetEnvironment("0p38pssr0fi3", "staging", "ready", cm.EnvironmentData{
			Name: "Staging Environment",
		})
	}()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("staging"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "id", "0p38pssr0fi3/staging"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "space_id", "0p38pssr0fi3"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "environment_id", "staging"),
					resource.TestCheckResourceAttr("data.contentful_environment_status_ready.test", "status", "ready"),
				),
			},
		},
	})
}

func TestAccEnvironmentStatusReadyDataSourceNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Provider produced null object`),
			},
		},
	})
}
