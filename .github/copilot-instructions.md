# Copilot Instructions for terraform-provider-contentful

This repository contains a Terraform provider for Contentful, built with Go and terraform-plugin-framework.

## Technology Stack

- **Language**: Go 1.25.4
- **Framework**: terraform-plugin-framework v1.17.0
- **Testing**: hashicorp/terraform-plugin-testing v1.14.0
- **Code Generation**: ogen-go/ogen v1.17.0 (for Contentful Management API client)

## Build, Test, and Lint Commands

### Building
```bash
go build -v .
```

### Testing
```bash
# Run all unit tests
go test -v -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...

# Run provider acceptance tests (requires CONTENTFUL_MANAGEMENT_ACCESS_TOKEN)
TF_ACC=1 go test -v ./internal/provider/

# Run mocked acceptance tests
TF_ACC=1 TF_ACC_MOCKED=1 go test -v ./internal/provider/
```

### Linting
```bash
golangci-lint run
```

### Generate Documentation
```bash
go generate
```

## Coding Conventions and Patterns

### Naming Conventions

1. **Variable Naming in Terraform Provider Resources**:
   - Use `plan` for variables receiving data from `req.Plan.Get()`
   - Use `state` for variables receiving data from `req.State.Get()`
   - Use `data` for API responses
   - Example:
     ```go
     var plan WebhookModel
     diags := req.Plan.Get(ctx, &plan)
     ```

2. **Helper Function Parameters**:
   - Use domain-specific names (e.g., `entry`, `contentType`, `webhook`) instead of generic `data`
   - Example:
     ```go
     func createEntry(ctx context.Context, entry EntryModel, ...) { }
     func updateWebhook(ctx context.Context, webhook WebhookModel, ...) { }
     ```

3. **Short Variable Names**:
   - Acceptable single-letter names: `i`, `k`, `r`, `v`, `w` (configured in .golangci.yml)

### Resource Implementation Patterns

1. **Resources with ResourceWithIdentity Interface**:
   - Must set `resp.Identity` in Create, Read, and Update methods
   - Use `CopyAttributeValues` and `resp.Identity.Set`
   - Example:
     ```go
     var identity IdIdentityModel
     diags = CopyAttributeValues(ctx, space, &identity)
     resp.Diagnostics.Append(diags...)
     diags = resp.Identity.Set(ctx, identity)
     ```

2. **Update Method Pattern**:
   - Must set both `resp.Identity` and `resp.State`
   - Follow same pattern as Create and Read methods

3. **Delete Method Pattern**:
   - Do not use `HasError` checks before making API calls unless diagnostics were added between getting state and calling the client
   - Some resources have no-op Delete methods when the resource cannot be deleted from the API

4. **Import State**:
   - Multipart IDs should be validated in import functions

### List Resources Pattern

1. **Structure**:
   - List resource config structs and schemas are in separate `*_config.go` files
   - Implementation uses `stream.Results` as a generator function that yields `list.ListResult` items

2. **Implementation Details**:
   - Check `req.IncludeResource` before setting `result.Resource`
   - Use type switch to check response types before handling
   - Use `util.ErrorDetailFromContentfulManagementResponse` for all API error messages
   - Example:
     ```go
     if req.IncludeResource {
         diags = result.Resource.Set(ctx, entry)
         if !yield(result) {
             return
         }
     }
     ```

### Testing Practices

1. **Test Helper Functions**:
   - Use `ContentfulProviderMockedResourceTest` for tests with pre-populated server data
   - Use `ContentfulProviderMockableResourceTest` for tests that create resources from scratch or can optionally use mocks
   - Use `ContentfulProviderMockableResourceTest` for list resource tests

2. **Test Organization**:
   - Tests should be in `*_test.go` files in the same package with `_test` suffix (e.g., `provider_test`)
   - Use `t.Parallel()` for tests that can run in parallel
   - Test data is stored in `testdata/` directories

### Error Handling

1. **API Error Messages**:
   - Use `util.ErrorDetailFromContentfulManagementResponse` for all Contentful API error messages
   - Especially important in list resources

2. **Diagnostic Handling**:
   - Check and append diagnostics consistently
   - Use `resp.Diagnostics.Append(diags...)` pattern

### JSON Handling

1. **Normalization**:
   - Normalize JSON from API responses using `json.Unmarshal`/`json.Marshal` to ensure consistent formatting
   - Example from `entry_model_response.go`:
     ```go
     var normalized any
     if err := json.Unmarshal([]byte(fieldValue), &normalized); err != nil {
         return diags
     }
     normalizedJSON, err := json.Marshal(normalized)
     ```

2. **Link Handling**:
   - For taxonomy concept and tag links, extract/set only the ID from the sys field
   - Skip empty string IDs

### Code Organization

1. **Model Files**:
   - `*_model.go`: Core model definitions
   - `*_model_request.go`: Request transformation logic
   - `*_model_response.go`: Response transformation logic
   - `*_model_test.go`: Model tests

2. **Resource Files**:
   - `resource_*.go`: Resource CRUD implementation
   - `resource_*_schema.go`: Resource schema definitions
   - `resource_*_test.go`: Resource tests

3. **Data Source Files**:
   - `data_source_*.go`: Data source implementation
   - `data_source_*_schema.go`: Data source schema definitions
   - `data_source_*_test.go`: Data source tests

## Terraform Provider Specific Guidelines

1. **Schema Definitions**:
   - Keep schema definitions in separate `*_schema.go` files
   - Use terraform-plugin-framework types consistently

2. **State Management**:
   - Always use `resp.State.Set()` to update state
   - Always use `req.State.Get()` to read current state
   - Always use `req.Plan.Get()` to read planned values

3. **Resource Identity**:
   - Resources implementing `ResourceWithIdentity` must properly set identity in all CRUD operations

4. **Private Provider Data**:
   - Use for storing metadata that shouldn't be in state
   - Common pattern for version tracking

## API Client

- The Contentful Management API client is generated using ogen-go
- Client code is in `internal/contentful-management-go/`
- Mock servers for testing are in `internal/contentful-management-go/testing/`

## Documentation

- Provider documentation is auto-generated using terraform-plugin-docs
- Examples are in `examples/` directory
- Run `go generate` to regenerate documentation
- Format Terraform files with `terraform fmt -recursive ./examples/`

## Common Pitfalls to Avoid

1. Don't use generic variable names like `data` for domain-specific models
2. Don't forget to set both `resp.Identity` and `resp.State` in Update methods for resources with identity
3. Don't add unnecessary `HasError` checks in Delete methods
4. Don't forget to check `req.IncludeResource` in list resources before setting `result.Resource`
5. Don't use `ContentfulProviderMockableResourceTest` when you mean `ContentfulProviderMockedResourceTest`
6. Always use `util.ErrorDetailFromContentfulManagementResponse` for API errors in list resources
7. Always normalize JSON from API responses for consistent formatting

## Additional Notes

- The provider uses retryable HTTP client for API resilience
- Version is set by goreleaser during release builds
- The provider address is `registry.terraform.io/cysp/contentful`
