# Migration Guide: Deprecated API Endpoints

## Overview

Atlassian is deprecating multiple REST API endpoints as part of their Enhanced JQL Service rollout. This SDK has been updated to support the new endpoints with better performance and scalability.

**Affected Endpoints:**
1. `/rest/api/3/search` → `/rest/api/3/search/jql` (Removal: October 31, 2025)
2. `/rest/api/3/expression/eval` → `/rest/api/3/expression/evaluate` (Removal: August 1, 2025)

## Timeline

- **May 1, 2025**: Endpoint officially deprecated (still functional)
- **May 5 - July 31, 2025**: Enhanced JQL Service rolled out under the hood for old APIs
- **August 1 - October 31, 2025**: Progressive shutdown of deprecated APIs
- **After October 31, 2025**: Complete removal of `/rest/api/3/search` endpoint

## Key Differences

### Pagination
- **Old API**: Offset-based pagination using `StartAt`
- **New API**: Token-based pagination using `NextPageToken`

### Default Fields
- **Old API**: Returns `*navigable` fields by default
- **New API**: Returns only `id` field by default (you must explicitly request other fields)

### Total Count
- **Old API**: Provides total count of matching issues
- **New API**: No total count (for performance reasons)

### Max Results
- **Old API**: Maximum 100 results per page
- **New API**: Up to 5000 results per page (when fewer fields are requested)

## Migration Examples

### Basic Search

**Old API (Deprecated):**
```go
results, err := client.Search.Search(ctx, &search.SearchOptions{
    JQL: "project = PROJ AND status = Open",
    MaxResults: 50,
})
```

**New API:**
```go
results, err := client.Search.SearchJQL(ctx, &search.SearchJQLOptions{
    JQL: "project = PROJ AND status = Open",
    Fields: []string{"summary", "status", "assignee"}, // Explicitly specify fields
    MaxResults: 50,
})
```

### Pagination

**Old API (Deprecated):**
```go
opts := &search.SearchOptions{
    JQL: "project = PROJ",
    MaxResults: 50,
    StartAt: 0,
}

for {
    results, err := client.Search.Search(ctx, opts)
    if err != nil {
        return err
    }

    // Process results...

    if opts.StartAt + len(results.Issues) >= results.Total {
        break
    }
    opts.StartAt += len(results.Issues)
}
```

**New API:**
```go
opts := &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"summary", "status"},
    MaxResults: 50,
}

for {
    results, err := client.Search.SearchJQL(ctx, opts)
    if err != nil {
        return err
    }

    // Process results...

    if !results.HasNextPage() {
        break
    }
    opts.NextPageToken = results.NextPageToken
}
```

### Iterator Pattern

**Old API (Deprecated):**
```go
iter := client.Search.NewSearchIterator(ctx, &search.SearchOptions{
    JQL: "project = PROJ",
    MaxResults: 50,
})

for iter.Next() {
    issue := iter.Issue()
    fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
}

if err := iter.Err(); err != nil {
    return err
}
```

**New API:**
```go
iter := client.Search.NewSearchJQLIterator(ctx, &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"summary", "status", "assignee"},
    MaxResults: 100, // Can use higher values
})

for iter.Next() {
    issue := iter.Issue()
    fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
}

if err := iter.Err(); err != nil {
    return err
}
```

### Specifying Fields

The new API requires explicit field specification. Here are common patterns:

```go
// Get only specific fields (recommended for performance)
opts := &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"summary", "status", "assignee", "priority"},
}

// Get all fields
opts := &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"*all"},
}

// Get navigable fields (similar to old default)
opts := &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"*navigable"},
}
```

## Migration Checklist

