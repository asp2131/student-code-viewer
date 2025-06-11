package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

// TerminalSizeInfo provides enhanced terminal size information and categorization
type TerminalSizeInfo struct {
	Width, Height int
	IsSmall       bool // < 80 cols
	IsTiny        bool // < 60 cols
	IsNarrow      bool // < 100 cols
	IsTall        bool // > 30 rows
	IsVeryTall    bool // > 50 rows
}

// GetLayoutMode returns the appropriate layout mode based on terminal size
func (t TerminalSizeInfo) GetLayoutMode() string {
	if t.IsTiny {
		return "compact"
	}
	if t.IsSmall {
		return "minimal"
	}
	if t.IsNarrow {
		return "standard"
	}
	return "wide"
}

// GetMaxMenuItems returns the recommended maximum menu items for the current size
func (t TerminalSizeInfo) GetMaxMenuItems() int {
	switch t.GetLayoutMode() {
	case "compact":
		return 5
	case "minimal":
		return 8
	case "standard":
		return 12
	default:
		return 20
	}
}

// ShouldShowDescriptions returns whether menu item descriptions should be shown
func (t TerminalSizeInfo) ShouldShowDescriptions() bool {
	return !t.IsSmall
}

// ShouldShowBreadcrumbs returns whether breadcrumb navigation should be shown
func (t TerminalSizeInfo) ShouldShowBreadcrumbs() bool {
	return !t.IsTiny
}

// NewTerminalSizeInfo creates a TerminalSizeInfo from width and height
func NewTerminalSizeInfo(width, height int) TerminalSizeInfo {
	return TerminalSizeInfo{
		Width:      width,
		Height:     height,
		IsSmall:    width < 80,
		IsTiny:     width < 60,
		IsNarrow:   width < 100,
		IsTall:     height > 30,
		IsVeryTall: height > 50,
	}
}

// LayoutConfig defines how the UI should be laid out based on terminal size
type LayoutConfig struct {
	ShowBreadcrumbs  bool
	ShowDescriptions bool
	TableStyle       string // "minimal", "compact", "standard", "full"
	MaxMenuItems     int
	ListWidth        int
	ListHeight       int
	InputWidth       int
	TableWidth       int
	UseCompactTables bool
	UseVerticalStack bool // For very narrow terminals
}

// GetLayoutConfig returns the appropriate layout configuration for the terminal size
func GetLayoutConfig(termSize TerminalSizeInfo) LayoutConfig {
	config := LayoutConfig{
		ShowBreadcrumbs:  termSize.ShouldShowBreadcrumbs(),
		ShowDescriptions: termSize.ShouldShowDescriptions(),
		MaxMenuItems:     termSize.GetMaxMenuItems(),
	}

	switch termSize.GetLayoutMode() {
	case "compact":
		config.TableStyle = "minimal"
		config.ListWidth = max(30, termSize.Width-8)
		config.ListHeight = max(8, termSize.Height-6)
		config.InputWidth = max(20, termSize.Width-10)
		config.TableWidth = max(40, termSize.Width-8)
		config.UseCompactTables = true
		config.UseVerticalStack = true
		
	case "minimal":
		config.TableStyle = "compact"
		config.ListWidth = max(40, termSize.Width-10)
		config.ListHeight = max(10, termSize.Height-8)
		config.InputWidth = max(30, termSize.Width-15)
		config.TableWidth = max(50, termSize.Width-10)
		config.UseCompactTables = true
		config.UseVerticalStack = false
		
	case "standard":
		config.TableStyle = "standard"
		config.ListWidth = max(50, min(termSize.Width-10, 80))
		config.ListHeight = max(12, min(termSize.Height-8, 18))
		config.InputWidth = max(40, min(termSize.Width-20, 60))
		config.TableWidth = max(60, min(termSize.Width-10, 100))
		config.UseCompactTables = false
		config.UseVerticalStack = false
		
	default: // "wide"
		config.TableStyle = "full"
		config.ListWidth = max(60, min(termSize.Width-15, 100))
		config.ListHeight = max(15, min(termSize.Height-10, 25))
		config.InputWidth = max(50, min(termSize.Width-30, 80))
		config.TableWidth = max(80, min(termSize.Width-15, 120))
		config.UseCompactTables = false
		config.UseVerticalStack = false
	}

	return config
}

// Item represents a menu item
type Item struct {
	title       string
	description string
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
	breadcrumbPath []string // Full navigation path for enhanced breadcrumb display

	// Data for specific states/operations
	studentList []string // List of students for selection (e.g., for deletion)

	// Loading state additions
	spinner        spinner.Model // For loading animations
	loadingMessage string        // Message to display during loading

	// Terminal size for responsive design
	terminalSize TerminalSizeInfo // Enhanced terminal size information

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
