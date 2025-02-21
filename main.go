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

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

type GithubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
}

var db *sql.DB

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B5")).
			MarginLeft(2)

	outputBoxStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF75B5"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFF00"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	paginationStyle = list.DefaultStyles().
			PaginationStyle.
			PaddingLeft(4)

	helpStyle = list.DefaultStyles().
			HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	docStyle = lipgloss.NewStyle().
			Margin(1, 2)
)

const (
	iconSuccess = "✔" // single-width check
	iconWarning = "!" // single-width exclamation
	iconError   = "✖" // single-width cross
)

// Model states
const (
	stateMainMenu = iota
	stateClassInput
	stateStudentInput
	stateOutput
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
	list         list.Model
	state        int
	classInput   textinput.Model
	studentInput textinput.Model
	className    string
	err          error
	output       string // holds command output to be rendered in stateOutput
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

func initialModel() model {
	// Create main menu items
	items := []list.Item{
		item{title: "Add Class", description: "Create a new class"},
		item{title: "Remove Class", description: "Remove a class and its students"},
		item{title: "List Classes", description: "Show all classes"},
		item{title: "Add Students", description: "Add students to a class"},
		item{title: "Remove Students", description: "Remove students from a class"},
		item{title: "List Students", description: "Show all students in a class"},
		item{title: "Clone Repositories", description: "Clone all student repositories"},
		item{title: "Pull Changes", description: "Update all repositories"},
		item{title: "Clean Changes", description: "Revert local changes"},
		item{title: "Check Activity", description: "View recent student activity"},
		item{title: "Week History", description: "Show weekly activity grid"},
		item{title: "Quit", description: "Exit the application"},
	}

	// Setup list
	l := list.New(items, list.NewDefaultDelegate(), 40, 12)
	l.Title = "Student Code Viewer"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Setup inputs
	classInput := textinput.New()
	classInput.Placeholder = "Enter class name"
	classInput.Focus()

	studentInput := textinput.New()
	studentInput.Placeholder = "Enter student usernames (space-separated)"
	studentInput.Focus()

	return model{
		list:         l,
		state:        stateMainMenu,
		classInput:   classInput,
		studentInput: studentInput,
		output:       "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If we're in the output view, any Enter or Esc returns to the main menu.
	if m.state == stateOutput {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
				m.state = stateMainMenu
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.state == stateMainMenu {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Quit":
						return m, tea.Quit
					case "Add Class", "Remove Class", "List Students", "Clone Repositories",
						"Pull Changes", "Clean Changes", "Check Activity", "Week History":
						m.state = stateClassInput
						return m, nil
					case "Add Students":
						m.state = stateClassInput
						m.className = ""
						return m, nil
					case "List Classes":
						rows, err := db.Query("SELECT name FROM classes")
						if err != nil {
							m.err = err
							return m, nil
						}
						defer rows.Close()

						var sb strings.Builder
						sb.WriteString("Classes:\n")
						for rows.Next() {
							var name string
							rows.Scan(&name)
							sb.WriteString(fmt.Sprintf("- %s\n", name))
						}
						m.output = sb.String()
						m.state = stateOutput
						return m, nil
					}
				}
			} else if m.state == stateClassInput {
				m.className = m.classInput.Value()
				i, _ := m.list.SelectedItem().(item)

				if i.title == "Add Students" {
					m.state = stateStudentInput
					return m, nil
				}

				switch i.title {
				case "Add Class":
					_, err := db.Exec("INSERT INTO classes (name) VALUES (?)", m.className)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.output = fmt.Sprintf("Added class: %s\n", m.className)
					m.state = stateOutput
					return m, nil

				case "Remove Class":
					tx, err := db.Begin()
					if err != nil {
						m.err = err
						return m, nil
					}
					defer tx.Rollback()

					var classID int
					err = tx.QueryRow("SELECT id FROM classes WHERE name = ?", m.className).Scan(&classID)
					if err != nil {
						m.err = fmt.Errorf("class not found: %s", m.className)
						return m, nil
					}

					_, err = tx.Exec("DELETE FROM students WHERE class_id = ?", classID)
					if err != nil {
						m.err = fmt.Errorf("failed to remove students: %v", err)
						return m, nil
					}

					_, err = tx.Exec("DELETE FROM classes WHERE id = ?", classID)
					if err != nil {
						m.err = fmt.Errorf("failed to remove class: %v", err)
						return m, nil
					}

					if err := tx.Commit(); err != nil {
						m.err = fmt.Errorf("failed to commit changes: %v", err)
						return m, nil
					}

					m.output = fmt.Sprintf("Removed class: %s and all its students\n", m.className)
					m.state = stateOutput
					return m, nil

				case "List Students":
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?
						ORDER BY s.username`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("Students in %s:\n", m.className))
					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}
						sb.WriteString(fmt.Sprintf("- %s\n", username))
					}
					m.output = sb.String()
					m.state = stateOutput
					return m, nil

				case "Clone Repositories":
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}

						cmd := exec.Command("git", "clone", fmt.Sprintf("https://github.com/%s/%s.github.io", username, username), username)
						if err := cmd.Run(); err != nil {
							sb.WriteString(fmt.Sprintf("Failed to clone repository for %s: %v\n", username, err))
							continue
						}
						sb.WriteString(fmt.Sprintf("Cloned repository for: %s\n", username))
					}
					m.output = sb.String()
					m.state = stateOutput
					return m, nil

				case "Pull Changes":
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}

						if _, err := os.Stat(username); err == nil {
							cmd := exec.Command("git", "-C", username, "pull")
							if err := cmd.Run(); err != nil {
								sb.WriteString(fmt.Sprintf("Failed to pull repository for %s: %v\n", username, err))
								continue
							}
							sb.WriteString(fmt.Sprintf("Pulled latest changes for: %s\n", username))
						}
					}
					m.output = sb.String()
					m.state = stateOutput
					return m, nil

				case "Clean Changes":
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}

						if _, err := os.Stat(username); err == nil {
							cmd := exec.Command("git", "-C", username, "checkout", ".")
							if err := cmd.Run(); err != nil {
								sb.WriteString(fmt.Sprintf("Failed to clean repository for %s: %v\n", username, err))
								continue
							}
							sb.WriteString(fmt.Sprintf("Cleaned repository for: %s\n", username))
						}
					}
					m.output = sb.String()
					m.state = stateOutput
					return m, nil

				case "Check Activity":
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?
						ORDER BY s.username`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("Activity Report for %s:\n", m.className))
					sb.WriteString("----------------------------------------\n")

					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}

						lastPush, err := getLastPushTime(username)
						if err != nil {
							sb.WriteString(fmt.Sprintf("%s %s: Error checking activity - %v\n",
								errorStyle.Render("❌"),
								errorStyle.Render(username),
								err,
							))
							continue
						}

						timeSince := time.Since(lastPush)
						switch {
						case timeSince < 24*time.Hour:
							sb.WriteString(fmt.Sprintf("%s %s: Last push %s ago\n",
								successStyle.Render(iconSuccess),
								successStyle.Render(username),
								formatDuration(timeSince),
							))
						case timeSince < 72*time.Hour:
							sb.WriteString(fmt.Sprintf("%s %s: Last push %s ago\n",
								warningStyle.Render(iconWarning),
								warningStyle.Render(username),
								formatDuration(timeSince),
							))
						default:
							sb.WriteString(fmt.Sprintf("%s %s: Last push %s ago\n",
								errorStyle.Render(iconError),
								errorStyle.Render(username),
								formatDuration(timeSince),
							))
						}
					}

					m.output = sb.String()
					m.state = stateOutput
					return m, nil

				case "Week History":
					start, end := getGridDateRange()
					rows, err := db.Query(`
						SELECT s.username 
						FROM students s
						JOIN classes c ON s.class_id = c.id
						WHERE c.name = ?
						ORDER BY s.username`,
						m.className)
					if err != nil {
						m.err = fmt.Errorf("failed to query students: %v", err)
						return m, nil
					}
					defer rows.Close()

					var sb strings.Builder
					// Build header and date mapping
					var dates []struct {
						display string
						lookup  string
					}

					for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
						dates = append(dates, struct {
							display string
							lookup  string
						}{
							display: d.Format("Mon 01/02"),
							lookup:  d.Format("2006-01-02"),
						})
					}

					sb.WriteString(fmt.Sprintf("Activity Grid for %s:\n", m.className))
					sb.WriteString(fmt.Sprintf("%-20s", "Username"))
					for _, d := range dates {
						sb.WriteString(fmt.Sprintf("| %-9s", d.display))
					}
					sb.WriteString("\n")

					// Print separator row
					sb.WriteString(strings.Repeat("-", 20))
					for range dates {
						sb.WriteString("+----------")
					}
					sb.WriteString("\n")

					for rows.Next() {
						var username string
						if err := rows.Scan(&username); err != nil {
							m.err = err
							return m, nil
						}

						pushDates, err := getUserPushDates(username, start, end)
						if err != nil {
							sb.WriteString(fmt.Sprintf("%-20s| %s\n", username, errorStyle.Render(fmt.Sprintf("Error: %v", err))))
							continue
						}

						sb.WriteString(fmt.Sprintf("%-20s", username))
						for _, d := range dates {
							if pushDates[d.lookup] {
								sb.WriteString(fmt.Sprintf("| %-9s", successStyle.Render(iconSuccess)))
							} else {
								sb.WriteString(fmt.Sprintf("| %-9s", errorStyle.Render(iconError)))
							}
						}
						sb.WriteString("\n")
					}
					sb.WriteString("\nLegend:\n")
					sb.WriteString(successStyle.Render(fmt.Sprintf("%s - Pushed code on this day\n", iconSuccess)))
					sb.WriteString(errorStyle.Render(fmt.Sprintf("%s - No push activity\n", iconError)))

					m.output = sb.String()
					m.state = stateOutput
					return m, nil
				}
				return m, tea.Quit
			} else if m.state == stateStudentInput {
				usernames := strings.Fields(m.studentInput.Value())

				var classID int
				err := db.QueryRow("SELECT id FROM classes WHERE name = ?", m.className).Scan(&classID)
				if err != nil {
					m.err = fmt.Errorf("class not found: %s", m.className)
					return m, nil
				}

				var sb strings.Builder
				for _, username := range usernames {
					_, err := db.Exec("INSERT OR IGNORE INTO students (username, class_id) VALUES (?, ?)",
						username, classID)
					if err != nil {
						m.err = err
						return m, nil
					}
					sb.WriteString(fmt.Sprintf("Added student: %s to class: %s\n", username, m.className))
				}
				m.output = sb.String()
				m.state = stateOutput
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case stateMainMenu:
		m.list, cmd = m.list.Update(msg)
	case stateClassInput:
		m.classInput, cmd = m.classInput.Update(msg)
	case stateStudentInput:
		m.studentInput, cmd = m.studentInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case stateMainMenu:
		return docStyle.Render(m.list.View())
	case stateClassInput:
		return docStyle.Render(
			titleStyle.Render("Enter Class Name") + "\n\n" +
				m.classInput.View(),
		)
	case stateStudentInput:
		return docStyle.Render(
			titleStyle.Render("Enter Student Usernames") + "\n" +
				"(Space-separated list of GitHub usernames)\n\n" +
				m.studentInput.View(),
		)
	case stateOutput:
		return docStyle.Render(
			outputBoxStyle.Render(m.output + "\n\nPress Enter/Esc to go back."),
		)
	default:
		return "Loading..."
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./students.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
