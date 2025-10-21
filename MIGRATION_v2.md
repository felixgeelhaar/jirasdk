# Migration Guide: v1.3.0 to v1.4.0

This guide helps you migrate from jirasdk v1.3.0 to v1.4.0. Version 1.4.0 introduces breaking changes to align with Jira Cloud REST API v3 requirements for ADF (Atlassian Document Format).

## Breaking Changes Overview

Version 1.4.0 updates the following fields to use `*ADF` instead of `string`:

1. **Comment operations** - Comment bodies now use ADF format
2. **Issue link comments** - Link comments now use ADF format
3. **Environment field** - New field added with ADF format

## 1. Comment Operations

### What Changed

All comment-related structures now use `*ADF` for the body field:

- `Comment.Body`: `string` → `*ADF`
- `AddCommentInput.Body`: `string` → `*ADF`
- `UpdateCommentInput.Body`: `string` → `*ADF`

### Migration Path

#### Before (v1.x)

```go
// Adding a comment
input := &issue.AddCommentInput{
    Body: "This is my comment",
}
comment, err := client.Issue.AddComment(ctx, issueKey, input)

// Reading comment body
fmt.Println(comment.Body) // Direct string access
```

#### After (v1.4.0) - Recommended Approach

```go
// Adding a comment - use SetBodyText() convenience method
input := &issue.AddCommentInput{}
input.SetBodyText("This is my comment")
comment, err := client.Issue.AddComment(ctx, issueKey, input)

// Reading comment body - use GetBodyText() safe accessor
fmt.Println(comment.GetBodyText()) // Safe, never panics
```

#### After (v1.4.0) - Rich Formatting

```go
// Adding a rich formatted comment
input := &issue.AddCommentInput{}
adf := issue.NewADF().
    AddHeading("Status Update", 3).
    AddParagraph("Work completed successfully").
    AddBulletList([]string{"Feature A", "Feature B"})
input.SetBody(adf)

comment, err := client.Issue.AddComment(ctx, issueKey, input)
```

### Update Comments

#### Before (v1.x)

```go
updateInput := &issue.UpdateCommentInput{
    Body: "Updated comment text",
}
updated, err := client.Issue.UpdateComment(ctx, issueKey, commentID, updateInput)
```

#### After (v1.4.0)

```go
updateInput := &issue.UpdateCommentInput{}
updateInput.SetBodyText("Updated comment text")
updated, err := client.Issue.UpdateComment(ctx, issueKey, commentID, updateInput)
```

## 2. Issue Link Comments

### What Changed

`LinkComment.Body` now uses `*ADF` instead of `string`.

### Migration Path

#### Before (v1.x)

```go
err := client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
    Type:         issue.BlocksLinkType(),
    InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
    OutwardIssue: &issue.IssueRef{Key: "PROJ-456"},
    Comment: &issue.LinkComment{
        Body: "These issues are related",
    },
})
```

#### After (v1.4.0)

```go
linkComment := &issue.LinkComment{}
linkComment.SetBodyText("These issues are related")

err := client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
    Type:         issue.BlocksLinkType(),
    InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
    OutwardIssue: &issue.IssueRef{Key: "PROJ-456"},
    Comment:      linkComment,
})
```

## 3. Environment Field (New)

Version 2.0 adds the `Environment` field to `IssueFields`, which uses ADF format.

### Usage

#### Creating Issues with Environment

```go
fields := &issue.IssueFields{
    Project:   &issue.Project{Key: "PROJ"},
    Summary:   "Bug report",
    IssueType: &issue.IssueType{Name: "Bug"},
}
fields.SetEnvironmentText("Production - AWS us-east-1")

created, err := client.Issue.Create(ctx, &issue.CreateInput{Fields: fields})
```

#### Reading Environment

```go
// Get environment as plain text
env := issue.GetEnvironmentText()
fmt.Println(env)

// Get environment as ADF (for advanced use)
adf := issue.GetEnvironment()
if adf != nil && !adf.IsEmpty() {
    // Process ADF structure
}
```

## 4. Safe Accessor Methods

Version 2.0 introduces safe accessor methods to prevent nil pointer panics.

### Comment Accessors

#### Before (v1.x) - Unsafe

```go
// Could panic if Author or Created is nil!
authorName := comment.Author.DisplayName
createdTime := comment.Created.Format(time.RFC3339)
```

