package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var db *sql.DB

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B5"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
)

type GithubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./students.db")
	if err != nil {
		return err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS classes (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE
	);
	CREATE TABLE IF NOT EXISTS students (
		username TEXT,
		class_id INTEGER,
		FOREIGN KEY(class_id) REFERENCES classes(id),
		UNIQUE(username, class_id)
	);`

	_, err = db.Exec(createTable)
	return err
}

func getLastPushTime(username string) (time.Time, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return time.Time{}, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events/public", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return time.Time{}, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var events []GithubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return time.Time{}, err
	}

	var lastPushTime time.Time
	for _, event := range events {
		if event.Type == "PushEvent" {
			if lastPushTime.IsZero() || event.CreatedAt.After(lastPushTime) {
				lastPushTime = event.CreatedAt
			}
		}
	}

	if lastPushTime.IsZero() {
		return time.Time{}, fmt.Errorf("no push events found")
	}

	return lastPushTime, nil
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

var rootCmd = &cobra.Command{
	Use:   "scv",
	Short: "Student Code Viewer - A CLI tool for managing student GitHub repositories",
}

var addClass = &cobra.Command{
	Use:   "add-class [name]",
	Short: "Add a new class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := db.Exec("INSERT INTO classes (name) VALUES (?)", args[0])
		if err != nil {
			return fmt.Errorf("failed to add class: %v", err)
		}
		fmt.Printf("Added class: %s\n", args[0])
		return nil
	},
}

var listClasses = &cobra.Command{
	Use:   "list-classes",
	Short: "List all classes",
	RunE: func(cmd *cobra.Command, args []string) error {
		rows, err := db.Query("SELECT name FROM classes")
		if err != nil {
			return fmt.Errorf("failed to query classes: %v", err)
		}
		defer rows.Close()

		fmt.Println("Classes:")
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			fmt.Printf("- %s\n", name)
		}
		return nil
	},
}

var addStudent = &cobra.Command{
	Use:   "add-student [class] [username...]",
	Short: "Add one or more students to a class",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]
		usernames := args[1:]

		var classID int
		err := db.QueryRow("SELECT id FROM classes WHERE name = ?", className).Scan(&classID)
		if err != nil {
			return fmt.Errorf("class not found: %s", className)
		}

		for _, username := range usernames {
			_, err := db.Exec("INSERT OR IGNORE INTO students (username, class_id) VALUES (?, ?)",
				username, classID)
			if err != nil {
				return fmt.Errorf("failed to add student %s: %v", username, err)
			}
			fmt.Printf("Added student: %s to class: %s\n", username, className)
		}
		return nil
	},
}

var listStudents = &cobra.Command{
	Use:   "list-students [class]",
	Short: "List all students in a class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]
		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?
			ORDER BY s.username`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		fmt.Printf("Students in %s:\n", className)
		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}
			fmt.Printf("- %s\n", username)
		}
		return nil
	},
}

