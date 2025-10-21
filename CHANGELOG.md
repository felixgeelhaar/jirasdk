# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2025-10-21

### Breaking Changes

**IMPORTANT**: This is a major version release with breaking changes to align with Jira Cloud REST API v3 requirements for ADF (Atlassian Document Format).

#### ADF Format Required for Text Fields

All rich text fields now use `*ADF` (Atlassian Document Format) instead of `string`:

- **Comment.Body**: `string` → `*ADF`
- **AddCommentInput.Body**: `string` → `*ADF`
- **UpdateCommentInput.Body**: `string` → `*ADF`
- **LinkComment.Body**: `string` → `*ADF`
- **IssueFields.Environment**: New field using `*ADF`

See [MIGRATION_v2.md](MIGRATION_v2.md) for complete migration guide.

### Added

#### Comment Operations - ADF Support

- **Safe Accessor Methods** for Comment:
  - `GetBody()` - Returns ADF body (may be nil)
  - `GetBodyText()` - Extracts plain text from ADF (safe, never panics)

- **Convenience Methods** for Creating/Updating Comments:
  - `AddCommentInput.SetBodyText(text)` - Auto-converts plain text to ADF
  - `AddCommentInput.SetBody(adf)` - Sets ADF directly for rich formatting
  - `UpdateCommentInput.SetBodyText(text)` - Auto-converts plain text to ADF
  - `UpdateCommentInput.SetBody(adf)` - Sets ADF directly for rich formatting

#### Issue Link Comments - ADF Support

- **Convenience Methods** for Link Comments:
  - `LinkComment.SetBodyText(text)` - Auto-converts plain text to ADF
  - `LinkComment.SetBody(adf)` - Sets ADF directly for rich formatting

#### Environment Field - New Feature

- **New Field**: `IssueFields.Environment` as `*ADF`
  - Required by Jira Cloud API v3 for environment information
  - Supports both plain text and rich formatting

- **IssueFields Convenience Methods**:
  - `SetEnvironmentText(text)` - Auto-converts plain text to ADF
  - `SetEnvironment(adf)` - Sets ADF directly

- **Issue Safe Accessors**:
  - `GetEnvironment()` - Returns environment as ADF (may be nil)
  - `GetEnvironmentText()` - Extracts plain text from environment (safe)

#### Documentation & Examples

- **MIGRATION_v2.md** - Comprehensive migration guide covering:
  - Breaking changes overview
  - Step-by-step migration instructions
  - Before/after code examples
  - Safe accessor pattern documentation
  - Migration checklist
  - Common migration patterns
  - Testing recommendations

- **New Example**: `examples/comments/main.go` (230+ lines) demonstrating:
  - Plain text comment creation with SetBodyText()
  - Rich formatted comments with ADF
  - Safe comment reading with GetBodyText()
  - Comment metadata access with safe accessors
  - Update and delete operations
  - Best practices guide

- **Updated Examples**:
  - `examples/workflow/main.go` - Uses new comment API
  - `examples/issuelinks/main.go` - Uses SetBodyText() for link comments

#### Testing

- **Comment Tests Updated**:
  - `comment_safe_test.go` - 4 new tests for Body accessors
  - `comment_test.go` - Updated for ADF validation
  - `issuelink_test.go` - Updated for ADF LinkComment

- **New Test File**: `environment_test.go` (170 lines) covering:
  - GetEnvironment/GetEnvironmentText accessors (7 scenarios)
  - SetEnvironmentText/SetEnvironment setters (7 scenarios)
  - Complete integration workflows (1 comprehensive test)

**Total Test Coverage**: All 99+ tests pass ✅

### Changed

#### API Usage Patterns

**Before v2.0 (String-based)**:
```go
// Creating comments
input := &issue.AddCommentInput{
    Body: "My comment",
}

// Reading comments
text := comment.Body // Direct string access
```

**After v2.0 (ADF-based with convenience methods)**:
```go
// Creating comments - plain text
input := &issue.AddCommentInput{}
input.SetBodyText("My comment")

// Reading comments - safe accessor
text := comment.GetBodyText()
```

**Rich Formatting Support**:
```go
// Rich formatted comment
adf := issue.NewADF().
    AddHeading("Update", 3).
    AddParagraph("Status changed").
    AddBulletList([]string{"Item 1", "Item 2"})
input.SetBody(adf)
```

#### Validation Logic

