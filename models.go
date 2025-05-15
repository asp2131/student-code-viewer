package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
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

// Model represents the application state
type Model struct {
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
	menuItems    []Item
	selectedItem int    // index of the currently selected item
	studentList  []string // List of students for selection (e.g., for deletion)
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
	stateManageRepos      // Placeholder for future state
	stateViewGHActivity   // Placeholder for future state
	stateStudentSelectionForDelete // New state for selecting a student to delete
	stateDeleteConfirmation
)