var checkLastPush = &cobra.Command{
	Use:   "check-activity [class]",
	Short: "Check when students last pushed code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?
			ORDER BY s.username`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		fmt.Printf("\nActivity Report for %s:\n", className)
		fmt.Println("----------------------------------------")

		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}

			lastPush, err := getLastPushTime(username)
			if err != nil {
				fmt.Printf("❌ %s: Error checking activity - %v\n", username, err)
				continue
			}

			timeSince := time.Since(lastPush)

			switch {
			case timeSince < 24*time.Hour:
				fmt.Printf("✅ %s: Last push %s ago\n", username, formatDuration(timeSince))
			case timeSince < 72*time.Hour:
				fmt.Printf("🟡 %s: Last push %s ago\n", username, formatDuration(timeSince))
			default:
				fmt.Printf("⚠️ %s: Last push %s ago\n", username, formatDuration(timeSince))
			}
		}

		fmt.Println("\nLegend:")
		fmt.Println("✅ - Pushed within last 24 hours")
		fmt.Println("🟡 - Pushed within last 72 hours")
		fmt.Println("⚠️ - No push in over 72 hours")
		fmt.Println("❌ - Error checking activity")

		return nil
	},
}

var cloneRepos = &cobra.Command{
	Use:   "clone [class]",
	Short: "Clone repositories for all students in a class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}

			cmd := exec.Command("git", "clone", fmt.Sprintf("https://github.com/%s/%s.github.io", username, username), username)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to clone repository for %s: %v\n", username, err)
				continue
			}
			fmt.Printf("Cloned repository for: %s\n", username)
		}
		return nil
	},
}

var pullRepos = &cobra.Command{
	Use:   "pull [class]",
	Short: "Pull latest changes for all repositories in a class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}

			if _, err := os.Stat(username); err == nil {
				cmd := exec.Command("git", "-C", username, "pull")
				if err := cmd.Run(); err != nil {
					fmt.Printf("Failed to pull repository for %s: %v\n", username, err)
					continue
				}
				fmt.Printf("Pulled latest changes for: %s\n", username)
			}
		}
		return nil
	},
}

var removeStudent = &cobra.Command{
	Use:   "remove-student [class] [username...]",
	Short: "Remove one or more students from a class",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]
		usernames := args[1:]

		var classID int
		err := db.QueryRow("SELECT id FROM classes WHERE name = ?", className).Scan(&classID)
		if err != nil {
			return fmt.Errorf("class not found: %s", className)
		}

		for _, username := range usernames {
			result, err := db.Exec("DELETE FROM students WHERE username = ? AND class_id = ?",
				username, classID)
			if err != nil {
				return fmt.Errorf("failed to remove student %s: %v", username, err)
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("Removed student: %s from class: %s\n", username, className)
			} else {
				fmt.Printf("Student not found: %s in class: %s\n", username, className)
			}
		}
		return nil
	},
}

var removeClass = &cobra.Command{
	Use:   "remove-class [name]",
	Short: "Remove a class and all its students",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Get class ID
		var classID int
		err = tx.QueryRow("SELECT id FROM classes WHERE name = ?", className).Scan(&classID)
		if err != nil {
			return fmt.Errorf("class not found: %s", className)
		}

		// Remove students first
		_, err = tx.Exec("DELETE FROM students WHERE class_id = ?", classID)
		if err != nil {
			return fmt.Errorf("failed to remove students: %v", err)
		}

		// Remove class
		_, err = tx.Exec("DELETE FROM classes WHERE id = ?", classID)
		if err != nil {
			return fmt.Errorf("failed to remove class: %v", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit changes: %v", err)
		}

		fmt.Printf("Removed class: %s and all its students\n", className)
		return nil
	},
}

// Get the date range for the grid (Monday-Friday)
func getGridDateRange() (time.Time, time.Time) {
	now := time.Now()

	// If it's weekend, show last week's Monday-Friday
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		// Go back to Friday
		for now.Weekday() != time.Friday {
			now = now.AddDate(0, 0, -1)
		}
	}

	// Find Monday
	start := now
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}

	// End is either today or Friday, whichever comes first
	end := now
	if end.Weekday() > time.Friday {
		for end.Weekday() != time.Friday {
			end = end.AddDate(0, 0, -1)
		}
	}

	return start, end
}

// Get all push events for a user within a date range
func getUserPushDates(username string, start, end time.Time) (map[string]bool, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events/public", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var events []GithubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	// Create map of dates with pushes
	pushDates := make(map[string]bool)
	for _, event := range events {
		if event.Type == "PushEvent" {
			date := event.CreatedAt.Format("2006-01-02")
			if event.CreatedAt.After(start) && event.CreatedAt.Before(end.AddDate(0, 0, 1)) {
				pushDates[date] = true
			}
		}
	}

	return pushDates, nil
}

var weekHistory = &cobra.Command{
	Use:   "week-history [class]",
	Short: "Show weekly activity history for the class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		// Get date range
		start, end := getGridDateRange()

		// Get all students
		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?
			ORDER BY s.username`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		// Build header
		dates := []string{}
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d.Format("Mon 01/02"))
		}

		// Print header
		fmt.Printf("\nActivity Grid for %s:\n", className)
		fmt.Printf("%-20s", "Username")
		for _, date := range dates {
			fmt.Printf("| %-10s", date)
		}
		fmt.Println()

		// Print separator
		fmt.Printf("%-20s", strings.Repeat("-", 20))
		for range dates {
			fmt.Printf("+-%-10s", strings.Repeat("-", 10))
		}
		fmt.Println()

		// For each student
		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}

			pushDates, err := getUserPushDates(username, start, end)
			if err != nil {
				fmt.Printf("%-20s| ❌ Error checking activity: %v\n", username, err)
				continue
			}

			// Print student row
			fmt.Printf("%-20s", username)
			for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
				date := d.Format("2006-01-02")
				if pushDates[date] {
					fmt.Printf("| %-10s", "✅")
				} else {
					fmt.Printf("| %-10s", "❌")
				}
			}
			fmt.Println()
		}

		fmt.Println("\nLegend:")
		fmt.Println("✅ - Pushed code on this day")
		fmt.Println("❌ - No push activity")

		return nil
	},
}
var cleanRepos = &cobra.Command{
	Use:   "clean [class]",
	Short: "Clean all repositories in a class",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		className := args[0]

		rows, err := db.Query(`
			SELECT s.username 
			FROM students s
			JOIN classes c ON s.class_id = c.id
			WHERE c.name = ?`,
			className)
		if err != nil {
			return fmt.Errorf("failed to query students: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return err
			}

			if _, err := os.Stat(username); err == nil {
				cmd := exec.Command("git", "-C", username, "checkout", ".")
				if err := cmd.Run(); err != nil {
					fmt.Printf("Failed to clean repository for %s: %v\n", username, err)
					continue
				}
				fmt.Printf("Cleaned repository for: %s\n", username)
			}
		}
		return nil
	},
}

func main() {
	if err := initDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	rootCmd.AddCommand(addClass)
	rootCmd.AddCommand(removeClass)
	rootCmd.AddCommand(listClasses)
	rootCmd.AddCommand(addStudent)
	rootCmd.AddCommand(removeStudent)
	rootCmd.AddCommand(listStudents)
	rootCmd.AddCommand(cloneRepos)
	rootCmd.AddCommand(pullRepos)
	rootCmd.AddCommand(cleanRepos)
	rootCmd.AddCommand(checkLastPush)
	rootCmd.AddCommand(weekHistory)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