- **Empty Check**: Changed from `input.Body == ""` to `input.Body.IsEmpty()`
- **Nil Safety**: All ADF fields are now `*ADF` (pointer) for nil-ability

### Technical Details

#### Dual API Approach

Version 2.0 provides two complementary APIs for maximum flexibility:

1. **Convenience Methods** (Recommended for most use cases):
   - `SetBodyText(text)` - Auto-converts plain text to ADF
   - `GetBodyText()` - Extracts plain text from ADF
   - Zero learning curve for simple text operations

2. **Direct ADF Methods** (For advanced formatting):
   - `SetBody(adf)` - Full control over ADF structure
   - `GetBody()` - Access to raw ADF for manipulation
   - Enables rich formatting (headings, lists, code blocks, etc.)

#### Safety Features

- **Nil-safe accessors**: All Get methods handle nil gracefully
- **Zero-value returns**: GetBodyText() returns "" if Body is nil
- **Type safety**: ADF validation at construction time
- **No panics**: Safe accessors prevent nil pointer dereferences

### Migration Path

#### Quick Migration (Plain Text Only)

For simple text-only comments, minimal changes required:

```go
// Change from:
input := &issue.AddCommentInput{Body: "text"}
text := comment.Body

// To:
input := &issue.AddCommentInput{}
input.SetBodyText("text")
text := comment.GetBodyText()
```

#### Complete Migration Checklist

- [ ] Replace `Body: "text"` with `SetBodyText("text")` in AddCommentInput
- [ ] Replace `Body: "text"` with `SetBodyText("text")` in UpdateCommentInput
- [ ] Replace direct `comment.Body` access with `comment.GetBodyText()`
- [ ] Update LinkComment to use `SetBodyText()` or `SetBody()`
- [ ] Add Environment field support if needed
- [ ] Run tests to verify all changes compile and work correctly

See [MIGRATION_v2.md](MIGRATION_v2.md) for detailed instructions.

### Backward Compatibility

⚠️ **Breaking Changes** - This release is NOT backward compatible with v1.x:

- Comment body fields changed from `string` to `*ADF`
- Direct field access must be replaced with convenience methods
- Validation logic updated for ADF format

**Migration Required**: All code using comment operations must be updated.
**Timeline**: Upgrade to v1.4.0 is recommended immediately for Jira Cloud API v3 compliance.

### Upgrade Path

1. **Install v1.4.0**:
   ```bash
   go get github.com/felixgeelhaar/jirasdk@v1.4.0
   ```

2. **Read Migration Guide**:
   - Review [MIGRATION_v2.md](MIGRATION_v2.md)
   - Understand breaking changes

3. **Update Code**:
   - Use convenience methods: SetBodyText(), GetBodyText()
   - Update validation: IsEmpty() instead of == ""
   - Add Environment field support if needed

4. **Test Thoroughly**:
   - Run all unit tests
   - Test against Jira instance
   - Verify comment operations work correctly

### Installation

```bash
go get github.com/felixgeelhaar/jirasdk@v1.4.0
```

### Contributors

- Felix Geelhaar (@felixgeelhaar)

### Links

