# Observability

`jirasdk` does not emit OpenTelemetry spans itself; the `go.opentelemetry.io/otel`
dependency is transitive (via `bolt` and `fortify`). Tracing is wired on the
**consumer side** by wrapping the SDK's HTTP transport.

This page documents the recommended pattern, plus span attribute conventions
so traces are searchable and aligned with LLM/agent observability tooling.

## Wiring otelhttp around the Jira client

Use `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` and pass
the wrapped `http.Client` via `jira.WithHTTPClient`.

```go
import (
    "net/http"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    jira "github.com/felixgeelhaar/jirasdk"
)

httpClient := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport,
        otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
            return "jira " + r.Method + " " + r.URL.Path
        }),
    ),
}

client, _ := jira.NewClient(
    jira.WithBaseURL(os.Getenv("JIRA_BASE_URL")),
    jira.WithAPIToken(email, token),
    jira.WithHTTPClient(httpClient),
)
```

## Recommended span attribute schema

When enriching spans (e.g., from your handler before invoking the SDK or via a
custom RoundTripper), use these attribute keys for consistency. They follow
OTel HTTP semconv where applicable and are designed to be useful both for
operational dashboards and for LLM-agent traces that include Jira tool calls.

| Key                | Value                                | Notes                                  |
| ------------------ | ------------------------------------ | -------------------------------------- |
| `jira.endpoint`    | `/rest/api/3/issue/{key}`            | Path template, not the resolved value. |
| `jira.method`      | `GET` / `POST` / `PUT` / `DELETE`    | Mirrors `http.method`.                 |
| `jira.issue.key`   | `PROJ-123`                           | Set when an issue key is in scope.     |
| `jira.project.key` | `PROJ`                               | Set on project-scoped operations.      |
| `jira.jql`         | `project = PROJ AND status = Open`   | For `search.Search`. Truncate to 1 KiB. |
| `jira.bulk.size`   | `42`                                 | Set on bulk endpoints.                 |
| `jira.error.code`  | `issueNotFound` / Atlassian err key  | Map from response error payload.       |

For LLM/agent runs, also set the standard genai attributes on the parent span
(`gen_ai.system`, `gen_ai.tool.name="jira"`, `gen_ai.request.id`) so the Jira
HTTP child spans correlate with the agent step that produced them.

## Collector pipeline (minimal)

```yaml
receivers:
  otlp:
    protocols:
      http:
      grpc:

processors:
  batch:
  attributes/jira_tag:
    actions:
      - key: service.layer
        value: jira-sdk
        action: insert

exporters:
  otlphttp:
    endpoint: ${OTLP_ENDPOINT}

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, attributes/jira_tag]
      exporters: [otlphttp]
```

## See also

- `examples/observability/` — structured logging via `bolt` adapter.
- `resilience/fortify/` — retry / rate-limit / circuit-breaker around the SDK.
