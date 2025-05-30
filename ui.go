package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// UI Styles
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
		Foreground(lipgloss.Color("#888888")).
		MarginLeft(2).
		MarginBottom(1).
		Padding(0, 1)

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

	inputStyleFocused = lipgloss.NewStyle().BorderForeground(lipgloss.Color("205")).Padding(0, 1).Border(lipgloss.RoundedBorder()).Width(40)
	inputStyleBlurred = lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")).Padding(0, 1).Border(lipgloss.RoundedBorder()).Width(40)
	secondaryTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Icons for status indicators
const (
	iconSuccess = "✔" // single-width check
	iconWarning = "!" // single-width exclamation
	iconError   = "✖" // single-width cross
)

// createSimpleMenu creates a string-based menu display that shows all options at once
func createSimpleMenu(title string, options []Item) string {
	return createSimpleMenuWithSelection(title, options, 0)
}

// createSimpleMenuWithSelection creates a menu with a specific item selected
func createSimpleMenuWithSelection(title string, options []Item, selectedIndex int) string {
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

// buildBreadcrumb creates a formatted breadcrumb navigation string from a path
func buildBreadcrumb(path []string) string {
	if len(path) == 0 {
		return ""
	}
	
	// Join path elements with " > " separator
	breadcrumbText := strings.Join(path, " > ")
	
	// Add navigation icon at the beginning
	breadcrumbText = "📍 " + breadcrumbText
	
	return breadcrumbStyle.Render(breadcrumbText)
}

// createClassSelectionMenu creates a menu for selecting classes
func createClassSelectionMenu(classes []string) list.Model {
	items := make([]list.Item, len(classes))
	for i, class := range classes {
		items[i] = Item{title: class, description: "Select to manage this class"}
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

// createClassManagementMenu creates a menu for managing a selected class
func createClassManagementMenu(className string) list.Model {
	items := []list.Item{
		Item{title: "Manage Students", description: "Add or remove students"},
		Item{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
		Item{title: "View GH Activity", description: "Check student GitHub activity"},
		Item{title: "Delete Class", description: "Delete this class and its data"},
		Item{title: "Back", description: "Return to main menu"},
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

// createViewGHActivityMenu creates a menu for viewing GitHub activity for a selected class
func createViewGHActivityMenu(className string) list.Model {
	items := []list.Item{
		Item{title: "Week View", description: "View student activity for the past week"},
		Item{title: "Check Latest Activity", description: "Display the latest commit time for each student."},
		Item{title: "Back", description: "Return to class management menu"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 40, 12)
	l.Title = "GH Activity for: " + className
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

// getBreadcrumbPath returns the breadcrumb path based on the current menu state
func getBreadcrumbPath(m Model, menuName string) []string {
	switch m.state {
	case stateMainMenu:
		return []string{"Home"}
	case stateClassSelection:
		return []string{"Home", "Select Class"}
	case stateClassManagement:
		return []string{"Home", "Select Class", m.selectedClass}
	case stateManageStudents:
		return []string{"Home", "Select Class", m.selectedClass, "Manage Students"}
	case stateManageRepos:
		return []string{"Home", "Select Class", m.selectedClass, "Manage Repos"}
	case stateViewGHActivity:
		return []string{"Home", "Select Class", m.selectedClass, "GitHub Activity"}
	case stateClassInput:
		if strings.Contains(menuName, "Create Class") {
			return []string{"Home", "Create Class"}
		} else if strings.Contains(menuName, "Add Students") {
			return []string{"Home", "Select Class", m.selectedClass, "Manage Students", "Add Students"}
		}
		return []string{"Home", "Create Class"}
	case stateStudentInput:
		return []string{"Home", "Select Class", m.selectedClass, "Manage Students", "Add Students"}
	case stateStudentSelectionForDelete:
		return []string{"Home", "Select Class", m.selectedClass, "Manage Students", "Delete Student"}
	case stateDeleteConfirmation:
		return []string{"Home", "Select Class", m.selectedClass, "Delete Class"}
	case stateOutput:
		// Keep the previous breadcrumb path for output states
		if len(m.breadcrumbPath) > 0 {
			return m.breadcrumbPath
		}
		return []string{"Home", "Output"}
	default:
		// Fallback to current menu name
		if menuName != "" {
			return []string{"Home", menuName}
		}
		return []string{"Home"}
	}
}