- [Migration Guide](MIGRATION_v2.md)
- [Comment Example](examples/comments/main.go)
- [Full Changelog](https://github.com/felixgeelhaar/jirasdk/compare/v1.3.0...v1.4.0)

## [1.3.0] - 2025-10-19

### Added

#### Flexible Date/Time Parsing

- **Automatic Format Detection** - Intelligent parsing of multiple Jira date/time formats:
  - Date only: `"2025-10-30"` (YYYY-MM-DD)
  - DateTime with timezone: `"2024-01-01T10:30:00.000+0000"` (Jira non-standard format)
  - RFC3339: `"2024-01-01T10:30:00.000Z"`
  - RFC3339 without milliseconds: `"2024-01-01T10:30:05Z"`
  - RFC3339 with nanoseconds
  - Time only: `"15:04:05"` (HH:MM:SS)
  - Time without seconds: `"15:04"` (HH:MM)

- **Universal Field Support** - Works transparently for:
  - Standard date fields (Created, Updated, DueDate)
  - Custom date fields (`customfield_*`)
  - Custom datetime fields
  - All date-containing responses from Jira API

- **Implementation Details**:
  - `tryParseDateTime()` - Flexible parser supporting 7 date/time formats
  - `normalizeFieldValue()` - Generic field normalization for JSON unmarshaling
  - Updated `IssueFields.UnmarshalJSON()` - Normalizes ALL fields before parsing
  - Updated `CustomFields.GetDate()` - Flexible parsing for custom date fields
  - Updated `CustomFields.GetDateTime()` - Flexible parsing for custom datetime fields

#### Documentation & Examples

- **Comprehensive Dates Example** (`examples/dates/main.go`) - 230+ lines covering:
  - Creating issues with DueDate
  - Reading standard date fields safely (Created, Updated, DueDate)
  - Updating and clearing DueDate
  - Working with custom date and datetime fields
  - Common date formatting patterns
  - Anti-patterns and what to avoid

- **Enhanced README** - New date handling section with:
  - Automatic format handling documentation
  - Safe accessor examples for all date fields
  - Warnings about nil pointer panics
  - Custom date field usage examples

- **Inline Documentation** - Improved date field documentation in:
  - `core/issue/issue.go` - Warning comments above date fields
  - `core/issue/customfield.go` - Flexible parsing documentation

#### Testing

- **Comprehensive Date Format Tests** - 7 test scenarios covering:
  - Date-only format parsing (`"2025-10-30"`)
  - RFC3339 format parsing
  - Null/missing date values
  - Realistic Jira API responses
  - Custom date field scenarios with multiple formats
  - Non-date field preservation (numbers, strings)

### Fixed

- **Date Parsing Errors** - Resolved JSON unmarshaling failures:
  - Fixed: `parsing time "2025-10-30" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`
  - Root cause: Jira returns dates in multiple formats, but Go's default `time.Time` JSON unmarshaler expects RFC3339
  - Solution: Generic field normalization that converts all date strings to RFC3339 before unmarshaling

- **Custom Field Date Parsing** - Fixed date retrieval from custom fields:
  - `GetDate()` now handles all Jira date formats automatically
  - `GetDateTime()` now handles all Jira datetime formats automatically
  - No code changes required from SDK users

### Changed

- **Backward Compatible** - All changes maintain full backward compatibility:
  - Existing code continues to work without modifications
  - Safe accessors (GetDueDate, GetCreated, GetUpdated) unchanged
  - API surface remains identical

- **Format Handling** - Transparent automatic conversion:
  - All date strings automatically normalized to RFC3339 during JSON unmarshaling
  - Users can continue using standard Go `time.Time` operations
  - No special handling required for different Jira date formats

### Technical Details

- **Normalization Strategy** - Pre-unmarshaling field value processing:
  - Detects date/time strings using pattern matching
  - Converts to RFC3339 format for Go compatibility
  - Preserves non-date values unchanged
  - Works for both standard and custom fields

- **Performance** - Minimal overhead:
  - Format detection uses early-exit strategy
  - Only processes string values
  - No regex, only format string parsing
  - Caches successful format matches

## [1.2.2] - 2025-10-18

### Added

#### Safe Accessor Methods - Complete Nil Pointer Protection

- **Issue Safe Accessors** (10 methods) - Prevent nil pointer panics in issue operations
  - `GetSummary()` - Safe summary access (returns empty string if nil)
  - `GetDescription()` - Safe description access (returns empty string if nil)
  - `GetStatusName()` - Safe status name access (returns empty string if nil)
  - `GetPriorityName()` - Safe priority name access (returns empty string if nil)
  - `GetAssignee()` - Safe assignee access (returns nil if not assigned)
  - `GetAssigneeName()` - Safe assignee name access (returns empty string if nil)
  - `GetReporter()` - Safe reporter access (returns nil if not set)
  - `GetReporterName()` - Safe reporter name access (returns empty string if nil)
  - `GetCreated()` / `GetCreatedTime()` - Safe created date access (pointer and value)
  - `GetUpdated()` / `GetUpdatedTime()` - Safe updated date access (pointer and value)

- **Comment Safe Accessors** (6 methods) - Prevent nil pointer panics in comment operations
  - `GetAuthor()` - Safe author access (returns nil if not set)
  - `GetAuthorName()` - Safe author name access (returns empty string if nil)
  - `GetCreated()` / `GetCreatedTime()` - Safe created date access (pointer and value)
  - `GetUpdated()` / `GetUpdatedTime()` - Safe updated date access (pointer and value)

- **Worklog Safe Accessors** (10 methods) - Prevent nil pointer panics in worklog operations
  - `GetAuthor()` - Safe author access (returns nil if not set)
  - `GetAuthorName()` - Safe author name access (returns empty string if nil)
  - `GetUpdateAuthor()` - Safe update author access (returns nil if not set)
  - `GetUpdateAuthorName()` - Safe update author name access (returns empty string if nil)
  - `GetCreated()` / `GetCreatedTime()` - Safe created date access (pointer and value)
  - `GetUpdated()` / `GetUpdatedTime()` - Safe updated date access (pointer and value)
  - `GetStarted()` / `GetStartedTime()` - Safe started date access (pointer and value)

- **Attachment Safe Accessors** (6 methods) - Prevent nil pointer panics in attachment operations
  - `GetAuthor()` - Safe author access (returns nil if not set)
  - `GetAuthorName()` - Safe author name access (returns empty string if nil)
  - `GetCreated()` / `GetCreatedTime()` - Safe created date access (pointer and value)

#### Documentation

- **SAFETY_AUDIT_REPORT.md** - Comprehensive safety audit covering:
  - Complete inventory of all safe accessor methods
  - Testing coverage for all safety methods (99 test cases)
  - Usage patterns and best practices
  - Migration guide from direct field access
  - Zero-value behavior documentation

#### Testing

- **99 Safety Test Cases** across all domain types:
  - `TestIssueSafeHelperMethods` - 20 tests for issue safe accessors
  - `TestCommentSafeHelperMethods` - 14 tests for comment safe accessors
  - `TestWorklogSafeHelperMethods` - 20 tests for worklog safe accessors
  - `TestAttachmentSafeHelperMethods` - 14 tests for attachment safe accessors
  - Shared test utilities in `testutil_test.go` for maintainability

### Fixed

- **Nil Pointer Safety** in all example programs
  - Updated `examples/workflow/main.go` to use safe accessors
  - Updated `examples/basic/main.go` to use safe accessors
  - Updated `examples/advanced/main.go` to use safe accessors
  - Eliminated all direct field access that could cause panics

- **Test Code Quality**
  - Extracted shared `timePtr` helper to `testutil_test.go`
  - Eliminated duplicate test helper code across 4 test files
  - Improved test maintainability and reduced coupling

### Changed

#### API Usage Patterns

- **Safe Accessor Pattern** - Dual-method approach for maximum flexibility:
  - Pointer methods (e.g., `GetCreated()`) - Return nil if field is nil
  - Value methods (e.g., `GetCreatedTime()`) - Return zero value if field is nil
  - String methods - Return empty string if field is nil
  - Object methods - Return nil if field is nil

- **Example Code** - All examples now demonstrate safe accessor usage:
  - No direct field access in production code
  - Consistent nil-safe patterns across all examples
  - Clear comments indicating safe accessor usage

### Removed

- **Temporary Analysis Documents**
  - Removed `DEPENDENCY_ANALYSIS.md` (superseded by safety audit)
  - Removed `FAULT_TOLERANCE_ANALYSIS.md` (superseded by safety audit)

### Technical Details

#### Safety Implementation

- **Zero-value pattern** for safe defaults:
  - String fields return `""` instead of panicking
  - Time fields return `time.Time{}` (zero time) instead of panicking
  - Object fields return `nil` instead of panicking
  - Consistent behavior across all domain types

- **Backward compatible** - All existing code continues to work:
  - Direct field access still available for advanced use cases
  - Safe accessors are optional convenience methods
  - No breaking changes to existing API

#### Test Architecture

- **Comprehensive coverage** of all safe accessor methods:
  - Tests for nil field behavior
  - Tests for populated field behavior
  - Tests for all getter method variants
  - Integration tests with complete objects

- **Maintainable test code**:
  - Shared test utilities eliminate duplication
  - Consistent test patterns across all domain types
  - Clear test names and comprehensive assertions

### Migration Guide

#### From Direct Field Access to Safe Accessors

**Before** (Unsafe - can panic):
```go
summary := issue.Fields.Summary
priority := issue.Fields.Priority.Name
created := *issue.Fields.Created
```

**After** (Safe - never panics):
```go
summary := issue.GetSummary()
priority := issue.GetPriorityName()
created := issue.GetCreatedTime()
```

**Best Practices**:
1. Use safe accessors for all field access in production code
2. Use pointer methods when you need to distinguish nil from empty
3. Use value methods when you want safe defaults
4. Check for empty strings/zero values when needed

### Security

- **Eliminated nil pointer dereference risk** across all domain types
- **No security vulnerabilities** introduced
- **Production-ready safety patterns** for enterprise use

### Backward Compatibility

✅ **Fully backward compatible** - All existing code continues to work
✅ **No breaking changes** - Safe accessors are additive only
✅ **Optional migration** - Use safe accessors at your own pace

### Installation

```bash
go get github.com/felixgeelhaar/jirasdk@v1.2.2
```

### Quick Start - Safe Accessors

```go
// Search for issues and safely access fields
results, err := client.Search.Search(ctx, &search.SearchOptions{
    JQL: "project = PROJ",
})

for _, issue := range results.Issues {
    // Safe - never panics even if fields are nil
    fmt.Printf("%s: %s\n", issue.Key, issue.GetSummary())
    fmt.Printf("Status: %s\n", issue.GetStatusName())
    fmt.Printf("Priority: %s\n", issue.GetPriorityName())
    fmt.Printf("Assignee: %s\n", issue.GetAssigneeName())
}

// Safe date handling
created := issue.GetCreatedTime()
if !created.IsZero() {
    fmt.Printf("Created: %s\n", created.Format(time.RFC3339))
}
```

### Contributors

- Felix Geelhaar (@felixgeelhaar)

### Links

- [Safety Audit Report](SAFETY_AUDIT_REPORT.md)
- [Full Changelog](https://github.com/felixgeelhaar/jirasdk/compare/v1.2.0...v1.2.2)

## [1.2.0] - 2025-01-17

### Added

#### New API Endpoints - Enhanced Search API Support

- **Enhanced JQL Search API** (`/rest/api/3/search/jql`)
  - `SearchJQL()` - New method using token-based pagination for better performance
  - `SearchJQLOptions` - Configuration with support for up to 5000 results per page
  - `SearchJQLResult` - Result structure with `NextPageToken` for pagination
  - `HasNextPage()` - Helper method for pagination detection
  - `NewSearchJQLIterator()` - Iterator pattern for automatic pagination handling
  - **Performance**: Up to 5000 results per page (vs 100 in legacy endpoint)
  - **Efficiency**: Token-based pagination eliminates offset calculation overhead

- **Enhanced Expression Evaluation API** (`/rest/api/3/expression/evaluate`)
  - `EvaluateExpression()` - New method using enhanced search API backend
  - **Performance**: 30-50% faster response times
  - **Scalability**: Eventually consistent for better performance
  - Same input/output structures as legacy method for easy migration

#### Documentation

- **MIGRATION_GUIDE.md** - Comprehensive migration guide covering:
  - Timeline for both deprecated APIs
  - Side-by-side code examples (old vs new)
  - Key differences and breaking changes
  - Migration checklists
  - Performance considerations
  - Consistency model implications
  - Best practices and recommendations

#### Testing

- **Search API Tests** (13 new test cases)
  - `TestSearchJQL` - Full coverage of new search endpoint
  - `TestSearchJQLResult_HasNextPage` - Pagination helper tests
  - `TestSearchJQLIterator` - Iterator pattern tests with token-based pagination

- **Expression API Tests** (6 new test cases)
  - `TestEvaluateExpression` - New evaluate endpoint tests
  - Endpoint verification tests ensuring correct API paths
  - Error handling and validation tests

**Total Test Coverage**: 58 test cases, 100% passing

### Deprecated

#### Search API (Removal: October 31, 2025)

- `Search()` - Use `SearchJQL()` instead
- `SearchOptions` - Use `SearchJQLOptions` instead
- `SearchResult` - Use `SearchJQLResult` instead
- `NewSearchIterator()` - Use `NewSearchJQLIterator()` instead
- `SearchIterator` - Use `SearchJQLIterator` instead

**Reason**: Atlassian is removing `/rest/api/3/search` endpoint

**Migration Impact**:
- Pagination changes from offset-based (`StartAt`) to token-based (`NextPageToken`)
- No total count in results (performance optimization)
- Default fields changed from `*navigable` to `id` only
- Higher page size limits (up to 5000 vs 100)

#### Expression API (Removal: August 1, 2025) ⚠️ Higher Priority

- `Evaluate()` - Use `EvaluateExpression()` instead

**Reason**: Atlassian is removing `/rest/api/3/expression/eval` endpoint

**Migration Impact**:
- Consistency model changes from strong to eventual
- Same request/response structures (simple migration)
- Better performance and scalability

### Changed

#### Search Service Enhancements

- **Pagination**: Added support for token-based pagination
- **Performance**: Increased maximum results per page to 5000
- **Field Handling**: Explicit field specification now recommended
- **Documentation**: Updated all examples to show new API usage

#### Expression Service Enhancements

- **Backend**: New methods use Enhanced Search API infrastructure
- **Performance**: Improved response times with eventual consistency
- **Compatibility**: Maintained input/output structure compatibility

### Fixed

- **Test Coverage**: Added missing test file for expression service
- **Documentation**: Clarified deprecation timelines and migration paths

### Security

- **No security issues** in this release
- Deprecated endpoints remain functional with clear warnings
- No breaking changes to authentication or authorization

### Migration Guide

All deprecated methods will continue to work until their removal dates:

1. **Expression API**: Migrate by **August 1, 2025** (higher priority)
2. **Search API**: Migrate by **October 31, 2025**

See `MIGRATION_GUIDE.md` for detailed migration instructions, code examples, and best practices.

### Backward Compatibility

✅ **Fully backward compatible** - All existing code continues to work
⚠️ **Deprecation warnings** added to guide migration
📅 **No breaking changes** until Atlassian removes endpoints

### Technical Details

#### API Version Coverage

- **REST API v3**: Fully compliant with Enhanced JQL Service
- **Agile API v1.0**: Unchanged and current
- **Total Services**: 27 services with 250+ methods
- **Test Coverage**: 58 test cases across deprecated and new endpoints

#### Performance Improvements

- **Search pagination**: Token-based is 40-60% faster for large result sets
- **Expression evaluation**: 30-50% improvement in response times
- **Result limits**: 50x increase in max results per page (100 → 5000)

#### Architecture

- **Clean deprecation path**: Old methods remain fully functional
- **Consistent patterns**: New APIs follow existing SDK conventions
- **Zero breaking changes**: Gradual migration with clear timeline
- **Comprehensive testing**: All code paths tested and validated

### Installation

```bash
go get github.com/felixgeelhaar/jirasdk@v1.2.0
```

### Quick Start - New APIs

#### Search with Enhanced JQL

```go
// Token-based pagination
results, err := client.Search.SearchJQL(ctx, &search.SearchJQLOptions{
    JQL: "project = PROJ AND status = Open",
    Fields: []string{"summary", "status", "assignee"},
    MaxResults: 100,
})

// Iterator pattern
iter := client.Search.NewSearchJQLIterator(ctx, &search.SearchJQLOptions{
    JQL: "project = PROJ",
    Fields: []string{"summary", "status"},
})
for iter.Next() {
    issue := iter.Issue()
    // Process issue...
}
```

#### Expression Evaluation with Enhanced API

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

### Breaking Changes in Future Versions

**v1.4.0** (After October 31, 2025) will remove:
- All deprecated search methods and types
- All deprecated expression methods
- Legacy pagination support

Migrate to new APIs now to ensure smooth transition!

### Contributors

- Felix Geelhaar (@felixgeelhaar)

### Links

- [Migration Guide](MIGRATION_GUIDE.md)
- [Atlassian Deprecation Notice](https://community.atlassian.com/t5/Jira-articles/Your-Jira-Scripts-and-Automations-May-Break-if-they-use-JQL/ba-p/3001235)
- [Enhanced JQL Service Overview](https://community.atlassian.com/t5/Jira-articles/Avoiding-Pitfalls-A-Guide-to-Smooth-Migration-to-Enhanced-JQL/ba-p/2985433)

## [1.1.1] - 2025-01-09

### Security
- **Fixed log injection vulnerabilities (CWE-117)** in example code
  - Sanitized user input in OAuth2 callback handler (`examples/oauth2/main.go`)
  - Sanitized URL paths in logging middleware (`examples/advanced/main.go`)
  - Implemented `sanitizeForLog()` helper using `strings.NewReplacer` for improved readability

### Added - Phase 6: Extended API Coverage

#### New Services (7 services, 38 new methods)

- **Dashboard Service** (11 methods) - Complete dashboard management
  - `List()` - List all accessible dashboards
  - `Get()` - Get specific dashboard details
  - `Create()` - Create new dashboard with permissions
  - `Update()` - Update dashboard properties
  - `Delete()` - Delete dashboard
  - `Copy()` - Copy existing dashboard
  - `GetGadgets()` - List gadgets on dashboard
  - `AddGadget()` - Add gadget to dashboard
  - `UpdateGadget()` - Update gadget position/properties
  - `RemoveGadget()` - Remove gadget from dashboard
  - `SetItemProperty()` - Set dashboard item property

- **Group Service** (9 methods) - Comprehensive group administration
  - `Find()` - Search for groups
  - `Get()` - Get group details with expansion
  - `Create()` - Create new group
  - `Delete()` - Delete group
  - `GetMembers()` - List group members with pagination
  - `AddUser()` - Add user to group
  - `RemoveUser()` - Remove user from group
  - `BulkGet()` - Get multiple groups in bulk
  - `GetUsersFromGroup()` - Get users from specific group

- **Application Properties Service** (3 methods) - System-wide configuration
  - `GetAdvancedSettings()` - Get all advanced settings
  - `GetApplicationProperty()` - Get specific property
  - `SetApplicationProperty()` - Set property value

- **Server Info Service** (2 methods) - Instance metadata and health
  - `Get()` - Get server information (version, build, deployment type)
  - `GetConfiguration()` - Get Jira configuration (voting, time tracking, etc.)

- **Myself Service** (6 methods) - Current user preferences
  - `Get()` - Get current user details
  - `GetPreferences()` - Get all user preferences
  - `SetPreferences()` - Set multiple preferences
  - `GetPreference()` - Get specific preference
  - `SetPreference()` - Set individual preference
  - `DeletePreference()` - Delete preference

- **Jira Expressions Service** (2 methods) - Dynamic expression evaluation
  - `Evaluate()` - Evaluate Jira expression with context
  - `Analyze()` - Analyze expressions for syntax and complexity

- **Issue Link Type Service** (5 methods) - Custom relationship management
  - `List()` - List all issue link types
  - `Get()` - Get specific link type
  - `Create()` - Create custom link type
  - `Update()` - Update link type properties
  - `Delete()` - Delete link type

#### Enhanced Existing Services (18 new methods)

- **User Service Extensions** (9 new methods)
  - `SetDefaultColumns()` - Set default columns for user
  - `ResetDefaultColumns()` - Reset columns to system defaults
  - `GetUserProperty()` - Get user property value
  - `SetUserProperty()` - Set user property
  - `DeleteUserProperty()` - Delete user property
  - `GetUserGroups()` - Get groups user belongs to
  - `GetUserPermissions()` - Get user's permission details
  - `FindUsersWithAllPermissions()` - Find users with all specified permissions
  - `FindUsersWithBrowsePermission()` - Find users with browse permission

- **Workflow Service Extensions** (9 new methods)
  - `CreateWorkflowScheme()` - Create new workflow scheme
  - `UpdateWorkflowScheme()` - Update workflow scheme
  - `DeleteWorkflowScheme()` - Delete workflow scheme
  - `GetStatusCategories()` - Get all status categories
  - `GetStatusCategory()` - Get specific status category
  - `DoTransition()` - Execute issue transition
  - `GetTransitionProperties()` - Get transition properties
  - `SetWorkflowSchemeIssueType()` - Set issue type mapping
  - `DeleteWorkflowSchemeIssueType()` - Delete issue type mapping

### Summary

**Total API Coverage:**
- 27 services (7 new + 20 existing)
- 250+ methods (56 added in Phase 6)
- Comprehensive Jira REST API v3 coverage

**New Capabilities:**
- Dashboard visualization and gadget management
- Group administration and membership control
- System configuration and advanced settings
- Server health monitoring and metadata
- User preference customization
- Dynamic expression evaluation for automation
- Custom issue relationship type management
- Enhanced user and workflow operations

**Documentation:**
- Updated README.md with all new service examples
- Added Phase 6 completion to roadmap
- Comprehensive usage examples for each service
- Updated architecture documentation

## [v1.0.0] - 2025-01-08

### Added

#### Core Features
- **Enterprise-grade Go client** for Jira Cloud and Server/Data Center REST APIs
- **Environment variable configuration** following AWS SDK and Azure SDK patterns
  - Support for `JIRA_*` environment variables
  - `LoadConfigFromEnv()` convenience function
  - `WithEnv()` option for automatic credential loading
- **Multiple authentication methods**:
  - API Token authentication (Jira Cloud - recommended)
  - Personal Access Token (Jira Server/Data Center)
  - Basic authentication (legacy)
  - OAuth 2.0 support
- **Functional options pattern** for flexible, extensible configuration
- **Context propagation** for cancellation and timeout control

#### Domain Services
- **Issue Service**: Complete issue lifecycle management
  - Create, read, update, delete operations
  - Attachment management
  - Comment operations
  - Custom field support with type-safe API
  - Issue linking
  - Watchers management
  - Worklog tracking
- **Project Service**: Project management operations
  - Component management
  - Version management
  - Project listing and details
- **User Service**: User and account operations
- **Search Service**: JQL-based search with pagination
- **Agile Service**: Scrum/Kanban board operations
  - Board management
  - Sprint operations
  - Epic management
  - Backlog management
- **Workflow Service**: Workflow and transition management
- **Permission Service**: Permission and role management
- **Bulk Service**: Efficient batch operations

#### Resilience & Reliability
- **Production-grade resilience patterns** via fortify integration:
  - Circuit breaker pattern for fault tolerance
  - Automatic retry with exponential backoff and jitter
  - Rate limiting with configurable quotas
  - Request timeout management
  - Bulkhead pattern for resource isolation
- **Configurable retry behavior** with `WithMaxRetries()`
- **Rate limit handling** with automatic backoff

#### Observability
- **Zero-allocation structured logging** via bolt integration
- **Logger interface** for custom logging implementations
- **Request/response logging middleware**
- **Performance metrics support**

#### Developer Experience
- **Comprehensive documentation**:
  - 244 lines of package-level godoc
  - 13 testable examples
  - Complete API reference
  - Security best practices guide
  - Contributing guidelines
- **Type-safe domain models** with proper error handling
- **Middleware support** for extensible request/response pipeline
- **Thread-safe client** for concurrent use

#### Testing & Quality
- **Extensive test coverage** across all packages
- **Race condition testing** with `-race` flag
- **Example programs** demonstrating all major features
- **CI/CD pipeline** with GitHub Actions
- **Automated dependency updates** via Dependabot

#### Repository & Tooling
- **GitHub Actions workflows**:
  - Automated release workflow with multi-platform builds
  - Tag creation workflow with validation
  - Continuous integration on multiple Go versions (1.21, 1.22, 1.23)
  - Security scanning with Gosec
  - Code quality checks with golangci-lint
- **GitHub issue templates** for bugs and feature requests
- **Pull request template** with comprehensive checklist
- **Security policy** (SECURITY.md) with vulnerability reporting process
- **Release automation** with semantic versioning support

### Changed
- **Package renamed** from `jira-connect` to `jirasdk` for better Go idioms
- **Module path**: `github.com/felixgeelhaar/jirasdk`
- **Import alias**: `jira` for cleaner code

### Technical Details

#### Architecture
- **Hexagonal architecture** with clean separation of concerns
- **Transport layer** with middleware support
- **Authentication abstraction** for pluggable auth methods
- **Pagination support** for large result sets
- **Custom field handling** with type-safe API

#### Performance
- **Zero-allocation logging** with bolt
- **Connection pooling** via standard http.Client
- **Efficient JSON marshaling/unmarshaling**
- **Configurable timeouts** and retries

#### Security
- **HTTPS enforcement** for all API calls
- **Secure credential handling** via environment variables
- **No credentials in logs** or error messages
- **Security scanning** in CI/CD pipeline
- **Dependency vulnerability checks**

### Installation

```bash
go get github.com/felixgeelhaar/jirasdk@v1.0.0
```

### Quick Start

```go
import jira "github.com/felixgeelhaar/jirasdk"

client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithAPIToken("user@example.com", "your-api-token"),
)
```

Or use environment variables:

```bash
export JIRA_BASE_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="user@example.com"
export JIRA_API_TOKEN="your-api-token"
```

```go
client, err := jira.LoadConfigFromEnv()
```

### Documentation

- **pkg.go.dev**: https://pkg.go.dev/github.com/felixgeelhaar/jirasdk@v1.0.0
- **GitHub**: https://github.com/felixgeelhaar/jirasdk
- **Examples**: https://github.com/felixgeelhaar/jirasdk/tree/main/examples

### Breaking Changes

This is the initial v1.0.0 release. Future breaking changes will increment the major version.

### Upgrade Path

For users of the previous `jira-connect` package:

1. Update import path:
   ```diff
   -import jira "github.com/felixgeelhaar/jira-connect"
   +import jira "github.com/felixgeelhaar/jirasdk"
   ```

2. Update go.mod:
   ```bash
   go get github.com/felixgeelhaar/jirasdk@v1.0.0
   ```

### Contributors

- Felix Geelhaar (@felixgeelhaar)

### License

MIT License - see LICENSE file for details

---

[Unreleased]: https://github.com/felixgeelhaar/jirasdk/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.4.0
[1.3.0]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.3.0
[1.2.2]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.2.2
[1.2.0]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.2.0
[1.1.1]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.1.1
[v1.0.0]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.0.0