- [ ] Identify all uses of `Search()` method in your codebase
- [ ] Identify all uses of `NewSearchIterator()` in your codebase
- [ ] Replace `Search()` with `SearchJQL()`
- [ ] Replace `NewSearchIterator()` with `NewSearchJQLIterator()`
- [ ] Update pagination logic from `StartAt` to `NextPageToken`
- [ ] Add explicit field specifications (don't rely on defaults)
- [ ] Remove code that depends on `Total` count (not available in new API)
- [ ] Update tests to use new API
- [ ] Test thoroughly before October 2025 deadline

## Performance Considerations

### Advantages of New API
- **Higher page sizes**: Up to 5000 results per page
- **Better performance**: Token-based pagination is more efficient
- **Eventual consistency**: May see slightly stale data, but faster responses

### Best Practices
1. **Request only needed fields**: The more fields you request, the smaller the max page size
2. **Use appropriate page sizes**: Balance between number of requests and response size
3. **Handle pagination properly**: Always check `HasNextPage()` or test for empty `NextPageToken`

## Backward Compatibility

The SDK maintains backward compatibility by keeping the old `Search()` method available with deprecation warnings. However:

- The old method will **stop working** after October 31, 2025
- You should migrate as soon as possible
- The old method is marked as `Deprecated` in the API documentation

## Additional Resources

- [Atlassian Deprecation Notice](https://community.atlassian.com/t5/Jira-articles/Your-Jira-Scripts-and-Automations-May-Break-if-they-use-JQL/ba-p/3001235)
- [JQL Search API Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/)

## Need Help?

If you encounter issues during migration:
1. Check the examples above
2. Review the test files in `core/search/search_test.go`
3. Open an issue on the GitHub repository with your specific use case

## Breaking Changes in Next Major Version

In the next major version (v2.0.0) of this SDK:
- The deprecated `Search()` method will be removed
- The deprecated `SearchIterator` will be removed
- `SearchJQL()` will become the only search method

Migrate now to ensure a smooth transition!

---

# Expression Evaluation API Migration

## Overview

The `/rest/api/3/expression/eval` endpoint is being replaced with `/rest/api/3/expression/evaluate` as part of Atlassian's Enhanced Search API rollout.

## Timeline

- **October 31, 2024**: Deprecation announced
- **August 1, 2025**: Complete removal of `/rest/api/3/expression/eval` endpoint

## Key Differences

### Consistency Model
- **Old API (`/eval`)**: Strongly consistent - always returns the most up-to-date data
- **New API (`/evaluate`)**: Eventually consistent - may return slightly stale data for better performance

### Performance
- **Old API**: Slower but guaranteed consistency
- **New API**: Faster responses with improved scalability

### Use Cases
- **Old API**: When you need guaranteed up-to-date data
- **New API**: When performance is more important than immediate consistency

## Migration Examples

### Basic Expression Evaluation

**Old API (Deprecated):**
```go
result, err := client.Expression.Evaluate(ctx, &expression.EvaluationInput{
    Expression: "issue.summary",
    Context: map[string]interface{}{
        "issue": map[string]interface{}{
            "key": "PROJ-123",
        },
    },
})
```

**New API:**
```go
result, err := client.Expression.EvaluateExpression(ctx, &expression.EvaluationInput{
    Expression: "issue.summary",
    Context: map[string]interface{}{
        "issue": map[string]interface{}{
            "key": "PROJ-123",
        },
    },
})
```

### Complex Expression with User Context

**Old API (Deprecated):**
```go
result, err := client.Expression.Evaluate(ctx, &expression.EvaluationInput{
    Expression: "issue.fields.status.name + ' - ' + user.displayName",
    Context: map[string]interface{}{
        "issue": map[string]interface{}{
            "key": "PROJ-123",
        },
    },
})
```

**New API:**
```go
result, err := client.Expression.EvaluateExpression(ctx, &expression.EvaluationInput{
    Expression: "issue.fields.status.name + ' - ' + user.displayName",
    Context: map[string]interface{}{
        "issue": map[string]interface{}{
            "key": "PROJ-123",
        },
    },
})
```

### Handling Results

The input and output structures remain the same. No changes needed to result handling:

```go
// Works with both old and new APIs
result, err := client.Expression.EvaluateExpression(ctx, input)
if err != nil {
    return fmt.Errorf("evaluation failed: %w", err)
}

// Check for evaluation errors
if len(result.Errors) > 0 {
    for _, evalErr := range result.Errors {
        fmt.Printf("Error: %s at line %d, column %d\n",
            evalErr.Message, evalErr.Line, evalErr.Column)
    }
    return fmt.Errorf("expression has errors")
}

// Use the result value
fmt.Printf("Result: %v\n", result.Value)

// Check complexity metadata if needed
if result.Meta != nil && result.Meta.Complexity != nil {
    fmt.Printf("Complexity: %d steps, %d expensive operations\n",
        result.Meta.Complexity.Steps,
        result.Meta.Complexity.ExpensiveOperations)
}
```

## Migration Checklist for Expression API

- [ ] Identify all uses of `Evaluate()` method in your codebase
- [ ] Replace `Evaluate()` with `EvaluateExpression()`
- [ ] Understand eventual consistency implications for your use case
- [ ] Test expressions with new endpoint
- [ ] Migrate before August 1, 2025 deadline

## Important Considerations

### When Strong Consistency Matters

If your application requires **strongly consistent** data (e.g., financial calculations, audit trails, critical business logic), you should:

1. **Migrate before August 1, 2025** (no choice - old endpoint will be removed)
2. **Review your use cases** - Understand where eventual consistency is acceptable
3. **Add retry logic** if needed - For critical operations that might see stale data
4. **Consider caching** - If you're making frequent identical requests

### When Eventual Consistency is Acceptable

Most Jira expression evaluations can tolerate eventual consistency:
- Displaying issue summaries and status
- Computing user-friendly labels
- Generating reports (non-real-time)
- Automation triggers (can tolerate small delays)

## Performance Benefits

The new endpoint provides significant performance improvements:

- **Faster response times**: 30-50% improvement in typical scenarios
- **Better scalability**: Can handle higher request volumes
- **Reduced load**: More efficient use of Jira Cloud infrastructure

## Backward Compatibility

The SDK maintains backward compatibility:

- Old `Evaluate()` method still works with deprecation warnings
- Method will **stop working** after August 1, 2025
- No changes to input/output structures
- Migration is straightforward - just change the method name

## Additional Resources

- [Jira Expression Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-jira-expressions/)
- [Enhanced JQL Service Overview](https://community.atlassian.com/t5/Jira-articles/Avoiding-Pitfalls-A-Guide-to-Smooth-Migration-to-Enhanced-JQL/ba-p/2985433)

## Need Help?

If you encounter issues during migration:
1. Review the test files in `core/expression/expression_test.go`
2. Check the examples above
3. Open an issue on the GitHub repository

---

# Summary of All Changes

## Timeline Overview

Endpoint | Removal Date | New Endpoint
----------|--------------|--------------
`/rest/api/3/search` | October 31, 2025 | `/rest/api/3/search/jql`
`/rest/api/3/expression/eval` | August 1, 2025 | `/rest/api/3/expression/evaluate`

## Migration Priority

1. **Expression Evaluation** (August 1, 2025) - Higher priority due to earlier deadline
2. **Search API** (October 31, 2025) - More complex migration due to pagination changes

Both migrations should be completed as soon as possible to avoid service disruptions.
