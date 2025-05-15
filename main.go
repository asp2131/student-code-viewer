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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1)

	outputBoxStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF75B5"))

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			MarginLeft(2).
			MarginBottom(1)

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
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
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
	stateClassSelection
	stateClassManagement
	stateManageStudents
	stateManageRepos
	stateViewActivity
	stateDeleteConfirmation
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
	list          list.Model
	state         int
	classInput    textinput.Model
	studentInput  textinput.Model
	className     string
	err           error
	output        string   // holds command output to be rendered in stateOutput
	menuHistory   []int    // stack to track menu navigation history for back functionality
	classList     []string // list of available classes for selection
	selectedClass string   // currently selected class
	currentMenu   string   // title of the current menu for breadcrumb navigation

	// Simple menu fields
	menuItems    []item // current menu items
	selectedItem int    // index of the currently selected item
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

func centerText(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
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

// createSimpleMenu creates a string-based menu display that shows all options at once
func createSimpleMenu(title string, options []item) string {
	return createSimpleMenuWithSelection(title, options, 0)
}

// createSimpleMenuWithSelection creates a menu with a specific item selected
func createSimpleMenuWithSelection(title string, options []item, selectedIndex int) string {
	var sb strings.Builder

	// Add title
	sb.WriteString(titleStyle.Render(title) + "\n\n")

	// Add all options
	for i, opt := range options {
		prefix := "  "
		if i == selectedIndex {
			// Highlight the selected option
			prefix = "→ "
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B5")).Render(prefix+opt.title) + "\n")
			sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(opt.description) + "\n\n")
		} else {
			sb.WriteString(prefix + opt.title + "\n")
			sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(opt.description) + "\n\n")
		}
	}

	// Add help text
	sb.WriteString("\n" + helpStyle.Render("↑/k up • ↓/j down • enter select • q quit"))

	return sb.String()
}

func initialModel() model {
	// Create top-level menu items (new structure with 3 main options)
	menuItems := []item{
		{title: "Manage Classes", description: "Select and manage an existing class"},
		{title: "Create Class", description: "Create a new class"},
		{title: "Quit", description: "Exit the application"},
	}

	// Setup list (keep for backward compatibility)
	items := []list.Item{
		item{title: "Manage Classes", description: "Select and manage an existing class"},
		item{title: "Create Class", description: "Create a new class"},
		item{title: "Quit", description: "Exit the application"},
	}
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
		list:          l,
		state:         stateMainMenu,
		classInput:    classInput,
		studentInput:  studentInput,
		output:        "",
		menuHistory:   []int{},
		classList:     []string{},
		selectedClass: "",
		currentMenu:   "Main Menu",
		menuItems:     menuItems,
		selectedItem:  0,
	}
}

// showWeekHistoryTview builds and runs a tview application that displays the weekly activity grid.
func showWeekHistoryTview(className string) error {
	start, end := getGridDateRange()
	rows, err := db.Query(`
		SELECT s.username 
		FROM students s
		JOIN classes c ON s.class_id = c.id
		WHERE c.name = ?
		ORDER BY s.username`,
		className)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Create a tview table.
	table := tview.NewTable().SetBorders(true)

	// Header row.
	table.SetCell(0, 0, tview.NewTableCell("Username").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter).
		SetSelectable(false))
	col := 1
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		header := d.Format("Mon 01/02")
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
		col++
	}

	// Fill in rows with student activity.
	rowIndex := 1
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return err
		}

		table.SetCell(rowIndex, 0, tview.NewTableCell(username).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter))

		col = 1
		pushDates, err := getUserPushDates(username, start, end)
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateKey := d.Format("2006-01-02")
			if err == nil && pushDates[dateKey] {
				table.SetCell(rowIndex, col, tview.NewTableCell("✓").
					SetTextColor(tcell.ColorGreen).
					SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(rowIndex, col, tview.NewTableCell("✖").
					SetTextColor(tcell.ColorRed).
					SetAlign(tview.AlignCenter))
			}
			col++
		}
		rowIndex++
	}

	// Create a legend.
	legend := tview.NewTextView().
		SetText("Press ESC or Enter to return to main menu").
		SetTextColor(tcell.ColorYellow).
		SetTextAlign(tview.AlignCenter)

	// Layout: table above the legend.
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(legend, 1, 1, false)

	// Create the tview application.
	app := tview.NewApplication()

	// Allow the user to exit the tview screen with Enter or Escape.
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			app.Stop()
			return nil
		}
		return event
	})

	// Run the tview application.
	err = app.SetRoot(flex, true).Run()
	// Show a loading animation while the terminal restores
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i := 0; i < 10; i++ {
		fmt.Printf("\r%s Returning to main menu...", spinner[i%len(spinner)])
		time.Sleep(200 * time.Millisecond)
	}
	// Add a short delay to help the terminal restore its state before returning.
	time.Sleep(1000 * time.Millisecond)

	fmt.Print("\r") // Clear the line
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

