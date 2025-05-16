package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// cloneRepo clones a GitHub repository for a student
func cloneRepo(username, repoName, className string) (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the directory structure for the class
	classDir := filepath.Join(homeDir, ".scv", "repos", className)
	if err := os.MkdirAll(classDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create class directory: %w", err)
	}

	// Clone the repository
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", username, repoName)
	studentDir := filepath.Join(classDir, username)

	// Check if the directory already exists
	if _, err := os.Stat(studentDir); !os.IsNotExist(err) {
		return "", fmt.Errorf("repository already exists at %s", studentDir)
	}

	cmd := exec.Command("git", "clone", "-q", repoURL, studentDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %s: %w", output, err)
	}

	return fmt.Sprintf("%s repo has been cloned successfully in %s", username, studentDir), nil
}

// pullRepo pulls the latest changes for a student's repository
func pullRepo(username, className string) (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Get the path to the student's repository
	studentDir := filepath.Join(homeDir, ".scv", "repos", className, username)

	// Check if the directory exists
	if _, err := os.Stat(studentDir); os.IsNotExist(err) {
		return "", fmt.Errorf("repository does not exist at %s", studentDir)
	}

	// Pull the latest changes
	cmd := exec.Command("git", "-C", studentDir, "pull", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to pull repository: %s: %w", output, err)
	}

	return fmt.Sprintf("Successfully pulled updates for %s repo in %s", username, studentDir), nil
}

// cleanRepo removes a student's repository
func cleanRepo(username, className string) (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Get the path to the student's repository
	studentDir := filepath.Join(homeDir, ".scv", "repos", className, username)

	// Check if the directory exists
	if _, err := os.Stat(studentDir); os.IsNotExist(err) {
		return "", fmt.Errorf("repository does not exist at %s", studentDir)
	}

	// Remove the repository
	if err := os.RemoveAll(studentDir); err != nil {
		return "", fmt.Errorf("failed to remove repository: %w", err)
	}

	return fmt.Sprintf("Successfully removed repository at %s", studentDir), nil
}

// getEvents fetches GitHub events for a student
func getEvents(username string) ([]GithubEvent, error) {
	// Fetch events from GitHub API
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var events []GithubEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return events, nil
}

// checkActivity checks if a student has been active on GitHub in the last week
func checkActivity(username string) (bool, error) {
	events, err := getEvents(username)
	if err != nil {
		return false, err
	}

	// Check if there are any events in the last week
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	for _, event := range events {
		if event.CreatedAt.After(oneWeekAgo) {
			return true, nil
		}
	}

	return false, nil
}

// formatActivityReport formats a student's GitHub activity report
func formatActivityReport(username string, events []GithubEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GitHub Activity for %s:\n\n", username))

	if len(events) == 0 {
		sb.WriteString("No recent activity found.")
		return sb.String()
	}

	// Group events by day
	eventsByDay := make(map[string][]GithubEvent)
	for _, event := range events {
		day := event.CreatedAt.Format("2006-01-02")
		eventsByDay[day] = append(eventsByDay[day], event)
	}

	// Format events by day
	for day, dayEvents := range eventsByDay {
		sb.WriteString(fmt.Sprintf("Date: %s\n", day))
		for _, event := range dayEvents {
			sb.WriteString(fmt.Sprintf("  • %s on %s\n", event.Type, event.Repo.Name))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
