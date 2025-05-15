package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cloneAllRepos clones the '[studentUsername].github.io' repository for each student in the class.
// It creates a directory structure like '[className]/[studentUsername]/[studentUsername].github.io'.
// Returns a slice of strings with logs of operations.
func cloneAllRepos(className string, studentUsernames []string) []string {
	logs := []string{}
	baseDir := filepath.Join(".", className) // Clones into a subdirectory named after the class

	if len(studentUsernames) == 0 {
		logs = append(logs, fmt.Sprintf("No students found for class '%s' to clone repos.", className))
		return logs
	}

	logs = append(logs, fmt.Sprintf("Starting to clone repositories for class: %s", className))

	for _, studentUsername := range studentUsernames {
		repoName := fmt.Sprintf("%s.github.io", studentUsername)
		repoURL := fmt.Sprintf("https://github.com/%s/%s.git", studentUsername, repoName)
		studentDir := filepath.Join(baseDir, studentUsername)
		targetPath := filepath.Join(studentDir, repoName)

		// Check if already cloned
		if _, err := os.Stat(targetPath); err == nil {
			logs = append(logs, fmt.Sprintf("Repository %s already exists for %s. Skipping clone.", repoName, studentUsername))
			continue
		}

		logs = append(logs, fmt.Sprintf("Cloning %s for student %s into %s...", repoURL, studentUsername, targetPath))

		if err := os.MkdirAll(studentDir, 0755); err != nil {
			logs = append(logs, fmt.Sprintf("Error creating directory %s: %v", studentDir, err))
			continue
		}

		cmd := exec.Command("git", "clone", repoURL, targetPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logs = append(logs, fmt.Sprintf("Error cloning %s: %v\nOutput: %s", repoURL, err, string(output)))
		} else {
			logs = append(logs, fmt.Sprintf("Successfully cloned %s.\nOutput: %s", repoURL, string(output)))
		}
	}
	return logs
}

// pullAllRepos pulls updates for the '[studentUsername].github.io' repository for each student.
// Assumes repositories are already cloned in '[className]/[studentUsername]/[studentUsername].github.io'.
// Returns a slice of strings with logs of operations.
func pullAllRepos(className string, studentUsernames []string) []string {
	logs := []string{}
	baseDir := filepath.Join(".", className)

	if len(studentUsernames) == 0 {
		logs = append(logs, fmt.Sprintf("No students found for class '%s' to pull repos.", className))
		return logs
	}

	logs = append(logs, fmt.Sprintf("Starting to pull repositories for class: %s", className))

	for _, studentUsername := range studentUsernames {
		repoName := fmt.Sprintf("%s.github.io", studentUsername)
		repoPath := filepath.Join(baseDir, studentUsername, repoName)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			logs = append(logs, fmt.Sprintf("Repository %s not found for %s. Skipping pull. Please clone first.", repoName, studentUsername))
			continue
		}

		logs = append(logs, fmt.Sprintf("Pulling updates for %s in %s...", repoName, repoPath))

		cmd := exec.Command("git", "-C", repoPath, "pull")
		output, err := cmd.CombinedOutput()
		if err != nil {
			logs = append(logs, fmt.Sprintf("Error pulling %s: %v\nOutput: %s", repoName, err, string(output)))
		} else {
			logs = append(logs, fmt.Sprintf("Successfully pulled %s.\nOutput: %s", repoName, string(output)))
		}
	}
	return logs
}

// cleanAllRepos removes the entire directory for the class, including all cloned student repos.
// Returns a slice of strings with logs of operations.
func cleanAllRepos(className string) []string {
	logs := []string{}
	classDir := filepath.Join(".", className)

	logs = append(logs, fmt.Sprintf("Attempting to clean (remove) directory: %s", classDir))

	if _, err := os.Stat(classDir); os.IsNotExist(err) {
		logs = append(logs, fmt.Sprintf("Directory %s does not exist. Nothing to clean.", classDir))
		return logs
	}

	err := os.RemoveAll(classDir)
	if err != nil {
		logs = append(logs, fmt.Sprintf("Error cleaning directory %s: %v", classDir, err))
	} else {
		logs = append(logs, fmt.Sprintf("Successfully cleaned directory %s.", classDir))
	}
	return logs
}

// Helper function to join log messages
func formatLogs(logs []string) string {
	return strings.Join(logs, "\n")
}