// Helper function to load available classes from the database
func loadClasses() ([]string, error) {
	rows, err := db.Query("SELECT name FROM classes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		classes = append(classes, name)
	}
	return classes, nil
}

// Helper function to create class selection menu
func createClassSelectionMenu(classes []string) list.Model {
	items := make([]list.Item, len(classes))
	for i, class := range classes {
		items[i] = item{title: class, description: "Select to manage this class"}
	}

	l := list.New(items, list.NewDefaultDelegate(), 40, 12)
	l.Title = "Select a Class"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

// Helper function to create class management menu
func createClassManagementMenu(className string) list.Model {
	items := []list.Item{
		item{title: "Manage Students", description: "Add or remove students"},
		item{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
		item{title: "View GH Activity", description: "Check student GitHub activity"},
		item{title: "Delete Class", description: "Delete this class and its data"},
		item{title: "Back", description: "Return to main menu"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 40, 12)
	l.Title = "Managing Class: " + className
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If we're in the output view, any Enter or Esc returns to the previous menu.
	if m.state == stateOutput {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
				// Pop the last state from the history if available
				if len(m.menuHistory) > 0 {
					lastIndex := len(m.menuHistory) - 1
					m.state = m.menuHistory[lastIndex]
					m.menuHistory = m.menuHistory[:lastIndex] // Remove the last item
				} else {
					m.state = stateMainMenu
				}
			}
		}
		return m, nil
	}

	// We'll handle list updates at the end of the function to ensure key presses are processed

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Allow escaping from class input back to main menu
			if m.state == stateClassInput && m.currentMenu == "Create Class" {
				// Return to main menu
				m.state = stateMainMenu
				m.currentMenu = "Main Menu"
				m.selectedItem = 0
				return m, nil
			}

		// Handle menu navigation
		case "up", "k":
			if m.state == stateMainMenu {
				// Move selection up in main menu
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateClassSelection {
				// Move selection up in class selection menu
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateClassManagement {
				// Move selection up in class management menu
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			}

		case "down", "j":
			if m.state == stateMainMenu {
				// Move selection down in main menu
				if m.selectedItem < len(m.menuItems)-1 {
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateClassSelection {
				// Move selection down in class selection menu
				// Account for the additional Back option
				if m.selectedItem < len(m.classList) { // Now includes the Back option
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateClassManagement {
				// Move selection down in class management menu
				if m.selectedItem < 4 { // 5 items in the management menu
					m.selectedItem++
				}
				return m, nil
			}

		case "enter":
			if m.state == stateMainMenu {
				// Use the selectedItem index to determine which option was selected
				if m.selectedItem < len(m.menuItems) {
					selectedOption := m.menuItems[m.selectedItem].title
					switch selectedOption {
					case "Quit":
						return m, tea.Quit
					case "Manage Classes":
						// Load available classes
						classes, err := loadClasses()
						if err != nil {
							m.err = err
							m.output = fmt.Sprintf("Error loading classes: %v", err)
							m.menuHistory = append(m.menuHistory, m.state)
							m.state = stateOutput
							return m, nil
						}

						if len(classes) == 0 {
							m.output = "No classes found. Please create a class first."
							m.menuHistory = append(m.menuHistory, m.state)
							m.state = stateOutput
							return m, nil
						}

						// Save the current state to history for back navigation
						m.menuHistory = append(m.menuHistory, m.state)

						// Create and set the class selection menu
						m.list = createClassSelectionMenu(classes)
						m.classList = classes
						m.currentMenu = "Class Selection"
						m.state = stateClassSelection
						return m, nil

					case "Create Class":
						// Save current state to history
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateClassInput
						m.currentMenu = "Create Class"
						return m, nil
					}
				}
			} else if m.state == stateClassSelection {
				// Handle class selection using the selectedItem index
				if m.selectedItem == len(m.classList) {
					// Back option selected - return to main menu
					// Pop the last state from history
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						m.state = m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]

						// Reset to main menu
						m.selectedItem = 0
						m.currentMenu = "Main Menu"
					} else {
						m.state = stateMainMenu
						m.selectedItem = 0
						m.currentMenu = "Main Menu"
					}
					return m, nil
				} else if m.selectedItem < len(m.classList) {
					// Save the selected class name
					m.selectedClass = m.classList[m.selectedItem]

					// Save current state to history
					m.menuHistory = append(m.menuHistory, m.state)

					// Create and set the class management menu
					m.list = createClassManagementMenu(m.selectedClass)
					m.currentMenu = "Class Management: " + m.selectedClass
					m.state = stateClassManagement
					m.selectedItem = 0 // Reset selection for the new menu
					return m, nil
				}
			} else if m.state == stateClassManagement {
				// Handle class management menu selection using selectedItem index
				// Define the options in the same order as they appear in the view
				managementOptions := []string{
					"Manage Students",
					"Manage Repos",
					"View GH Activity",
					"Delete Class",
					"Back",
				}

				if m.selectedItem < len(managementOptions) {
					selectedOption := managementOptions[m.selectedItem]
					switch selectedOption {
					case "Back":
						// Pop the last state from history
						if len(m.menuHistory) > 0 {
							lastIndex := len(m.menuHistory) - 1
							m.state = m.menuHistory[lastIndex]
							m.menuHistory = m.menuHistory[:lastIndex]

							// Reset to main menu
							items := []list.Item{
								item{title: "Manage Classes", description: "Select and manage an existing class"},
								item{title: "Create Class", description: "Create a new class"},
								item{title: "Quit", description: "Exit the application"},
							}
							m.list = list.New(items, list.NewDefaultDelegate(), 40, 12)
							m.list.Title = "Student Code Viewer"
							m.list.SetShowStatusBar(false)
							m.list.SetFilteringEnabled(false)
							m.list.Styles.Title = titleStyle
							m.list.Styles.PaginationStyle = paginationStyle
							m.list.Styles.HelpStyle = helpStyle
							m.currentMenu = "Main Menu"
						}
						return m, nil

					// Implement other class management options here
					// For now, we'll just show a placeholder message
					default:
						m.output = fmt.Sprintf("Selected option: %s for class %s\nThis feature will be implemented in Phase 2.", selectedOption, m.selectedClass)
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateOutput
						return m, nil
					}
				}
			} else if m.state == stateClassInput {
				m.className = m.classInput.Value()

				// Check if this is from the main menu's Create Class option
				if m.currentMenu == "Create Class" {
					_, err := db.Exec("INSERT INTO classes (name) VALUES (?)", m.className)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.output = fmt.Sprintf("Added class: %s\n\nPress Enter to return to main menu.", m.className)

					// Set up to return to main menu after showing output
					// Don't save class input state in history, so we go directly to main menu
					m.state = stateOutput
					// Clear the menu history to ensure we go back to main menu
					m.menuHistory = []int{stateMainMenu}
					return m, nil
				}

				// Handle the legacy code paths
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

				// NEW: Use tview for Week History.
				case "Week History":
					// Launch the tview-based week history view.
					if err := showWeekHistoryTview(m.className); err != nil {
						m.err = err
					} else {
						m.output = "Returned from Week History view."
					}
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
	case stateMainMenu, stateClassSelection, stateClassManagement:
		// Already handled above, but including here for completeness
		m.list, cmd = m.list.Update(msg)
	case stateClassInput:
		m.classInput, cmd = m.classInput.Update(msg)
	case stateStudentInput:
		m.studentInput, cmd = m.studentInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	// Handle any error
	if m.err != nil {
		errorMsg := errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		m.err = nil // Clear the error after displaying it
		return docStyle.Render(errorMsg + "\n\nPress Enter/Esc to continue.")
	}

	// Create breadcrumb navigation
	breadcrumb := breadcrumbStyle.Render(m.currentMenu)

	switch m.state {
	case stateMainMenu:
		// Use our new simple menu display with the selected item highlighted
		return docStyle.Render(createSimpleMenuWithSelection("Student Code Viewer", m.menuItems, m.selectedItem))

	case stateClassSelection:
		// Convert class list to menu items and add a Back option
		classItems := make([]item, len(m.classList)+1) // +1 for Back option
		for i, class := range m.classList {
			classItems[i] = item{title: class, description: "Select to manage this class"}
		}
		// Add Back option as the last item
		classItems[len(m.classList)] = item{title: "Back", description: "Return to main menu"}
		return docStyle.Render(createSimpleMenuWithSelection("Select a Class", classItems, m.selectedItem))

	case stateClassManagement:
		// Create class management menu items
		manageItems := []item{
			{title: "Manage Students", description: "Add or remove students"},
			{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
			{title: "View GH Activity", description: "Check student GitHub activity"},
			{title: "Delete Class", description: "Delete this class and its data"},
			{title: "Back", description: "Return to main menu"},
		}
		return docStyle.Render(createSimpleMenuWithSelection("Managing Class: "+m.selectedClass, manageItems, m.selectedItem))

	case stateClassInput:
		return docStyle.Render(
			breadcrumb + "\n" +
				titleStyle.Render("Enter Class Name") + "\n\n" +
				m.classInput.View() + "\n\n" +
				helpStyle.Render("Press Enter to confirm or Esc to cancel"),
		)

	case stateStudentInput:
		return docStyle.Render(
			breadcrumb + "\n" +
				titleStyle.Render("Enter Student Usernames") + "\n" +
				"(Space-separated list of GitHub usernames)\n\n" +
				m.studentInput.View(),
		)

	case stateOutput:
		return docStyle.Render(
			breadcrumbStyle.Render(m.currentMenu) + "\n" +
				outputBoxStyle.Render(m.output),
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
