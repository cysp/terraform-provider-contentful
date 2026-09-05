---
page_title: "Getting started"
description: |-
  Configure the Contentful provider and plan a first managed resource in an existing Contentful space.
---

# Getting started

The provider manages configuration and content inside existing Contentful spaces. It does not create spaces. Before starting, have a Contentful Management API token and the ID of an existing space and environment that the token can manage.

## Configure the provider

Declare the provider in your Terraform configuration. For production use, add an intentional provider version constraint that matches your upgrade policy rather than relying on an unconstrained latest version.

```terraform
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {}
```

Set the management token in the environment so it is not written directly into the configuration:

```shell
export CONTENTFUL_MANAGEMENT_ACCESS_TOKEN='...'
```

The `access_token` provider attribute is also supported. A configured non-null provider value takes precedence over `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN`.

## Manage an object in an existing space

The following example creates a private Contentful Tag in an existing environment. Declare the inputs explicitly so it is clear which existing Contentful objects the configuration depends on:

```terraform
variable "contentful_space_id" {
  type = string
}

variable "contentful_environment_id" {
  type    = string
  default = "master"
}

resource "contentful_tag" "example" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id
  tag_id         = "terraform-example"

  name       = "Terraform example"
  visibility = "private"
}
```

Initialize Terraform and review the first plan:

```shell
terraform init
terraform plan -var='contentful_space_id=SPACE_ID'
```

With a valid existing space and environment, the plan should propose creating one private `contentful_tag` resource. Review the plan before applying it; `terraform apply` would create that Tag in Contentful.

For existing objects that should be brought under Terraform rather than created, use the Import section of the corresponding resource page. For operation deadline and readiness behavior, see [Operation timeouts](operation-timeouts).
