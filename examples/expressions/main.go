package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/expression"
)

func main() {
	// Get credentials from environment
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")

	if email == "" || apiToken == "" || baseURL == "" {
		log.Fatal("JIRA_EMAIL, JIRA_API_TOKEN, and JIRA_BASE_URL must be set")
	}

	// Create authenticated client
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple Expression Evaluation
	fmt.Println("=== Simple Expression Evaluation ===")

	// Evaluate a simple expression
	result, err := client.Expression.Evaluate(ctx, &expression.EvaluationInput{
		Expression: "1 + 1",
	})
	if err != nil {
		log.Fatalf("Failed to evaluate expression: %v", err)
	}

	fmt.Printf("Expression: 1 + 1\n")
	fmt.Printf("Result: %v\n", result.Value)

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, evalErr := range result.Errors {
			fmt.Printf("  - %s (line %d, col %d)\n", evalErr.Message, evalErr.Line, evalErr.Column)
		}
	}

	// Example 2: Expression with Context
	fmt.Println("\n=== Expression with Context ===")

	// Create context for the expression
	exprContext := map[string]interface{}{
		"issue": map[string]interface{}{
			"key":     "DEMO-123",
			"summary": "Implement new feature",
			"priority": map[string]interface{}{
				"name": "High",
			},
		},
		"user": map[string]interface{}{
			"displayName": "John Doe",
			"emailAddress": "john@example.com",
		},
	}

	// Evaluate expression with context
	result, err = client.Expression.Evaluate(ctx, &expression.EvaluationInput{
		Expression: "issue.key + ': ' + issue.summary",
		Context:    exprContext,
	})
	if err != nil {
		log.Printf("Failed to evaluate expression: %v", err)
	} else {
		fmt.Printf("Expression: issue.key + ': ' + issue.summary\n")
		fmt.Printf("Result: %v\n", result.Value)
	}

	// Example 3: Conditional Expression
	fmt.Println("\n=== Conditional Expression ===")

	result, err = client.Expression.Evaluate(ctx, &expression.EvaluationInput{
		Expression: "issue.priority.name == 'High' ? 'URGENT' : 'Normal'",
		Context:    exprContext,
	})
	if err != nil {
		log.Printf("Failed to evaluate expression: %v", err)
	} else {
		fmt.Printf("Expression: issue.priority.name == 'High' ? 'URGENT' : 'Normal'\n")
		fmt.Printf("Result: %v\n", result.Value)
	}

	// Example 4: Expression Complexity Analysis
	fmt.Println("\n=== Expression Complexity Analysis ===")

	if result.Meta != nil && result.Meta.Complexity != nil {
		fmt.Println("Complexity Metrics:")
		fmt.Printf("  Steps: %d\n", result.Meta.Complexity.Steps)
		fmt.Printf("  Expensive Operations: %d\n", result.Meta.Complexity.ExpensiveOperations)
		fmt.Printf("  Beans: %d\n", result.Meta.Complexity.Beans)
		fmt.Printf("  Primitive Values: %d\n", result.Meta.Complexity.PrimitiveValues)
	}

	// Example 5: Analyze Multiple Expressions
	fmt.Println("\n=== Analyzing Multiple Expressions ===")

	expressions := []string{
		"issue.summary",
		"user.displayName",
		"issue.key + ' - ' + issue.summary",
		"issue.priority.name == 'High'",
		"issue.key.substring(0, 4)",
		"invalid syntax here !@#",
	}

	analysis, err := client.Expression.Analyze(ctx, &expression.AnalysisInput{
		Expressions: expressions,
		Context:     exprContext,
	})
	if err != nil {
		log.Fatalf("Failed to analyze expressions: %v", err)
	}

	fmt.Printf("Analyzed %d expressions:\n\n", len(analysis.Results))

	for i, result := range analysis.Results {
		fmt.Printf("%d. Expression: %s\n", i+1, result.Expression)
		fmt.Printf("   Valid: %t\n", result.Valid)

		if result.Type != "" {
			fmt.Printf("   Return Type: %s\n", result.Type)
		}

		if result.Complexity != nil {
			fmt.Printf("   Complexity: %d steps, %d expensive ops\n",
				result.Complexity.Steps, result.Complexity.ExpensiveOperations)
		}

		if len(result.Errors) > 0 {
			fmt.Println("   Errors:")
			for _, evalErr := range result.Errors {
				fmt.Printf("     - %s", evalErr.Message)
				if evalErr.Line > 0 {
					fmt.Printf(" (line %d, col %d)", evalErr.Line, evalErr.Column)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}

	// Example 6: Common Expression Patterns
	fmt.Println("=== Common Expression Patterns ===")

	patterns := []struct {
		name       string
		expression string
	}{
		{"Issue Key", "issue.key"},
		{"Issue Summary", "issue.summary"},
		{"Assignee Name", "issue.assignee?.displayName"},
		{"Issue Created Date", "issue.created"},
		{"Priority Check", "issue.priority.name in ['High', 'Critical']"},
		{"Status Category", "issue.status.statusCategory.key"},
		{"Has Labels", "issue.labels.length > 0"},
		{"Reporter Email", "issue.reporter.emailAddress"},
		{"Days Since Created", "(now() - issue.created).days"},
		{"Is Overdue", "issue.dueDate < now()"},
	}

	for _, pattern := range patterns {
		result, err := client.Expression.Analyze(ctx, &expression.AnalysisInput{
			Expressions: []string{pattern.expression},
		})

		if err != nil {
			log.Printf("Failed to analyze %s: %v", pattern.name, err)
			continue
		}

		if len(result.Results) > 0 {
			r := result.Results[0]
			status := "✅"
			if !r.Valid {
				status = "❌"
			}

			fmt.Printf("%s %s\n", status, pattern.name)
			fmt.Printf("   Expression: %s\n", pattern.expression)

			if r.Valid && r.Type != "" {
				fmt.Printf("   Returns: %s\n", r.Type)
			}

			if !r.Valid && len(r.Errors) > 0 {
				fmt.Printf("   Error: %s\n", r.Errors[0].Message)
			}
			fmt.Println()
		}
	}

	// Example 7: Error Handling in Expressions
	fmt.Println("=== Error Handling Examples ===")

	errorExpressions := []string{
		"issue.nonExistentField",
		"1 / 0",
		"issue.key +",
		"(unclosed parenthesis",
	}

	for _, expr := range errorExpressions {
		result, err := client.Expression.Evaluate(ctx, &expression.EvaluationInput{
			Expression: expr,
		})

		fmt.Printf("Expression: %s\n", expr)

		if err != nil {
			fmt.Printf("  API Error: %v\n", err)
		} else if len(result.Errors) > 0 {
			fmt.Println("  Evaluation Errors:")
			for _, evalErr := range result.Errors {
				fmt.Printf("    - Type: %s\n", evalErr.Type)
				fmt.Printf("      Message: %s\n", evalErr.Message)
				if evalErr.Line > 0 {
					fmt.Printf("      Location: line %d, column %d\n", evalErr.Line, evalErr.Column)
				}
			}
		} else {
			fmt.Printf("  Result: %v\n", result.Value)
		}
		fmt.Println()
	}

	// Example 8: Expression Performance Analysis
	fmt.Println("=== Expression Performance Analysis ===")

	performanceTests := []string{
		"1 + 1",                                // Simple
		"issue.key + ' - ' + issue.summary",    // String concatenation
		"issue.labels.map(l => l.toUpperCase())", // Array operation
		"[1, 2, 3, 4, 5].reduce((a, b) => a + b, 0)", // Complex operation
	}

	for _, expr := range performanceTests {
		result, err := client.Expression.Analyze(ctx, &expression.AnalysisInput{
			Expressions: []string{expr},
		})

		if err != nil {
			log.Printf("Failed to analyze: %v", err)
			continue
		}

		if len(result.Results) > 0 && result.Results[0].Complexity != nil {
			complexity := result.Results[0].Complexity
			fmt.Printf("Expression: %s\n", expr)
			fmt.Printf("  Steps: %d\n", complexity.Steps)
			fmt.Printf("  Expensive Operations: %d\n", complexity.ExpensiveOperations)

			// Determine performance category
			if complexity.Steps <= 5 && complexity.ExpensiveOperations == 0 {
				fmt.Println("  Performance: ✅ Fast")
			} else if complexity.Steps <= 20 || complexity.ExpensiveOperations <= 2 {
				fmt.Println("  Performance: ⚠️  Moderate")
			} else {
				fmt.Println("  Performance: ❌ Slow")
			}
			fmt.Println()
		}
	}

	// Example 9: Best Practices Summary
	fmt.Println("=== Best Practices for Jira Expressions ===")

	fmt.Println("✅ DO:")
	fmt.Println("  - Use null-safe navigation (?.)")
	fmt.Println("  - Keep expressions simple and readable")
	fmt.Println("  - Validate expressions before deployment")
	fmt.Println("  - Use analysis API to check complexity")
	fmt.Println("  - Handle potential errors gracefully")
	fmt.Println()

	fmt.Println("❌ DON'T:")
	fmt.Println("  - Create deeply nested expressions")
	fmt.Println("  - Use expensive operations in loops")
	fmt.Println("  - Rely on undocumented fields")
	fmt.Println("  - Ignore expression complexity warnings")

	fmt.Println("\n=== Expressions Example Complete ===")
}