#### After (v1.4.0) - Safe

```go
// Safe accessors - never panic
authorName := comment.GetAuthorName()
if authorName == "" {
    authorName = "Unknown"
}

createdTime := comment.GetCreatedTime()
if !createdTime.IsZero() {
    fmt.Println(createdTime.Format(time.RFC3339))
}
```

### Available Safe Accessors for Comments

| Method | Returns | Description |
|--------|---------|-------------|
| `GetBody()` | `*ADF` | Returns ADF body (may be nil) |
| `GetBodyText()` | `string` | Extracts plain text from ADF (safe) |
| `GetAuthor()` | `*User` | Returns author (may be nil) |
| `GetAuthorName()` | `string` | Returns author name or empty string |
| `GetCreated()` | `*time.Time` | Returns created timestamp (may be nil) |
| `GetCreatedTime()` | `time.Time` | Returns created time or zero value |
| `GetUpdated()` | `*time.Time` | Returns updated timestamp (may be nil) |
| `GetUpdatedTime()` | `time.Time` | Returns updated time or zero value |

## 5. Validation Changes

Empty body validation now uses `IsEmpty()` instead of string comparison.

#### Before (v1.x)

```go
if input.Body == "" {
    return errors.New("body required")
}
```

#### After (v1.4.0)

```go
if input.Body.IsEmpty() {
    return errors.New("body required")
}
```

## 6. Step-by-Step Migration Checklist

### For Comment Operations

- [ ] Replace `Body: "text"` with `SetBodyText("text")` in `AddCommentInput`
- [ ] Replace `Body: "text"` with `SetBodyText("text")` in `UpdateCommentInput`
- [ ] Replace direct `comment.Body` access with `comment.GetBodyText()`
- [ ] Replace `comment.Author.DisplayName` with `comment.GetAuthorName()`
- [ ] Replace `*comment.Created` with `comment.GetCreatedTime()`
- [ ] Replace `*comment.Updated` with `comment.GetUpdatedTime()`

### For Issue Link Comments

- [ ] Create `LinkComment` instance first
- [ ] Use `SetBodyText()` to set the comment body
- [ ] Pass the comment to `CreateIssueLinkInput`

### For Environment Field

- [ ] Use `SetEnvironmentText()` when creating/updating issues with environment data
- [ ] Use `GetEnvironmentText()` when reading environment information

## 7. Common Migration Patterns

### Pattern 1: Simple Text Comments

```go
// v1.x
input := &issue.AddCommentInput{Body: "My comment"}

// v1.4.0
input := &issue.AddCommentInput{}
input.SetBodyText("My comment")
```

### Pattern 2: Reading Comment Metadata

```go
// v1.x (unsafe)
for _, c := range comments {
    fmt.Printf("%s: %s\n", c.Author.DisplayName, c.Body)
}

// v1.4.0 (safe)
for _, c := range comments {
    author := c.GetAuthorName()
    if author == "" {
        author = "Unknown"
    }
    fmt.Printf("%s: %s\n", author, c.GetBodyText())
}
```

### Pattern 3: Time Handling

```go
// v1.x (unsafe)
if comment.Created != nil {
    fmt.Println(comment.Created.Format(time.RFC3339))
}

// v1.4.0 (safe)
created := comment.GetCreatedTime()
if !created.IsZero() {
    fmt.Println(created.Format(time.RFC3339))
}
```

## 8. Testing Your Migration

After migrating, ensure:

1. **Compile check**: Your code compiles without errors
2. **Test coverage**: Run your test suite to catch runtime issues
3. **Integration tests**: Test against a real Jira instance if possible
4. **Error handling**: Verify error handling for empty/invalid inputs

## 9. Need Help?

- Check the [examples](./examples) directory for complete working examples
- Review the [comment example](./examples/comments/main.go) for comprehensive comment usage
- See the [issue links example](./examples/issuelinks/main.go) for link comment usage
- Read the main [README](./README.md) for API overview

## 10. What's Next?

After migrating to v2.0:

- Consider using rich ADF formatting for better-formatted comments
- Use the new Environment field for bug reports and production issues
- Take advantage of safe accessor methods to prevent nil pointer panics
- Review the examples for additional patterns and best practices
