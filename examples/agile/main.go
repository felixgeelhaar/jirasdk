package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/agile"
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

	// Example 1: List all boards
	fmt.Println("=== Listing All Boards ===")
	boards, err := client.Agile.GetBoards(ctx, &agile.BoardsOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Fatalf("Failed to list boards: %v", err)
	}

	fmt.Printf("Found %d boards:\n", len(boards))
	for _, board := range boards {
		fmt.Printf("- %s (ID: %d, Type: %s)\n", board.Name, board.ID, board.Type)
		if board.Location != nil {
			fmt.Printf("  Project: %s\n", board.Location.ProjectName)
		}
	}

	// For the rest of the examples, we'll use the first board
	if len(boards) == 0 {
		log.Fatal("No boards found. Please create a board in Jira first.")
	}
	boardID := boards[0].ID

	// Example 2: Get specific board details
	fmt.Printf("\n=== Getting Board Details (ID: %d) ===\n", boardID)
	boardDetails, err := client.Agile.GetBoard(ctx, boardID)
	if err != nil {
		log.Fatalf("Failed to get board: %v", err)
	}

	fmt.Printf("ID: %d\n", boardDetails.ID)
	fmt.Printf("Name: %s\n", boardDetails.Name)
	fmt.Printf("Type: %s\n", boardDetails.Type)
	if boardDetails.Location != nil {
		fmt.Printf("Project: %s (%s)\n", boardDetails.Location.ProjectName, boardDetails.Location.ProjectKey)
	}

	// Example 3: Sprint Management
	fmt.Printf("\n=== Sprint Management (Board ID: %d) ===\n", boardID)

	// List all sprints
	sprints, err := client.Agile.GetBoardSprints(ctx, boardID, &agile.SprintsOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to list sprints: %v", err)
	} else {
		fmt.Printf("\nFound %d sprints:\n", len(sprints))

		// Group sprints by state
		var active, future, closed []*agile.Sprint
		for _, sprint := range sprints {
			switch sprint.State {
			case "active":
				active = append(active, sprint)
			case "future":
				future = append(future, sprint)
			case "closed":
				closed = append(closed, sprint)
			}
		}

		if len(active) > 0 {
			fmt.Println("\nActive Sprints:")
			for _, s := range active {
				fmt.Printf("  - %s (ID: %d)\n", s.Name, s.ID)
				if s.StartDate != "" {
					fmt.Printf("    Start: %s\n", s.StartDate)
				}
				if s.EndDate != "" {
					fmt.Printf("    End: %s\n", s.EndDate)
				}
				if s.Goal != "" {
					fmt.Printf("    Goal: %s\n", s.Goal)
				}
			}
		}

		if len(future) > 0 {
			fmt.Println("\nFuture Sprints:")
			for _, s := range future {
				fmt.Printf("  - %s (ID: %d)\n", s.Name, s.ID)
			}
		}

		if len(closed) > 0 {
			fmt.Printf("\nClosed Sprints: %d\n", len(closed))
		}
	}

	// Create a new sprint (if scrum board)
	if boardDetails.Type == "scrum" {
		fmt.Println("\nCreating new sprint...")
		newSprint, err := client.Agile.CreateSprint(ctx, &agile.CreateSprintInput{
			Name:          "Sprint Demo",
			OriginBoardID: boardID,
			StartDate:     "2024-06-01T09:00:00.000Z",
			EndDate:       "2024-06-14T17:00:00.000Z",
			Goal:          "Demonstrate agile API features",
		})
		if err != nil {
			log.Printf("Failed to create sprint: %v", err)
		} else {
			fmt.Printf("Created sprint: %s (ID: %d)\n", newSprint.Name, newSprint.ID)

			// Get sprint details
			sprintDetails, err := client.Agile.GetSprint(ctx, newSprint.ID)
			if err != nil {
				log.Printf("Failed to get sprint: %v", err)
			} else {
				fmt.Printf("Sprint Details:\n")
				fmt.Printf("  Name: %s\n", sprintDetails.Name)
				fmt.Printf("  State: %s\n", sprintDetails.State)
				fmt.Printf("  Start: %s\n", sprintDetails.StartDate)
				fmt.Printf("  End: %s\n", sprintDetails.EndDate)
				fmt.Printf("  Goal: %s\n", sprintDetails.Goal)
			}

			// Update sprint
			fmt.Println("\nUpdating sprint...")
			updatedSprint, err := client.Agile.UpdateSprint(ctx, newSprint.ID, &agile.UpdateSprintInput{
				Goal: "Updated: Comprehensive demonstration of agile API features",
			})
			if err != nil {
				log.Printf("Failed to update sprint: %v", err)
			} else {
				fmt.Printf("Updated sprint goal: %s\n", updatedSprint.Goal)
			}

			// Clean up: Delete the sprint
			fmt.Println("\nCleaning up: Deleting demo sprint...")
			err = client.Agile.DeleteSprint(ctx, newSprint.ID)
			if err != nil {
				log.Printf("Failed to delete sprint: %v", err)
			} else {
				fmt.Println("Sprint deleted successfully")
			}
		}
	}

	// Example 4: Epic Management
	fmt.Printf("\n=== Epic Management (Board ID: %d) ===\n", boardID)

	// List all epics
	epics, err := client.Agile.GetBoardEpics(ctx, boardID, &agile.EpicsOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to list epics: %v", err)
	} else {
		fmt.Printf("\nFound %d epics:\n", len(epics))

		// Group epics by completion status
		var done, inProgress []*agile.Epic
		for _, epic := range epics {
			if epic.Done {
				done = append(done, epic)
			} else {
				inProgress = append(inProgress, epic)
			}
		}

		if len(inProgress) > 0 {
			fmt.Println("\nIn Progress:")
			for _, e := range inProgress {
				fmt.Printf("  - %s (%s)\n", e.Name, e.Key)
				if e.Summary != "" {
					fmt.Printf("    %s\n", e.Summary)
				}
			}
		}

		if len(done) > 0 {
			fmt.Println("\nCompleted:")
			for _, e := range done {
				fmt.Printf("  - %s (%s)\n", e.Name, e.Key)
			}
		}

		// Get specific epic details
		if len(epics) > 0 {
			epicID := epics[0].ID
			fmt.Printf("\n=== Getting Epic Details (ID: %d) ===\n", epicID)
			epicDetails, err := client.Agile.GetEpic(ctx, epicID)
			if err != nil {
				log.Printf("Failed to get epic: %v", err)
			} else {
				fmt.Printf("Key: %s\n", epicDetails.Key)
				fmt.Printf("Name: %s\n", epicDetails.Name)
				fmt.Printf("Summary: %s\n", epicDetails.Summary)
				fmt.Printf("Done: %t\n", epicDetails.Done)
				if epicDetails.Color != nil {
					fmt.Printf("Color: %s\n", epicDetails.Color.Key)
				}
			}
		}
	}

	// Example 5: Backlog Management
	fmt.Printf("\n=== Backlog Management (Board ID: %d) ===\n", boardID)

	backlogIssues, err := client.Agile.GetBacklog(ctx, boardID, &agile.BoardsOptions{
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Failed to get backlog: %v", err)
	} else {
		fmt.Printf("Found %d backlog issues\n", len(backlogIssues))
	}

	// Example 6: Moving Issues to Sprint
	if len(sprints) > 0 && len(backlogIssues) > 0 {
		activeSprint := sprints[0]
		fmt.Printf("\n=== Moving Issues to Sprint (Sprint ID: %d) ===\n", activeSprint.ID)

		// Note: This is a demonstration. In real usage, you would get actual issue keys
		// from the backlog and move them to the sprint
		fmt.Println("To move issues to a sprint, use:")
		fmt.Println("  client.Agile.MoveIssuesToSprint(ctx, sprintID, &agile.MoveIssuesToSprintInput{")
		fmt.Println("    Issues: []string{\"PROJ-123\", \"PROJ-124\"},")
		fmt.Println("  })")
	}

	// Example 7: Board Statistics
	fmt.Println("\n=== Board Statistics ===")
	for _, board := range boards {
		fmt.Printf("\n%s (ID: %d):\n", board.Name, board.ID)

		// Get sprints for this board
		boardSprints, err := client.Agile.GetBoardSprints(ctx, board.ID, nil)
		if err == nil {
			activeCount := 0
			for _, s := range boardSprints {
				if s.State == "active" {
					activeCount++
				}
			}
			fmt.Printf("  Total Sprints: %d\n", len(boardSprints))
			fmt.Printf("  Active Sprints: %d\n", activeCount)
		}

		// Get epics for this board
		boardEpics, err := client.Agile.GetBoardEpics(ctx, board.ID, nil)
		if err == nil {
			doneCount := 0
			for _, e := range boardEpics {
				if e.Done {
					doneCount++
				}
			}
			fmt.Printf("  Total Epics: %d\n", len(boardEpics))
			fmt.Printf("  Completed Epics: %d\n", doneCount)
			if len(boardEpics) > 0 {
				completion := float64(doneCount) / float64(len(boardEpics)) * 100
				fmt.Printf("  Completion Rate: %.1f%%\n", completion)
			}
		}
	}

	// Example 8: Filter boards by type
	fmt.Println("\n=== Filtering Boards by Type ===")
	scrumBoards, err := client.Agile.GetBoards(ctx, &agile.BoardsOptions{
		Type:       "scrum",
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to list scrum boards: %v", err)
	} else {
		fmt.Printf("Scrum Boards: %d\n", len(scrumBoards))
	}

	kanbanBoards, err := client.Agile.GetBoards(ctx, &agile.BoardsOptions{
		Type:       "kanban",
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to list kanban boards: %v", err)
	} else {
		fmt.Printf("Kanban Boards: %d\n", len(kanbanBoards))
	}

	fmt.Println("\n=== Agile Example Complete ===")
}
