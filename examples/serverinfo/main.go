package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	jira "github.com/felixgeelhaar/jirasdk"
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

	// Example 1: Get Server Information
	fmt.Println("=== Jira Server Information ===")
	info, err := client.ServerInfo.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get server info: %v", err)
	}

	fmt.Printf("Server Title: %s\n", info.ServerTitle)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Build Number: %d\n", info.BuildNumber)
	fmt.Printf("Build Date: %s\n", info.BuildDate)
	fmt.Printf("Server Time: %s\n", info.ServerTime)
	fmt.Printf("SCM Info: %s\n", info.ScmInfo)
	fmt.Printf("Base URL: %s\n", info.BaseURL)
	fmt.Printf("Deployment Type: %s\n", info.DeploymentType)

	if len(info.VersionNumbers) > 0 {
		fmt.Printf("\nVersion Components: ")
		for i, v := range info.VersionNumbers {
			if i > 0 {
				fmt.Printf(".")
			}
			fmt.Printf("%d", v)
		}
		fmt.Println()
	}

	// Health checks (if available)
	if len(info.HealthChecks) > 0 {
		fmt.Println("\nHealth Checks:")
		for _, check := range info.HealthChecks {
			status := "❌ Failed"
			if check.Passed {
				status = "✅ Passed"
			}
			fmt.Printf("  %s - %s\n", status, check.Name)
			if check.Description != "" {
				fmt.Printf("    %s\n", check.Description)
			}
		}
	}

	// Example 2: Determine Jira Platform
	fmt.Println("\n=== Platform Analysis ===")
	if strings.Contains(strings.ToLower(info.DeploymentType), "cloud") {
		fmt.Println("Platform: Jira Cloud")
		fmt.Println("Features:")
		fmt.Println("  - Automatic updates")
		fmt.Println("  - Managed infrastructure")
		fmt.Println("  - API Token authentication recommended")
	} else {
		fmt.Println("Platform: Jira Server/Data Center")
		fmt.Println("Features:")
		fmt.Println("  - Self-hosted")
		fmt.Println("  - Custom plugins supported")
		fmt.Println("  - PAT or Basic Auth recommended")
	}

	// Determine if this is a current version
	fmt.Println("\nVersion Status:")
	if len(info.VersionNumbers) >= 1 {
		majorVersion := info.VersionNumbers[0]
		if majorVersion >= 9 {
			fmt.Println("  ✅ Recent version (v9+)")
		} else if majorVersion == 8 {
			fmt.Println("  ⚠️  Older version (v8)")
		} else {
			fmt.Println("  ⚠️  Very old version")
		}
	}

	// Example 3: Get Jira Configuration
	fmt.Println("\n=== Jira Configuration ===")
	config, err := client.ServerInfo.GetConfiguration(ctx)
	if err != nil {
		log.Printf("Failed to get configuration: %v", err)
	} else {
		fmt.Println("\nFeatures:")
		printFeature("Voting", config.VotingEnabled)
		printFeature("Watching", config.WatchingEnabled)
		printFeature("Unassigned Issues", config.UnassignedIssuesAllowed)
		printFeature("Sub-tasks", config.SubTasksEnabled)
		printFeature("Issue Linking", config.IssueLinkingEnabled)
		printFeature("Time Tracking", config.TimeTrackingEnabled)
		printFeature("Attachments", config.AttachmentsEnabled)

		// Time Tracking Configuration
		if config.TimeTrackingEnabled && config.TimeTrackingConfiguration != nil {
			fmt.Println("\nTime Tracking Settings:")
			fmt.Printf("  Working Hours per Day: %.1f\n", config.TimeTrackingConfiguration.WorkingHoursPerDay)
			fmt.Printf("  Working Days per Week: %.1f\n", config.TimeTrackingConfiguration.WorkingDaysPerWeek)
			fmt.Printf("  Time Format: %s\n", config.TimeTrackingConfiguration.TimeFormat)
			fmt.Printf("  Default Unit: %s\n", config.TimeTrackingConfiguration.DefaultUnit)

			// Calculate weekly working hours
			weeklyHours := config.TimeTrackingConfiguration.WorkingHoursPerDay *
				config.TimeTrackingConfiguration.WorkingDaysPerWeek
			fmt.Printf("  Weekly Working Hours: %.1f\n", weeklyHours)
		}
	}

	// Example 4: System Capabilities Summary
	fmt.Println("\n=== System Capabilities ===")
	fmt.Printf("Jira %s (Build %d)\n", info.Version, info.BuildNumber)
	fmt.Printf("Deployment: %s\n", info.DeploymentType)

	if config != nil {
		capabilities := []string{}
		if config.VotingEnabled {
			capabilities = append(capabilities, "Voting")
		}
		if config.WatchingEnabled {
			capabilities = append(capabilities, "Watching")
		}
		if config.TimeTrackingEnabled {
			capabilities = append(capabilities, "Time Tracking")
		}
		if config.IssueLinkingEnabled {
			capabilities = append(capabilities, "Issue Linking")
		}
		if config.SubTasksEnabled {
			capabilities = append(capabilities, "Sub-tasks")
		}
		if config.AttachmentsEnabled {
			capabilities = append(capabilities, "Attachments")
		}

		fmt.Printf("\nEnabled Capabilities (%d):\n", len(capabilities))
		for _, cap := range capabilities {
			fmt.Printf("  ✅ %s\n", cap)
		}

		// Disabled capabilities
		disabled := []string{}
		if !config.VotingEnabled {
			disabled = append(disabled, "Voting")
		}
		if !config.WatchingEnabled {
			disabled = append(disabled, "Watching")
		}
		if !config.TimeTrackingEnabled {
			disabled = append(disabled, "Time Tracking")
		}

		if len(disabled) > 0 {
			fmt.Printf("\nDisabled Capabilities (%d):\n", len(disabled))
			for _, cap := range disabled {
				fmt.Printf("  ❌ %s\n", cap)
			}
		}
	}

	// Example 5: Instance Metadata
	fmt.Println("\n=== Instance Metadata ===")
	fmt.Printf("Base URL: %s\n", info.BaseURL)
	fmt.Printf("Server Title: %s\n", info.ServerTitle)
	fmt.Printf("Build Date: %s\n", info.BuildDate)
	fmt.Printf("SCM Revision: %s\n", info.ScmInfo)

	// Example 6: Compatibility Check
	fmt.Println("\n=== API Compatibility ===")
	fmt.Println("This SDK is designed for Jira REST API v3")

	if len(info.VersionNumbers) >= 1 {
		majorVersion := info.VersionNumbers[0]

		switch {
		case majorVersion >= 8:
			fmt.Println("✅ Full API compatibility expected")
			fmt.Println("   Your Jira version supports all REST API v3 features")

		case majorVersion == 7:
			fmt.Println("⚠️  Partial API compatibility")
			fmt.Println("   Some newer API features may not be available")
			fmt.Println("   Consider upgrading for full feature support")

		default:
			fmt.Println("❌ Limited API compatibility")
			fmt.Println("   Upgrade strongly recommended for REST API v3 support")
		}
	}

	// Example 7: Configuration Recommendations
	fmt.Println("\n=== Configuration Recommendations ===")

	if config != nil {
		recommendations := []string{}

		if !config.TimeTrackingEnabled {
			recommendations = append(recommendations,
				"Enable Time Tracking for better project management and reporting")
		}

		if !config.IssueLinkingEnabled {
			recommendations = append(recommendations,
				"Enable Issue Linking to track relationships between issues")
		}

		if !config.SubTasksEnabled {
			recommendations = append(recommendations,
				"Enable Sub-tasks for better issue decomposition")
		}

		if config.UnassignedIssuesAllowed {
			recommendations = append(recommendations,
				"Consider requiring issue assignment for better accountability")
		}

		if len(recommendations) > 0 {
			fmt.Println("Consider the following to optimize your Jira instance:")
			for i, rec := range recommendations {
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		} else {
			fmt.Println("✅ Your Jira instance is well-configured!")
		}
	}

	fmt.Println("\n=== Server Info Example Complete ===")
}

// printFeature prints a feature status with icon
func printFeature(name string, enabled bool) {
	status := "❌ Disabled"
	if enabled {
		status = "✅ Enabled"
	}
	fmt.Printf("  %s: %s\n", name, status)
}
