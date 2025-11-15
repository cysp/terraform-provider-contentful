# Code Patterns and Best Practices

This document describes common patterns used in this provider and best practices for maintaining code quality.

## Pointer Handling

### Safe Pointer Dereferencing Pattern

The codebase uses a safe pattern for handling pointers from Terraform values:

```go
// SAFE: Pointer is dereferenced immediately
value := model.Default.ValueStringPointer()
if value != nil {
    parameter.Default = jx.Raw(*value)
}
```

The `NewOptPointerString` function also safely dereferences pointers:

```go
// From optstring.go - SAFE implementation
func NewOptPointerString(value *string) OptString {
    if value == nil {
        return OptString{}
    }
    return NewOptString(*value)  // Pointer dereferenced here
}
```

### Why This Is Safe

- The pointer returned by `ValueStringPointer()` is only used locally
- Values are dereferenced (`*value`) immediately when used
- No references to the pointer escape the function scope
- No risk of pointer aliasing or unintended mutations

## Variable Naming to Avoid Shadowing

### Problem Pattern (Avoid)

```go
// CONFUSING: path variable shadows the parameter
func process(ctx context.Context, path path.Path) {
    if something {
        path := path.AtName("field")  // Shadows the parameter!
        doSomething(path)
    }
}
```

While technically correct (RHS evaluated before LHS assignment), this pattern:
- Makes code harder to understand
- Could lead to bugs if refactored carelessly
- Requires careful reading to verify correctness

### Recommended Pattern

```go
// CLEAR: Use descriptive variable names
func process(ctx context.Context, path path.Path) {
    if something {
        fieldPath := path.AtName("field")  // Clear intent
        doSomething(fieldPath)
    }
}
```

Benefits:
- Immediately obvious which path is being used
- Easier to refactor without introducing bugs
- Self-documenting code

### Naming Conventions for Nested Paths

Use descriptive suffixes based on context:
- `parametersPath := path.AtName("parameters")`
- `installationPath := parametersPath.AtName("installation")`
- `fieldPath := path.AtListIndex(index)`
- `policyPath := path.AtListIndex(index)`

## Duplication vs. Readability

### When NOT to Extract Functions

The codebase intentionally avoids extracting some duplicated logic when:
1. The logic is straightforward and linear
2. Extraction would require jumping between files
3. The duplication makes the flow clearer
4. Context matters more than DRY principle

Example: Parameter conversion for installation vs. instance parameters is kept separate because:
- Each block is readable in isolation
- The differences (field names) are immediately visible
- Developers can understand the full flow without jumping to another file

### When TO Extract Functions

Extract when:
- The abstraction is clear and obvious
- The helper function has a clear, single purpose
- It significantly reduces complexity
- The function name clearly describes what it does

## Diagnostics Handling

Standard pattern for collecting diagnostics:

```go
func Convert(ctx context.Context, path path.Path) (Result, diag.Diagnostics) {
    diags := diag.Diagnostics{}
    
    value, valueDiags := ConvertValue(ctx, path.AtName("value"))
    diags.Append(valueDiags...)  // Accumulate diagnostics
    
    if diags.HasError() {
        return Result{}, diags  // Early return on error
    }
    
    return Result{Value: value}, diags
}
```

Key points:
- Always accumulate diagnostics with `Append`
- Return early only when necessary (errors prevent continuation)
- Return accumulated diagnostics even on success

## Path Construction

Paths help provide clear error messages by indicating exactly where in the Terraform configuration an issue occurred.

### Best Practices

```go
// Build paths as you traverse the data structure
func processConfig(ctx context.Context, basePath path.Path, config Config) diag.Diagnostics {
    diags := diag.Diagnostics{}
    
    if config.Parameters != nil {
        paramsPath := basePath.AtName("parameters")
        
        for index, param := range config.Parameters {
            paramPath := paramsPath.AtListIndex(index)
            // Use paramPath for error reporting
        }
    }
    
    return diags
}
```

### Path Methods

- `path.AtName("field")` - Navigate to a named field
- `path.AtListIndex(i)` - Navigate to a list element
- `path.AtMapKey(key)` - Navigate to a map value
- `path.Root("field")` - Start from root

## Loop Variable Safety

Go 1.22+ has per-iteration loop variables, but older patterns are still relevant:

### Safe Pattern (Always Correct)

```go
// SAFE: No pointer to loop variable
for index, item := range items {
    value := convertItem(item)  // item is copied by value
    results = append(results, value)
}
```

### Pattern to Avoid

```go
// PROBLEMATIC in Go < 1.22
for _, item := range items {
    items = append(items, &item)  // All pointers point to same variable!
}

// SAFE alternative
for _, item := range items {
    itemCopy := item
    items = append(items, &itemCopy)
}
```

The codebase doesn't take addresses of loop variables, which is the safe approach.
