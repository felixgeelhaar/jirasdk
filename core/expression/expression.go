// Package expression provides Jira Expression evaluation for Jira.
//
// Jira expressions enable custom automation and dynamic content.
package expression

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Jira Expressions.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Expression service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// EvaluationInput represents input for evaluating an expression.
type EvaluationInput struct {
	Expression string                 `json:"expression"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// EvaluationResult represents the result of expression evaluation.
type EvaluationResult struct {
	Value  interface{}        `json:"value"`
	Meta   *EvaluationMeta    `json:"meta,omitempty"`
	Errors []*EvaluationError `json:"errors,omitempty"`
}

// EvaluationMeta contains metadata about the evaluation.
type EvaluationMeta struct {
	Complexity *Complexity `json:"complexity,omitempty"`
	Issues     []string    `json:"issues,omitempty"`
}

// Complexity represents expression complexity.
type Complexity struct {
	Steps               int `json:"steps"`
	ExpensiveOperations int `json:"expensiveOperations"`
	Beans               int `json:"beans"`
	PrimitiveValues     int `json:"primitiveValues"`
}

// EvaluationError represents an error during evaluation.
type EvaluationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// Evaluate evaluates a Jira expression using the legacy endpoint.
//
// Deprecated: Use EvaluateExpression instead. The /rest/api/3/expression/eval endpoint
// will be removed by Atlassian on August 1, 2025. The new endpoint uses the enhanced
// search API for better performance and scalability (eventual consistency instead of
// strong consistency).
//
// Example:
//
//	result, err := client.Expression.Evaluate(ctx, &expression.EvaluationInput{
//		Expression: "issue.summary",
//		Context: map[string]interface{}{
//			"issue": map[string]interface{}{
//				"key": "PROJ-123",
//			},
//		},
//	})
func (s *Service) Evaluate(ctx context.Context, input *EvaluationInput) (*EvaluationResult, error) {
	if input == nil || input.Expression == "" {
		return nil, fmt.Errorf("expression is required")
	}

	path := "/rest/api/3/expression/eval"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result EvaluationResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// EvaluateExpression evaluates a Jira expression using the new Enhanced Search API.
// This method provides better performance and scalability than the legacy Evaluate method.
//
// Key differences from Evaluate():
//   - Uses enhanced search API with eventual consistency (vs strong consistency)
//   - Better performance and scalability
//   - Same input/output structures
//
// Example:
//
//	result, err := client.Expression.EvaluateExpression(ctx, &expression.EvaluationInput{
//		Expression: "issue.summary",
//		Context: map[string]interface{}{
//			"issue": map[string]interface{}{
//				"key": "PROJ-123",
//			},
//		},
//	})
func (s *Service) EvaluateExpression(ctx context.Context, input *EvaluationInput) (*EvaluationResult, error) {
	if input == nil || input.Expression == "" {
		return nil, fmt.Errorf("expression is required")
	}

	path := "/rest/api/3/expression/evaluate"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result EvaluationResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnalysisInput represents input for analyzing an expression.
type AnalysisInput struct {
	Expressions []string               `json:"expressions"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// AnalysisResult represents the result of expression analysis.
type AnalysisResult struct {
	Results []*ExpressionAnalysis `json:"results"`
}

// ExpressionAnalysis represents analysis of a single expression.
type ExpressionAnalysis struct {
	Expression string             `json:"expression"`
	Valid      bool               `json:"valid"`
	Errors     []*EvaluationError `json:"errors,omitempty"`
	Type       string             `json:"type,omitempty"`
	Complexity *Complexity        `json:"complexity,omitempty"`
}

// Analyze analyzes Jira expressions for syntax and complexity.
//
// Example:
//
//	result, err := client.Expression.Analyze(ctx, &expression.AnalysisInput{
//		Expressions: []string{
//			"issue.summary",
//			"user.displayName",
//		},
//	})
func (s *Service) Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisResult, error) {
	if input == nil || len(input.Expressions) == 0 {
		return nil, fmt.Errorf("at least one expression is required")
	}

	path := "/rest/api/3/expression/analyse"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result AnalysisResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
