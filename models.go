package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

// Item represents a menu item
type Item struct {
	title       string
	description string
}

// GithubEvent represents a GitHub event from the API
type GithubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
}

// Model represents the state of the application.
type Model struct {
	list  list.Model // For complex list views (if still used)
	state int        // Current application state

	// Input fields
	classInput   textinput.Model // For new class name input
	studentInput textinput.Model // For new student name input
	className    string          // Stores name of class being created or context

	output        string   // Holds command output to be rendered in stateOutput
	menuHistory   []int    // Stack to track menu navigation history for back functionality
	classList     []string // List of available classes for selection
	selectedClass string   // Class context for operations like add/delete student, manage repos

	// Fields for simple, list-based menus (primary UI component)
	menuItems    []Item
	selectedItem int    // Index of the currently selected item in simple menus
	currentMenu  string // Title of the current menu for breadcrumb navigation

	// Data for specific states/operations
	studentList []string // List of students for selection (e.g., for deletion)

	// Loading state additions
	spinner        spinner.Model // For loading animations
	loadingMessage string        // Message to display during loading

	err error // For general error messages in the model
}

// Required methods for list.Item interface
func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

// Model states
const (
	stateMainMenu = iota
	stateClassInput
	stateStudentInput
	stateOutput
	stateClassSelection
	stateClassManagement
	stateManageStudents
	stateManageRepos               // Placeholder for future state
	stateViewGHActivity            // New state for View GitHub Activity submenu
	stateStudentSelectionForDelete // New state for selecting a student to delete
	stateDeleteConfirmation
	stateLoading // New state for loading indicator
)

// operationResultMsg is sent when a background operation (like cloning) completes.
// It carries the logs or an error if one occurred.
type operationResultMsg struct {
	logs []string
	err  error
}
